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
	Context             context.Context
	Method              string
	URL                 *url.URL
	Proxy               *url.URL
	Headers             map[string]string
	Parameters          map[string]string
	Accept              string
	PayloadType         string         // if not provided, it is computed. payload==struct => json, payload==map => form
	Payload             interface{}
	Attachment          *ContentReader // binary data that should be attached to the paylod (e.g.: multipart forms)
	Authorization       string
	RequestID           string
	UserAgent           string
	Attempts            int
	InterAttemptDelay   time.Duration
	Timeout             time.Duration
	RequestBodyLogSize  int            // how many characters of the request body should be logged, if possible (<0 => nothing logged)
	ResponseBodyLogSize int            // how many characters of the response body should be logged (<0 => nothing logged)
	Logger              *logger.Logger
}

// Error is returned when an HTTP Status is >= 400
type Error struct {
	StatusCode int    `json:"statusCode"`
	Status     string `json:"status"`
}

// DefaultAttempts defines the number of attempts for requests by default
const DefaultAttempts = 5

// DefaultTimeout defunes the timeout for a request
const DefaultTimeout  = 2 * time.Second

// DefaultInterAttemptDelay defines the sleep delay between 2 attempts
const DefaultInterAttemptDelay = 1 * time.Second

// DefaultRequestBodyLogSize  defines the maximum size of the request body that should be logge
const DefaultRequestBodyLogSize = 2048

// DefaultResponseBodyLogSize  defines the maximum size of the response body that should be logge
const DefaultResponseBodyLogSize = 2048

func (err Error) Error() string {
	return err.Status
}

// SendRequest sends an HTTP request
func Send(options *Options, results interface{}) (*ContentReader, error) {
	var err error

	if options.Context == nil {
		options.Context = context.Background()
	}
	if options.URL == nil {
		return nil, errors.ArgumentMissingError.WithWhat("URL")
	}
	if len(options.RequestID) == 0 {
		options.RequestID = uuid.Must(uuid.NewRandom()).String()
	}

	// without a logger, let's log into the "void"
	log := logger.Create("request", options.Logger).Child("", "request", "reqid", options.RequestID)

	if options.RequestBodyLogSize == 0 {
		options.RequestBodyLogSize = DefaultRequestBodyLogSize
	} else if options.RequestBodyLogSize < 0 {
		options.RequestBodyLogSize = 0
	}

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
	}
	log = log.Record("method", options.Method)

	if len(options.Accept) == 0 {
		if results != nil {
			options.Accept = "application/json"
		} else {
			options.Accept = "*"
		}
	}

	if options.Attempts < 1 {
		options.Attempts = DefaultAttempts
	}

	if options.InterAttemptDelay < 1 * time.Second {
		options.InterAttemptDelay = time.Duration(DefaultInterAttemptDelay)
	}

	if options.ResponseBodyLogSize == 0 {
		options.ResponseBodyLogSize = DefaultResponseBodyLogSize
	} else if options.ResponseBodyLogSize < 0 {
		options.ResponseBodyLogSize = 0
	}

	if options.Timeout == 0 {
		options.Timeout = time.Duration(DefaultTimeout)
	}

	if options.Parameters != nil {
		log.Tracef("Adding query parameters")
		query := options.URL.Query()
		for key, value := range options.Parameters {
			query.Add(key, value)
		}
		options.URL.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(options.Context, options.Method, options.URL.String(), reqContent.Reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Setting request headers
	req.Header.Set("User-Agent",   options.UserAgent)
	req.Header.Set("Accept",       options.Accept)
	req.Header.Set("X-Request-Id", options.RequestID)
	if len(options.Authorization) > 0 {
		req.Header.Set("Authorization", options.Authorization)
	}
	if len(reqContent.Type) > 0 {
		req.Header.Set("Content-Type", reqContent.Type)
	}
	if reqContent.Length > 0 {
		req.Header.Set("Content-Length", strconv.FormatInt(reqContent.Length, 10))
	}
	for key, value := range options.Headers {
		req.Header.Set(key, value)
	}

	httpclient := http.Client{
		CheckRedirect: func(r *http.Request, via []*http.Request) error {
			r.URL.Opaque = r.URL.Path
			return nil
		},
		Timeout: options.Timeout,
	}
	if options.Proxy != nil {
		httpclient.Transport = &http.Transport{Proxy: http.ProxyURL(options.Proxy)}
	}
	// Sending the request...
	for attempt := 0; attempt < options.Attempts; attempt++ {
		log.Tracef("HTTP %s %s #%d/%d", req.Method, req.URL.String(), attempt, options.Attempts)
		log.Tracef("Request Headers: %#v", req.Header)
		start    := time.Now()
		res, err := httpclient.Do(req)
		duration := time.Since(start)
		log      = log.Record("duration", duration)
		if err != nil {
			log.Errorf("Failed to send request", err)
			if attempt + 1 < options.Attempts {
				log.Infof("Waiting for %s before trying again", options.InterAttemptDelay)
				time.Sleep(options.InterAttemptDelay)
			}
			continue
		}
		defer res.Body.Close()
		log.Tracef("Response %s in %s", res.Status, duration)
		log.Tracef("Response Headers: %#v", res.Header)

		// Reading the response body
		resContent, err := ContentFromReader(res.Body, res.Header.Get("Content-Type"), core.Atoi(res.Header.Get("Content-Length"), 0))
		if err != nil {
			log.Errorf("Failed to read response body", err)
			return nil, errors.WithStack(err)
		}
		// some servers give the wrong mime type for JPEG files
		if resContent.Type == "image/jpg" {
			resContent.Type = "image/jpeg"
		}
		if len(resContent.Type) == 0 || resContent.Type == "application/octet-stream" {
			if len(options.Accept) > 0 && options.Accept != "*" {
				// TODO: well... Accept is not always a simple mime type...
				resContent.Type = options.Accept
			}
			if resContent.Type == "application/octet-stream" {
				_ = mime.AddExtensionType(".mp3",  "audio/mpeg3")
				_ = mime.AddExtensionType(".m4a",  "audio/x-m4a")
				_ = mime.AddExtensionType(".wav",  "audio/wav")
				_ = mime.AddExtensionType(".jpeg", "image/jpg")
				if restype := mime.TypeByExtension(filepath.Ext(options.URL.Path)); len(restype) > 0 {
					resContent.Type = restype
				}
			}
		}
		if options.ResponseBodyLogSize > 0 {
			log.Tracef("Response body: %s, %d bytes: \n%s", resContent.Type, resContent.Length, string(resContent.Data[:int(math.Min(float64(options.ResponseBodyLogSize), float64(resContent.Length)))]))
		} else {
			log.Tracef("Response body: %s, %d bytes", resContent.Type, resContent.Length)
		}

		// Processing the status
		if res.StatusCode >= 400 {
			return resContent.Reader(), errors.FromHTTPStatusCode(res.StatusCode)
		}

		// Unmarshaling the response content if requested
		if results != nil {
			err = json.Unmarshal(resContent.Data, results)
			if err != nil {
				log.Errorf("Failed to decode response, use the ContentReader", err)
			}
		}
		return resContent.Reader(), nil
	}
	// If we get here, there is an error
	return nil, errors.Wrapf(errors.HTTPStatusRequestTimeoutError, "Giving up after %d attempts", options.Attempts)
}

// buildRequestContent builds a Content for the request
func buildRequestContent(log *logger.Logger, options *Options) (*ContentReader, error) {
	// Analyze payload
	if options.Payload == nil {
		if options.Attachment == nil {
			content := Content{}
			return content.Reader(), nil
		}
		// We have an attachment, so the user meant it to be the payload
		options.Payload = options.Attachment
	}

	payloadType := reflect.TypeOf(options.Payload)
	if contentReader, ok := options.Payload.(ContentReader); ok {
		// Here we ignore options.PayloadType as the ContentReader embeds its ContentType
		log.Tracef("Payload is a ContentReader (Type: %s, size: %d)", contentReader.Type, contentReader.Length)
		if len(contentReader.Type) == 0 {
			contentReader.Type = "application/octet-stream"
		}
		if contentReader.Length == 0 {
			// Let's try to get a length from the reader
			content, err := ContentFromReader(contentReader, contentReader.Type)
			if err != nil { return nil, err }
			return content.Reader(), nil
		}
		return &contentReader, nil
		/*
			} else if contentReader, ok := options.Payload.(*ContentReader); ok {
				log.Tracef("Payload is a *ContentReader (Type: %s, size: %d)", contentReader.Type, contentReader.Length)
				if contentReader.Length > 0 {
					content, _ = ContentFromReader(contentReader, contentReader.Type)
					if len(content.Type) == 0 {
						content.Type = "application/octet-stream"
					}
				} // else the returned contentReader will be empty
		*/
	} else if content, ok := options.Payload.(Content); ok {
		// Here we ignore options.PayloadType as the Content embeds its ContentType
		log.Tracef("Payload is a Content (Type: %s, size: %d)", content.Type, content.Length)
		return content.Reader(), nil
	} else if payloadType.Kind() == reflect.Struct || (payloadType.Kind() == reflect.Ptr && reflect.Indirect(reflect.ValueOf(options.Payload)).Kind() == reflect.Struct) { // JSONify the payload
		log.Tracef("Payload is a Struct, JSONifying it")
		// TODO: Add other payload types like XML, etc
		if len(options.PayloadType) == 0 {
			options.PayloadType = "application/json"
		}
		payload, err := json.Marshal(options.Payload)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to encode payload into JSON")
		}
		if options.RequestBodyLogSize > 0 {
			log.Tracef("Request body %d bytes: \n%s", len(payload), string(payload[:int(math.Min(float64(options.RequestBodyLogSize), float64(len(payload))))]))
		} else {
			log.Tracef("Request body %d bytes", len(payload))
		}
		return ContentWithData(payload, options.PayloadType).Reader(), nil
	} else if payloadType.Kind() == reflect.Array || payloadType.Kind() == reflect.Slice {
		log.Tracef("Payload is an array or a slice, JSONifying it")
		// TODO: Add other payload types like XML, etc
		if len(options.PayloadType) == 0 {
			options.PayloadType = "application/json"
		}
		payload, err := json.Marshal(options.Payload)
		if err != nil {
			return nil, errors.Wrap(err, "Failed to encode payload into JSON")
		}
		if options.RequestBodyLogSize > 0 {
			log.Tracef("Request body %d bytes: \n%s", len(payload), string(payload[:int(math.Min(float64(options.RequestBodyLogSize), float64(len(payload))))]))
		} else {
			log.Tracef("Request body %d bytes", len(payload))
		}
		return ContentWithData(payload, options.PayloadType).Reader(), nil
	} else if payloadType.Kind() == reflect.Map {
		// Collect the attributes from the map
		attributes := map[string]string{}
		if stringMap, ok := options.Payload.(map[string]string); ok {
			log.Tracef("Payload is a StringMap")
			attributes = stringMap
		} else if stringerMap, ok := options.Payload.(map[string]fmt.Stringer); ok {
			log.Tracef("Payload is a StringerMap")
			for key, value := range stringerMap {
				attributes[key] = value.String()
			}
		} else {
			keyType   := payloadType.Key()
			valueType := payloadType.Elem()
			return nil, errors.Errorf("Unsupported Payload map (map[%s]%s)", keyType.String(), valueType.String())
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
			return ContentWithData([]byte(form.Encode()), options.PayloadType).Reader(), nil
		}

		log.Tracef("Building a multipart data form with 1 attachment")
		body   := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		for key, value := range attributes {
			if strings.HasPrefix(key, ">") {
				key = strings.TrimPrefix(key, ">")
				if len(value) == 0 {
					return nil, errors.Errorf("Empty value for multipart form field %s", key)
				}
				part, err := writer.CreateFormFile(key, value)
				if err != nil {
					return nil, errors.Wrapf(err, "Failed to create multipart for field %s", key)
				}
				if options.Attachment.Length == 0 {
					return nil, errors.Errorf("Missing/Empty Attachment for multipart form field %s", key)
				}
				written, err := io.Copy(part, options.Attachment)
				if err != nil {
					return nil, errors.Errorf("Failed to write attachment to multipart form field %s", key)
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
		content, _ := ContentFromReader(body, writer.FormDataContentType())
		return content.Reader(), nil
	}
	return nil, errors.Errorf("Unsupported Payload: %s", payloadType.Kind().String())
}