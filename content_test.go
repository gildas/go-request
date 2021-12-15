package request_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-request"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanCreateContentWithURL(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	url, _ := url.Parse("https://www.acme.com")
	content := request.ContentWithData(data, url)
	require.NotNil(t, content, "Content should not be nil")
	assert.Equal(t, int64(len(data)), content.Length)
	assert.Equal(t, data[0], content.Data[0])
	assert.Equal(t, url, content.URL)
}

func TestCanCreateContentWithType(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	mime := "image/png"
	content := request.ContentWithData(data, mime)
	require.NotNil(t, content, "Content should not be nil")
	assert.Equal(t, int64(len(data)), content.Length)
	assert.Equal(t, data[0], content.Data[0])
	assert.Equal(t, mime, content.Type)
}

func TestCanCreateContentWithLength(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data, len(data))
	require.NotNil(t, content, "Content should not be nil")
	assert.Equal(t, int64(len(data)), content.Length)
	assert.Equal(t, data[0], content.Data[0])
}

func TestCanCreateContentWithLength64(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data, int64(len(data)))
	require.NotNil(t, content, "Content should not be nil")
	assert.Equal(t, int64(len(data)), content.Length)
	assert.Equal(t, data[0], content.Data[0])
}

func TestCanCreateContentWithCookies(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	url, _ := url.Parse("https://www.acme.com")
	cookies := []*http.Cookie{{Name: "Test", Value: "1234", Secure: true, HttpOnly: true}}
	content := request.ContentWithData(data, url, cookies)
	require.NotNil(t, content, "Content should not be nil")
	assert.Equal(t, int64(len(data)), content.Length)
	assert.Equal(t, data[0], content.Data[0])
	assert.Equal(t, url, content.URL)
	assert.Equal(t, 1, len(content.Cookies))
	assert.Equal(t, "Test", content.Cookies[0].Name)
}

func TestCanCreateContentWithHeaders(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	url, _ := url.Parse("https://www.acme.com")
	header := http.Header{}
	header.Set("Custom-Header", "custom-value")
	content := request.ContentWithData(data, url, header)
	require.NotNil(t, content, "Content should not be nil")
	assert.Equal(t, int64(len(data)), content.Length)
	assert.Equal(t, data[0], content.Data[0])
	assert.Equal(t, url, content.URL)
	require.NotNil(t, content.Headers)
	assert.Equal(t, "custom-value", content.Headers.Get("Custom-Header"))
}

func TestCanCreateContentFromReader(t *testing.T) {
	data := bytes.NewBuffer([]byte{1, 2, 3, 4, 5})
	content, err := request.ContentFromReader(data)
	require.Nil(t, err, "Failed to create Content, err=%+v", err)
	require.NotNil(t, content, "Content should not be nil")
	assert.Equal(t, int64(5), content.Length)
}

func TestShouldFailCreateContentFromBogusReader(t *testing.T) {
	data := failingReader(0)
	_, err := request.ContentFromReader(data)
	require.NotNil(t, err, "Should fail create content")
}

func TestCanCreateReaderFromContent(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data)
	require.NotNil(t, content, "Content should not be nil")
	assert.Equal(t, int64(len(data)), content.Length)
	assert.Equal(t, data[0], content.Data[0])
	reader := content.Reader()
	require.NotNil(t, reader, "ContentReader should not be nil")
	require.Implements(t, (*io.Reader)(nil), reader)
}

func TestCanCreateReadCloserFromContent(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data)
	require.NotNil(t, content, "Content should not be nil")
	assert.Equal(t, int64(len(data)), content.Length)
	assert.Equal(t, data[0], content.Data[0])
	reader := content.ReadCloser()
	require.NotNil(t, reader, "ContentReader should not be nil")
	require.Implements(t, (*io.ReadCloser)(nil), reader)
}

func TestCanReadFromContentReader(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data)
	require.NotNil(t, content, "Content should not be nil")

	content.Length = 0 // just to force the length to be computed again
	reader := content.Reader()
	require.NotNil(t, reader, "ContentReader should not be nil")
	length, err := reader.Read(content.Data)
	require.Nil(t, err, "ContentReader should be able to read data, err=%+v", err)
	assert.Equal(t, 5, length, "ContentReader should have read 5 bytes")
	assert.Equal(t, data[0], content.Data[0])
}

func TestContentCanUnmarshallData(t *testing.T) {
	data := stuff{"12345"}
	payload, _ := json.Marshal(data)
	content := request.ContentWithData(payload)
	require.NotNil(t, content, "Content should not be nil")

	value := stuff{}
	err := content.UnmarshalContentJSON(&value)
	require.Nil(t, err, "Content failed unmarshaling, err=%+v", err)
	assert.Equal(t, data.ID, value.ID)
}

func TestShouldFailUnmarshallContentReaderWithBogusData(t *testing.T) {
	content := request.ContentWithData([]byte(`{"ID": 1234}`), "application/json")
	data := stuff{}
	err := content.UnmarshalContentJSON(&data)
	require.NotNil(t, err, "Should fail unmarshal content")
	assert.Truef(t, errors.Is(err, errors.JSONUnmarshalError), "Error should be a JSON Unmarshal Error")
	var details errors.Error
	require.True(t, errors.As(err, &details), "Error chain should contain an errors.Error")
	assert.Equal(t, "error.json.unmarshal", details.ID, "Error's ID is wrong (%s)", details.ID)
}

func TestContentShouldLogBinaryContent(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data)
	require.NotNil(t, content, "Content should not be nil")
	content.Type = "image/png"

	assert.Equal(t, "image/png, 5 bytes: 0102030405", content.LogString(10))
}

func TestContentShouldLogTextContent(t *testing.T) {
	data := []byte("Hello")
	content := request.ContentWithData(data)
	require.NotNil(t, content, "Content should not be nil")
	content.Type = "text/plain"

	assert.Equal(t, "text/plain, 5 bytes: Hello", content.LogString(10))
}

func TestContentShouldLogJSONContent(t *testing.T) {
	data := []byte(`{"data": "Hello"}`)
	content := request.ContentWithData(data)
	require.NotNil(t, content, "Content should not be nil")
	content.Type = "application/json"

	assert.Equal(t, `application/json, 17 bytes: {"data": "Hello"}`, content.LogString(20))
}
