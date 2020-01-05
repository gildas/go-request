package request_test

import (
	"bytes"
	"io/ioutil"
	"encoding/json"
	"github.com/gildas/go-request"
	"github.com/gildas/go-errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"testing"
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

func TestCanCreateContentFromReader(t *testing.T) {
	data := bytes.NewBuffer([]byte{1, 2, 3, 4, 5})
	content, err := request.ContentFromReader(data)
	require.Nil(t, err, "Failed to create Content, err=%+v", err)
	require.NotNil(t, content, "Content should not be nil")
	assert.Equal(t, int64(5), content.Length)
}

func TestCanCreateContentFromContentReader(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data)
	require.NotNil(t, content, "Content should not be nil")
	reader := content.Reader()
	require.NotNil(t, reader, "ContentReader should not be nil")
	another, err := reader.ReadContent()
	require.Nil(t, err, "Failed to create Content, err=%+v", err)
	require.Equal(t, another, content)
}

type failingReader int
func (r failingReader) Read(data []byte) (int, error) {
	return 0, errors.NotImplementedError.New()
}
func (r failingReader)Close() error {
	return nil
}

func TestShouldFailCreateContentFromBogusReader(t *testing.T) {
	data := failingReader(0)
	_, err := request.ContentFromReader(data)
	require.NotNil(t, err, "Should fail create content")
}

func TestCanCreateContentReaderFromContent(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data)
	require.NotNil(t, content, "Content should not be nil")
	assert.Equal(t, int64(len(data)), content.Length)
	assert.Equal(t, data[0], content.Data[0])
	reader := content.Reader()
	require.NotNil(t, reader, "ContentReader should not be nil")
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

type stuff struct {
	ID string
}
func TestContentReaderCanUnmarshallData(t *testing.T) {
	data := stuff{"12345"}
	payload, _ := json.Marshal(data)
	content := request.ContentWithData(payload)
	require.NotNil(t, content, "Content should not be nil")

	value := stuff{}
	err := content.Reader().UnmarshalContentJSON(&value)
	require.Nil(t, err, "Content failed unmarshaling, err=%+v", err)
	assert.Equal(t, data.ID, value.ID)
}

func TestShouldFailUnmarshallContentReaderWithBogusReader(t *testing.T) {
	reader := request.ContentReader{"application/json", 0, failingReader(0)}
	data := stuff{}
	err := reader.UnmarshalContentJSON(&data)
	require.NotNil(t, err, "Should fail unmarshal content")
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestShouldFailUnmarshallContentReaderWithBogusData(t *testing.T) {
	reader := request.ContentReader{"application/json", 0, ioutil.NopCloser(bytes.NewBufferString(`{"ID": 1234}`))}
	data := stuff{}
	err := reader.UnmarshalContentJSON(&data)
	require.NotNil(t, err, "Should fail unmarshal content")
	assert.Truef(t, errors.Is(err, errors.JSONUnmarshalError), "Error should be a JSON Unmarshal Error")
	var details *errors.Error
	require.True(t, errors.As(err, &details), "Error chain should contain an errors.Error")
	assert.Equal(t, "error.json.unmarshal", details.ID, "Error's ID is wrong (%s)", details.ID)
}

func TestContentReaderShouldHaveSamePropertiesAsContent(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data)
	require.NotNil(t, content, "Content should not be nil")
	content.Type = "image/png"

	reader := content.Reader()
	require.NotNil(t, reader, "ContentReader should not be nil")
	assert.Equal(t, content.Type, reader.Type, "ContentReader does not have the same type as the Content")
	assert.Equal(t, content.Length, reader.Length, "ContentReader does not have the same length as the Content")
}

func TestContentShouldHaveSamePropertiesAsContentReader(t *testing.T) {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data)
	require.NotNil(t, content, "Content should not be nil")
	content.Type = "image/png"

	reader := content.Reader()
	require.NotNil(t, reader, "ContentReader should not be nil")
	assert.Equal(t, content.Type, reader.Type, "ContentReader does not have the same type as the Content")
	assert.Equal(t, content.Length, reader.Length, "ContentReader does not have the same length as the Content")

	another, err := reader.ReadContent()
	require.Nil(t, err, "Failed to create Content, err=%+v", err)
	assert.Equal(t, reader.Type, another.Type, "Content does not have the same type as the ContentReader")
	assert.Equal(t, reader.Length, another.Length, "Content does not have the same length as the ContentReader")
}