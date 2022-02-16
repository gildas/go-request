package request

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
)

// Content defines some content
type Content struct {
	Type    string         `json:"contentType"`
	URL     *url.URL       `json:"contentUrl"`
	Length  int64          `json:"contentLength"`
	Data    []byte         `json:"contentData"`
	Headers http.Header    `json:"contentHeaders"`
	Cookies []*http.Cookie `json:"cookies"`
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
			content.Length = option
		case int:
			content.Length = int64(option)
		case string:
			content.Type = option
		case http.Header:
			content.Headers = option
		case []*http.Cookie:
			content.Cookies = option
		}
	}
	if content.Length == 0 {
		content.Length = int64(len(content.Data))
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

// UnmarshalContentJSON reads the content of an I/O reader and unmarshals it into JSON
func (content Content) UnmarshalContentJSON(v interface{}) (err error) {
	if err = json.Unmarshal(content.Data, &v); err != nil {
		return errors.JSONUnmarshalError.Wrap(err)
	}
	return nil
}

func (content Content) LogString(maxSize uint64) string {
	sb := strings.Builder{}
	sb.WriteString(content.Type)
	sb.WriteString(", ")
	sb.WriteString(strconv.FormatInt(content.Length, 10))
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
