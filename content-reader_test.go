package request_test

import (
	"bytes"
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

func TestShouldFailCreateContentFromNilReader(t *testing.T) {
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
func TestContentCanUnmarshallData(t *testing.T) {
	data := stuff{"12345"}
	payload, _ := json.Marshal(data)
	content := request.ContentWithData(payload)
	require.NotNil(t, content, "Content should not be nil")

	value := stuff{}
	err := content.Reader().UnmarshalContentJSON(&value)
	require.Nil(t, err, "Content failed unmarshaling, err=%+v", err)
	assert.Equal(t, data.ID, value.ID)
}