package request

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/google/uuid"
)

// Options defines options of an HTTP request
type Options struct {
	Context              context.Context
	Method               string
	URL                  *url.URL
	Proxy                *url.URL
	Headers              map[string]string
	Cookies              []*http.Cookie
	Parameters           map[string]string
	Accept               string
	PayloadType          string // if not provided, it is computed. payload==struct => json, payload==map => form
	Payload              interface{}
	AttachmentType       string
	Attachment           io.Reader // binary data that should be attached to the paylod (e.g.: multipart forms)
	Authorization        string
	RequestID            string
	UserAgent            string
	Transport            *http.Transport
	RetryableStatusCodes []int
	Attempts             int
	InterAttemptDelay    time.Duration
	Timeout              time.Duration
	RequestBodyLogSize   int // how many characters of the request body should be logged, if possible (<0 => nothing logged)
	ResponseBodyLogSize  int // how many characters of the response body should be logged (<0 => nothing logged)
	Logger               *logger.Logger
}

// DefaultAttempts defines the number of attempts for requests by default
const DefaultAttempts = 5

// DefaultTimeout defines the timeout for a request
const DefaultTimeout = 2 * time.Second

// DefaultInterAttemptDelay defines the sleep delay between 2 attempts
const DefaultInterAttemptDelay = 1 * time.Second

// DefaultRequestBodyLogSize  defines the maximum size of the request body that should be logged
const DefaultRequestBodyLogSize = 2048

// DefaultResponseBodyLogSize  defines the maximum size of the response body that should be logged
const DefaultResponseBodyLogSize = 2048

// Send sends an HTTP request
func Send(options *Options, results interface{}) (*Content, error) {
	var err error

	if err = normalizeOptions(options, results); err != nil {
		return nil, err
	}
	log := options.Logger.Child(nil, "request", "reqid", options.RequestID, "method", options.Method)

	log.Debugf("HTTP %s %s", options.Method, options.URL.String())
	req, err := buildRequest(log, options)
	if err != nil {
		return nil, err // err is already decorated
	}
	log = log.Record("method", options.Method)

	httpclient := http.Client{
		Transport: options.Transport,
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			log.Tracef("Following WEB Link: %s", r.URL)
			for _, v := range via {
				log.Tracef("Via: %s", v.URL)
			}
			return nil
		},
		Timeout: options.Timeout,
	}
	// Sending the request...
	for attempt := 0; attempt < options.Attempts; attempt++ {
		log.Tracef("Attempt #%d/%d", attempt+1, options.Attempts)
		req.Header.Set("X-Attempt", strconv.Itoa(attempt+1))
		log.Tracef("Request Headers: %#v", req.Header)
		start := time.Now()
		res, err := httpclient.Do(req)
		duration := time.Since(start)
		log = log.Record("duration", duration/time.Millisecond)
		if err != nil {
			urlErr := &url.Error{}
			if errors.As(err, &urlErr) {
				if urlErr.Timeout() || urlErr.Temporary() || urlErr.Unwrap() == io.EOF || errors.Is(err, context.DeadlineExceeded) {
					if attempt+1 < options.Attempts {
						log.Warnf("Temporary failed to send request (duration: %s/%s), Error: %s", duration, options.Timeout, err.Error()) // we don't want the stack here
						log.Infof("Waiting for %s before trying again", options.InterAttemptDelay)
						time.Sleep(options.InterAttemptDelay)
						req, _ = buildRequest(log, options)
						continue
					}
					break
				} else {
					log.Errorf("URL Error, temporary=%t, timeout=%t, unwrap=%s", urlErr.Temporary(), urlErr.Timeout(), urlErr.Unwrap(), err)
					return nil, errors.WithStack(err)
				}
			}
			return nil, err
		}
		defer res.Body.Close()
		log.Debugf("Response %s in %s", res.Status, duration)
		log.Tracef("Response Headers: %#v", res.Header)

		// Processing the status
		if res.StatusCode >= 400 {
			if isRetryable(res.StatusCode, options.RetryableStatusCodes) {
				if attempt+1 < options.Attempts {
					log.Warnf("Retryable Response Status: %s", res.Status)
					log.Infof("Waiting for %s before trying again", options.InterAttemptDelay)
					time.Sleep(options.InterAttemptDelay)
					req, _ = buildRequest(log, options)
					continue
				}
			}
			// Read the body to get the error message
			resContent, err := ContentFromReader(res.Body, res.Header.Get("Content-Type"), core.Atoi(res.Header.Get("Content-Length"), 0), res.Header, res.Cookies(), log)
			if err != nil {
				return nil, errors.FromHTTPStatusCode(res.StatusCode)
			}
			log.Tracef("Response body: %s", resContent.LogString(uint64(options.ResponseBodyLogSize)))
			return resContent, errors.FromHTTPStatusCode(res.StatusCode)
		}

		// Analyze the response content type
		resContentType := res.Header.Get("Content-Type")

		// some servers give the wrong mime type for JPEG files
		if resContentType == "image/jpg" {
			resContentType = "image/jpeg"
		}
		if len(resContentType) == 0 || resContentType == "application/octet-stream" {
			if len(options.Accept) > 0 && options.Accept != "*" {
				// TODO: well... Accept is not always a simple mime type...
				resContentType = options.Accept
			} else {
				if mimetype := mime.TypeByExtension(filepath.Ext(options.URL.Path)); len(mimetype) > 0 {
					resContentType = mimetype
				}
			}
		}
		log.Tracef("Computed Response Content-Type: %s", resContentType)

		// Reading the response body

		if writer, ok := results.(io.Writer); ok {
			bytesRead, err := io.Copy(writer, res.Body)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			log.Tracef("Read %d bytes", bytesRead)
			resContent := ContentWithData([]byte{}, resContentType, bytesRead, res.Header, res.Cookies())
			return resContent, nil
		} else if results != nil { // Unmarshaling the response body if requested (structs, arrays, maps, etc)
			resContent, err := ContentFromReader(res.Body, resContentType, res.Header, res.Cookies(), log)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			log.Tracef("Response body: %s", resContent.LogString(uint64(options.ResponseBodyLogSize)))
			err = json.Unmarshal(resContent.Data, results)
			if err != nil {
				log.Debugf("Failed to unmarshal response body, use the Content, JSON Error: %s", err)
			}
			return resContent, nil
		}

		// Reading all the response body into the Content
		resContent, err := ContentFromReader(res.Body, resContentType, core.Atoi(res.Header.Get("Content-Length"), 0), res.Header, res.Cookies(), log)
		if err != nil {
			log.Errorf("Failed to read response body: %v%s", err, "") // the extra string arg is to prevent the logger to dump the stack trace
			return nil, err                                           // err is already "decorated" by ContentReader
		}
		log.Tracef("Response body: %s", resContent.LogString(uint64(options.ResponseBodyLogSize)))

		return resContent, nil
	}
	// If we get here, there is an error
	return nil, errors.Wrapf(errors.HTTPStatusRequestTimeout, "Giving up after %d attempts", options.Attempts)
}

func normalizeOptions(options *Options, results interface{}) (err error) {
	if options == nil {
		return errors.ArgumentMissing.With("options")
	}
	if options.URL == nil {
		return errors.ArgumentMissing.With("URL")
	}
	if options.Context == nil {
		options.Context = context.Background()
	}
	if options.Logger == nil {
		options.Logger, err = logger.FromContext(options.Context)
		if err != nil {
			options.Logger = logger.Create("request") // without a logger, let's log into the "void"
		}
	}
	if options.RequestBodyLogSize == 0 {
		options.RequestBodyLogSize = DefaultRequestBodyLogSize
	} else if options.RequestBodyLogSize < 0 {
		options.RequestBodyLogSize = 0
	}
	if options.ResponseBodyLogSize == 0 {
		options.ResponseBodyLogSize = DefaultResponseBodyLogSize
	} else if options.ResponseBodyLogSize < 0 {
		options.ResponseBodyLogSize = 0
	}
	if len(options.RequestID) == 0 {
		options.RequestID = uuid.Must(uuid.NewRandom()).String()
	}
	if len(options.UserAgent) == 0 {
		options.UserAgent = "Request " + VERSION
	}
	if len(options.Accept) == 0 {
		if _, ok := results.(io.Writer); !ok && results != nil {
			options.Accept = "application/json"
		} else {
			options.Accept = "*"
		}
	}
	if options.Timeout == 0 {
		options.Timeout = time.Duration(DefaultTimeout)
	}
	if options.Attempts < 1 {
		options.Attempts = DefaultAttempts
	}
	if options.InterAttemptDelay < 1*time.Second {
		options.InterAttemptDelay = time.Duration(DefaultInterAttemptDelay)
	}
	if options.Parameters != nil {
		query := options.URL.Query()
		for key, value := range options.Parameters {
			query.Add(key, value)
		}
		options.URL.RawQuery = query.Encode()
	}
	if options.Transport == nil  {
		options.Transport = http.DefaultTransport.(*http.Transport).Clone()
	}
	if options.Proxy != nil {
		options.Transport.Proxy = http.ProxyURL(options.Proxy)
	}
	if options.Attempts > 1 {
		if options.Payload != nil {
			if _, ok := options.Payload.(io.Reader); ok {
				if _, ok := options.Payload.(io.Seeker); !ok {
					return errors.New("Payload must be an io.Seeker").(errors.Error).Wrap(errors.ArgumentInvalid.With("payload"))
				}
			}
		}
		if options.Attachment != nil {
			if _, ok := options.Attachment.(io.Seeker); !ok {
				return errors.New("Attachment must be an io.Seeker").(errors.Error).Wrap(errors.ArgumentInvalid.With("attachment"))
			}
		}
	}

	return nil
}

// buildRequestContent builds a Content for the request
func buildRequestContent(log *logger.Logger, options *Options) (content *Content, err error) {
	// Analyze payload
	if options.Payload == nil {
		if options.Attachment == nil {
			return &Content{}, nil
		}
		// We have an attachment, so the user meant it to be the payload
		options.Payload = options.Attachment
		if len(options.AttachmentType) > 0 {
			options.PayloadType = options.AttachmentType
		}
	}

	if _content, ok := options.Payload.(Content); ok {
		log.Tracef("Payload is a Content (Type: %s, size: %d)", _content.Type, _content.Length)
		if len(_content.Type) == 0 {
			if len(options.PayloadType) > 0 {
				_content.Type = options.PayloadType
			} else {
				_content.Type = "application/octet-stream"
			}
		}
		content = &_content
	} else if _content, ok := options.Payload.(*Content); ok {
		log.Tracef("Payload is a *Content (Type: %s, size: %d)", _content.Type, _content.Length)
		if len(_content.Type) == 0 {
			if len(options.PayloadType) > 0 {
				_content.Type = options.PayloadType
			} else {
				_content.Type = "application/octet-stream"
			}
		}
		content = _content
	} else if reader, ok := options.Payload.(io.Reader); ok {
		log.Tracef("Payload is a Reader (Data Type: %s)", options.PayloadType)
		content, _ = ContentFromReader(reader, options.PayloadType, 0, nil, nil)
	} else {
		payloadType := reflect.TypeOf(options.Payload)
		if payloadType.Kind() == reflect.Struct || (payloadType.Kind() == reflect.Ptr && reflect.Indirect(reflect.ValueOf(options.Payload)).Kind() == reflect.Struct) { // JSONify the payload
			log.Tracef("Payload is a Struct, JSONifying it")
			// TODO: Add other payload types like XML, etc
			if len(options.PayloadType) == 0 {
				options.PayloadType = "application/json"
			}
			payload, err := json.Marshal(options.Payload)
			if err != nil {
				if errors.Is(err, errors.JSONMarshalError) {
					return nil, err
				}
				return nil, errors.JSONMarshalError.Wrap(err)
			}
			content = ContentWithData(payload, options.PayloadType)
		} else if payloadType.Kind() == reflect.Array || payloadType.Kind() == reflect.Slice {
			log.Tracef("Payload is an array or a slice, JSONifying it")
			// TODO: Add other payload types like XML, etc
			if len(options.PayloadType) == 0 {
				options.PayloadType = "application/json"
			}
			payload, err := json.Marshal(options.Payload)
			if err != nil {
				if errors.Is(err, errors.JSONMarshalError) {
					return nil, err
				}
				return nil, errors.JSONMarshalError.Wrap(err)
			}
			content = ContentWithData(payload, options.PayloadType)
		} else if payloadType.Kind() == reflect.Map {
			// Collect the attributes from the map
			attributes := map[string]string{}
			if stringMap, ok := options.Payload.(map[string]string); ok {
				log.Tracef("Payload is a StringMap")
				attributes = stringMap
			} else { // traverse the map, collecting values if they are Stringer. Note: This can be slow...
				log.Tracef("Payload is a Map")
				items := reflect.ValueOf(options.Payload)
				for _, item := range items.MapKeys() {
					value := items.MapIndex(item)
					if stringer, ok := value.Interface().(fmt.Stringer); ok {
						attributes[item.String()] = stringer.String()
					}
				}
			}

			// Build the content as a Form or a Multipart Data Form
			if options.Attachment == nil {
				log.Tracef("Building a form (no attachment)")
				if len(options.PayloadType) == 0 {
					options.PayloadType = "application/x-www-form-urlencoded"
				}
				form := url.Values{}
				for key, value := range attributes {
					form.Set(key, value)
				}
				return ContentWithData([]byte(form.Encode()), options.PayloadType), nil
			}

			log.Tracef("Building a multipart data form with 1 attachment")
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			for key, value := range attributes {
				if strings.HasPrefix(key, ">") {
					key = strings.TrimPrefix(key, ">")
					if len(key) == 0 {
						return nil, errors.Errorf("Empty key for multipart form field with attachment")
					}
					if len(value) == 0 {
						return nil, errors.Errorf("Empty value for multipart form field %s", key)
					}
					partHeader := textproto.MIMEHeader{}
					partHeader.Add("Content-Disposition", fmt.Sprintf("form-data; name=\"%s\"; filename=\"%s\"", key, value))
					if len(options.AttachmentType) > 0 {
						partHeader.Add("Content-Type", options.AttachmentType)
					}
					part, err := writer.CreatePart(partHeader)
					if err != nil {
						return nil, errors.Wrapf(err, "Failed to create multipart for field %s", key)
					}
					_, err = options.Attachment.(io.Seeker).Seek(0, io.SeekStart)
					if err != nil {
						return nil, errors.Wrapf(err, "Failed to seek to beginning of attachment for field %s", key)
					}
					written, err := io.Copy(part, options.Attachment)
					if err != nil {
						return nil, errors.Errorf("Failed to write attachment to multipart form field %s", key)
					}
					if written == 0 {
						return nil, errors.Errorf("Missing/Empty Attachment for multipart form field %s", key)
					}
					log.Tracef("Wrote %d bytes to multipart form field %s", written, key)
				} else {
					if err := writer.WriteField(key, value); err != nil {
						return nil, errors.Wrapf(err, "Failed to create multipart form field %s", key)
					}
					log.Tracef("  Added field %s = %s", key, value)
				}
			}
			if err := writer.Close(); err != nil {
				return nil, errors.Wrap(err, "Failed to create multipart data")
			}
			content, _ = ContentFromReader(body, writer.FormDataContentType())
		}
	}
	if content != nil {
		if options.RequestBodyLogSize > 0 {
			log.Tracef("Request body %d bytes: \n%s", content.Length, string(content.Data[:int(math.Min(float64(options.RequestBodyLogSize), float64(content.Length)))]))
		} else {
			log.Tracef("Request body %d bytes", content.Length)
		}
		return content, nil
	}
	return nil, errors.ArgumentInvalid.With("payload")
}

func buildRequest(log *logger.Logger, options *Options) (*http.Request, error) {
	reqContent, err := buildRequestContent(log, options)
	if err != nil {
		return nil, err // err is already decorated
	}
	if len(options.Method) == 0 {
		if reqContent.Length > 0 {
			options.Method = "POST"
		} else {
			options.Method = "GET"
		}
		log.Tracef("Computed HTTP method: %s", options.Method)
	}

	req, err := http.NewRequestWithContext(options.Context, options.Method, options.URL.String(), reqContent.Reader())
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Close indicates to close the connection or after sending this request and reading its response.
	// setting this field prevents re-use of TCP connections between requests to the same hosts, as if Transport.DisableKeepAlives were set.
	req.Close = true

	// Setting request headers
	req.Header.Set("User-Agent", options.UserAgent)
	req.Header.Set("Accept", options.Accept)
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Add("Accept-Encoding", "deflate")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("X-Request-Id", options.RequestID)
	if len(options.Authorization) > 0 {
		req.Header.Set("Authorization", options.Authorization)
	}
	if len(reqContent.Type) > 0 {
		req.Header.Set("Content-Type", reqContent.Type)
	}
	if reqContent.Length > 0 {
		req.Header.Set("Content-Length", strconv.FormatUint(reqContent.Length, 10))
	}
	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}

	if len(options.Cookies) > 0 {
		for _, cookie := range options.Cookies {
			req.AddCookie(cookie)
		}
	}
	return req, nil
}

func isRetryable(statusCode int, retryableStatusCodes []int) bool {
	for _, retryable := range retryableStatusCodes {
		if statusCode == retryable {
			return true
		}
	}
	return false
}
