package core

import (
	"encoding/json"
	"bytes"
	"io"
	"io/ioutil"
	"net/url"
)

// Content defines some content
type Content struct {
	Type   string        `json:"contentType"`
	URL    *url.URL      `json:"contentUrl"`
	Length int64         `json:"contentLength"`
	Data   []byte        `json:"contentData"`
}

// ContentReader defines a content Reader (like data from an HTTP request)
type ContentReader struct {
	Type   string        `json:"contentType"`
	Length int64         `json:"contentLength"`
	Reader io.ReadCloser `json:"-"`
}

// ContentWithData instantiates a Content from a simple byte array
func ContentWithData(data []byte, options... interface{}) *Content {
	content := &Content{}
	content.Data = data
	for _, option := range options {
		if u, ok := option.(*url.URL); ok {
			content.URL = u
		} else if t, ok := option.(string); ok {
			content.Type = t
		} else if l, ok := option.(int64); ok  && l > 0 {
			content.Length = l
		} else if l, ok := option.(int); ok  && l > 0 {
			content.Length = int64(l)
		}
	}
	if content.Length == 0 {
		content.Length = int64(len(content.Data))
	}
	return content
}

// ContentFromReader instantiates a Content from an I/O reader
func ContentFromReader(reader io.Reader, options... interface{}) (*Content, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return ContentWithData(data, options...), nil
}

// ReadContent instantiates a Content from an I/O reader
func (reader ContentReader) ReadContent(options... interface{}) (*Content, error) {
	return ContentFromReader(reader, options...)
}

// UnmarshalContentJSON reads the content of an I/O reader and unmarshals it into JSON
func (reader ContentReader) UnmarshalContentJSON(v interface{}) (err error) {
	content, err := reader.ReadContent()
	if err != nil { return err }
	return json.Unmarshal(content.Data, &v)
}

// Reader gets a ContentReader from this Content
func (content *Content) Reader() *ContentReader {
	reader := &ContentReader{
		Type:   content.Type,
		Length: content.Length,
		Reader: ioutil.NopCloser(bytes.NewReader(content.Data)),
	}
	if reader.Length == 0 {
		reader.Length = int64(len(content.Data))
	}
	return reader
}

func (reader ContentReader) Read(data []byte) (int, error) {
	return reader.Reader.Read(data)
}