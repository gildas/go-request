package request

import (
	"bytes"
	"compress/gzip"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
)

// Content defines some content
type Content struct {
	Type    string         `json:"Type"`
	Name    string         `json:"Name,omitempty"`
	URL     *url.URL       `json:"-"`
	Length  uint64         `json:"Length"`
	Data    []byte         `json:"Data"`
	Headers http.Header    `json:"headers,omitempty"`
	Cookies []*http.Cookie `json:"-"`
}

// ContentWithData instantiates a Content from a simple byte array
func ContentWithData(data []byte, options ...interface{}) *Content {
	log := logger.Create("REQUEST", &logger.NilStream{})
	content := &Content{}
	content.Data = data
	for _, raw := range options {
		switch option := raw.(type) {
		case *url.URL:
			content.URL = option
		case *logger.Logger:
			log = option
		case int64:
			content.Length = uint64(option)
		case uint64:
			content.Length = option
		case uint:
			content.Length = uint64(option)
		case int:
			content.Length = uint64(option)
		case string:
			content.Type = option
		case http.Header:
			content.Headers = option
		case []*http.Cookie:
			content.Cookies = option
		}
	}
	if content.Headers.Get("Content-Encoding") == "gzip" {
		log.Tracef("Content is gzipped (%d bytes)", len(content.Data))
		buffer := bytes.NewBuffer(content.Data)
		if reader, err := gzip.NewReader(buffer); err == nil {
			if uncompressed, err := io.ReadAll(reader); err == nil {
				content.Data = uncompressed
				content.Length = uint64(len(uncompressed))
				log.Tracef("Uncompressed data (%d bytes)", content.Length)
			} else {
				log.Errorf("Failed to uncompress data", err)
			}
			reader.Close()
		} else {
			log.Errorf("Failed to create a GZIP reader", err)
		}
	}
	if content.Length == 0 {
		content.Length = uint64(len(content.Data))
	}
	if content.Headers == nil {
		content.Headers = http.Header{}
	}
	return content
}

// ContentFromReader instantiates a Content from an I/O reader
func ContentFromReader(reader io.Reader, options ...interface{}) (*Content, error) {
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return ContentWithData(data, options...), nil
}

// Reader gets an io.Reader from this Content
func (content *Content) Reader() io.Reader {
	return bytes.NewReader(content.Data)
}

// ReadCloser gets an io.ReadCloser from this Content
func (content *Content) ReadCloser() io.ReadCloser {
	return io.NopCloser(bytes.NewReader(content.Data))
}

// UnmarshalContentJSON unmarshals its Data into JSON
func (content Content) UnmarshalContentJSON(v interface{}) (err error) {
	if err = json.Unmarshal(content.Data, &v); err != nil {
		return errors.JSONUnmarshalError.WrapIfNotMe(err)
	}
	return nil
}

// LogString generates a string suitable for logging
func (content Content) LogString(maxSize uint64) string {
	sb := strings.Builder{}
	sb.WriteString(content.Type)
	sb.WriteString(", ")
	sb.WriteString(strconv.FormatUint(content.Length, 10))
	sb.WriteString(" bytes")
	if maxSize > 0 {
		if len(content.Data) > 0 {
			sb.WriteString(": ")
			switch {
			case strings.HasPrefix(content.Type, "application/json"):
				fallthrough
			case strings.HasPrefix(content.Type, "application/xml"):
				fallthrough
			case strings.HasPrefix(content.Type, "text/"):
				sb.WriteString(string(content.Data[:int(math.Min(float64(maxSize), float64(content.Length)))]))
			default:
				sb.WriteString(hex.EncodeToString(content.Data[:int(math.Min(float64(maxSize), float64(content.Length)))]))
			}
		}
	}
	sb.WriteString("")
	return sb.String()
}

// MarshalJSON marshals the Content into JSON
//
// implements json.Marshaler
func (content Content) MarshalJSON() ([]byte, error) {
	type surrogate Content
	cookies := make([]*cookie, len(content.Cookies))
	for i, c := range content.Cookies {
		cookies[i] = (*cookie)(c)
	}
	data, err := json.Marshal(struct {
		surrogate
		URL     *core.URL `json:"url,omitempty"`
		Cookies []*cookie `json:"cookies,omitempty"`
	}{
		surrogate: surrogate(content),
		URL:       (*core.URL)(content.URL),
		Cookies:   cookies,
	})
	return data, errors.JSONMarshalError.Wrap(err)
}

// UnmarshalJSON unmarshals the Content from JSON
//
// implements json.Unmarshaler
func (content *Content) UnmarshalJSON(payload []byte) error {
	type surrogate Content
	var inner struct {
		surrogate
		URL     *core.URL `json:"url,omitempty"`
		Cookies []*cookie `json:"cookies,omitempty"`
	}
	if err := json.Unmarshal(payload, &inner); err != nil {
		return errors.JSONUnmarshalError.WrapIfNotMe(err)
	}
	*content = Content(inner.surrogate)
	content.URL = (*url.URL)(inner.URL)
	if len(inner.Cookies) > 0 {
		content.Cookies = make([]*http.Cookie, len(inner.Cookies))
		for i, c := range inner.Cookies {
			content.Cookies[i] = (*http.Cookie)(c)
		}
	}
	return nil
}
