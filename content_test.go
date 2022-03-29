package request_test

import (
	"bytes"
	"crypto/aes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/gildas/go-request"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type ContentSuite struct {
	suite.Suite
	Name   string
	Start  time.Time
	Logger *logger.Logger
}

func TestContentSuite(t *testing.T) {
	suite.Run(t, new(ContentSuite))
}

// *****************************************************************************
// Suite Tools

func (suite *ContentSuite) SetupSuite() {
	_ = godotenv.Load()
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
	suite.Logger = logger.Create("test",
		&logger.FileStream{
			Path:        fmt.Sprintf("./log/test-%s.log", strings.ToLower(suite.Name)),
			Unbuffered:  true,
			FilterLevel: logger.TRACE,
		},
	).Child("test", "test")
	suite.Logger.Infof("Suite Start: %s %s", suite.Name, strings.Repeat("=", 80-14-len(suite.Name)))
}

func (suite *ContentSuite) TearDownSuite() {
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
	}
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))
	suite.Logger.Close()
}

func (suite *ContentSuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
}

func (suite *ContentSuite) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	if suite.T().Failed() {
		suite.Logger.Errorf("Test %s failed", testName)
	}
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}

// *****************************************************************************

func (suite *ContentSuite) TestCanCreateWithURL() {
	data := []byte{1, 2, 3, 4, 5}
	url, _ := url.Parse("https://www.acme.com")
	content := request.ContentWithData(data, url)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal(uint64(len(data)), content.Length)
	suite.Assert().Equal(data[0], content.Data[0])
	suite.Assert().Equal(url, content.URL)
}

func (suite *ContentSuite) TestCanCreateWithType() {
	data := []byte{1, 2, 3, 4, 5}
	mime := "image/png"
	content := request.ContentWithData(data, mime)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal(uint64(len(data)), content.Length)
	suite.Assert().Equal(data[0], content.Data[0])
	suite.Assert().Equal(mime, content.Type)
}

func (suite *ContentSuite) TestCanCreateWithLength() {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data, len(data))
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal(uint64(len(data)), content.Length)
	suite.Assert().Equal(data[0], content.Data[0])
}

func (suite *ContentSuite) TestCanCreateWithUintLength() {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data, uint(len(data)))
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal(uint64(len(data)), content.Length)
	suite.Assert().Equal(data[0], content.Data[0])
}

func (suite *ContentSuite) TestCanCreateWithInt64Length() {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data, int64(len(data)))
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal(uint64(len(data)), content.Length)
	suite.Assert().Equal(data[0], content.Data[0])
}

func (suite *ContentSuite) TestCanCreateWithUint64Length() {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data, uint64(len(data)))
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal(uint64(len(data)), content.Length)
	suite.Assert().Equal(data[0], content.Data[0])
}

func (suite *ContentSuite) TestCanCreateWithCookies() {
	data := []byte{1, 2, 3, 4, 5}
	url, _ := url.Parse("https://www.acme.com")
	cookies := []*http.Cookie{{Name: "Test", Value: "1234", Secure: true, HttpOnly: true}}
	content := request.ContentWithData(data, url, cookies)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal(uint64(len(data)), content.Length)
	suite.Assert().Equal(data[0], content.Data[0])
	suite.Assert().Equal(url, content.URL)
	suite.Assert().Equal(1, len(content.Cookies))
	suite.Assert().Equal("Test", content.Cookies[0].Name)
}

func (suite *ContentSuite) TestCanCreateWithHeaders() {
	data := []byte{1, 2, 3, 4, 5}
	url, _ := url.Parse("https://www.acme.com")
	header := http.Header{}
	header.Set("Custom-Header", "custom-value")
	content := request.ContentWithData(data, url, header)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal(uint64(len(data)), content.Length)
	suite.Assert().Equal(data[0], content.Data[0])
	suite.Assert().Equal(url, content.URL)
	suite.Require().NotNil(content.Headers)
	suite.Assert().Equal("custom-value", content.Headers.Get("Custom-Header"))
}

func (suite *ContentSuite) TestCanCreateFromReader() {
	data := bytes.NewBuffer([]byte{1, 2, 3, 4, 5})
	content, err := request.ContentFromReader(data)
	suite.Require().NoErrorf(err, "Failed to create Content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal(uint64(5), content.Length)
}

func (suite *ContentSuite) TestShouldFailCreateFromBogusReader() {
	data := failingReader(0)
	_, err := request.ContentFromReader(data)
	suite.Require().Error(err, "Should fail create content")
}

func (suite *ContentSuite) TestCanCreateReaderFromContent() {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal(uint64(len(data)), content.Length)
	suite.Assert().Equal(data[0], content.Data[0])
	reader := content.Reader()
	suite.Require().NotNil(reader, "ContentReader should not be nil")
	suite.Require().Implements((*io.Reader)(nil), reader)
}

func (suite *ContentSuite) TestCanCreateReadCloserFromContent() {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal(uint64(len(data)), content.Length)
	suite.Assert().Equal(data[0], content.Data[0])
	reader := content.ReadCloser()
	suite.Require().NotNil(reader, "ContentReader should not be nil")
	suite.Require().Implements((*io.ReadCloser)(nil), reader)
}

func (suite *ContentSuite) TestCanReadFromContentReader() {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data)
	suite.Require().NotNil(content, "Content should not be nil")

	content.Length = 0 // just to force the length to be computed again
	reader := content.Reader()
	suite.Require().NotNil(reader, "ContentReader should not be nil")
	length, err := reader.Read(content.Data)
	suite.Require().NoErrorf(err, "ContentReader should be able to read data, err=%+v", err)
	suite.Assert().Equal(5, length, "ContentReader should have read 5 bytes")
	suite.Assert().Equal(data[0], content.Data[0])
}

func (suite *ContentSuite) TestCanUnmarshallData() {
	data := stuff{"12345"}
	payload, _ := json.Marshal(data)
	content := request.ContentWithData(payload)
	suite.Require().NotNil(content, "Content should not be nil")

	value := stuff{}
	err := content.UnmarshalContentJSON(&value)
	suite.Require().NoErrorf(err, "Content failed unmarshaling, err=%+v", err)
	suite.Assert().Equal(data.ID, value.ID)
}

func (suite *ContentSuite) TestShouldFailUnmarshallContentWithBogusData() {
	content := request.ContentWithData([]byte(`{"ID": 1234}`), "application/json")
	data := stuff{}
	err := content.UnmarshalContentJSON(&data)
	suite.Require().Error(err, "Should fail unmarshal content")
	suite.Assert().Truef(errors.Is(err, errors.JSONUnmarshalError), "Error should be a JSON Unmarshal Error")
	var details errors.Error
	suite.Require().True(errors.As(err, &details), "Error chain should contain an errors.Error")
	suite.Assert().Equal("error.json.unmarshal", details.ID, "Error's ID is wrong (%s)", details.ID)
}

func (suite *ContentSuite) TestShouldLogBinaryContent() {
	data := []byte{1, 2, 3, 4, 5}
	content := request.ContentWithData(data)
	suite.Require().NotNil(content, "Content should not be nil")
	content.Type = "image/png"

	suite.Assert().Equal("image/png, 5 bytes: 0102030405", content.LogString(10))
}

func (suite *ContentSuite) TestShouldLogTextContent() {
	data := []byte("Hello")
	content := request.ContentWithData(data)
	suite.Require().NotNil(content, "Content should not be nil")
	content.Type = "text/plain"

	suite.Assert().Equal("text/plain, 5 bytes: Hello", content.LogString(10))
}

func (suite *ContentSuite) TestShouldLogJSONContent() {
	data := []byte(`{"data": "Hello"}`)
	content := request.ContentWithData(data)
	suite.Require().NotNil(content, "Content should not be nil")
	content.Type = "application/json"

	suite.Assert().Equal(`application/json, 17 bytes: {"data": "Hello"}`, content.LogString(20))
}

func (suite *ContentSuite) TestCanCreateFromCompressedData() {
	data := []byte{
		31, 139, 8, 0, 0, 0, 0, 0, 2, 255, 69, 142, 61, 15, 130, 64, 12, 134, 119, 126, 69, 195, 172,
		23, 7, 38, 86, 212, 17, 22, 157, 77, 229, 10, 185, 8, 215, 179, 119, 136, 9, 225, 191, 91, 162,
		137, 75, 147, 231, 125, 250, 181, 100, 0, 57, 137, 176, 228, 37, 44, 10, 138, 45, 91, 82, 42,
		14, 197, 238, 27, 140, 20, 35, 246, 91, 166, 110, 52, 61, 115, 63, 144, 193, 16, 162, 233, 4,
		71, 154, 89, 30, 70, 232, 57, 81, 76, 166, 66, 207, 222, 181, 56, 84, 186, 230, 244, 110, 41,
		36, 199, 190, 132, 154, 65, 120, 134, 142, 39, 111, 181, 10, 56, 11, 206, 191, 112, 112, 118,
		111, 9, 237, 157, 168, 131, 109, 70, 91, 155, 203, 237, 220, 92, 235, 99, 254, 123, 32, 38,
		76, 83, 220, 238, 255, 149, 154, 53, 91, 179, 15, 240, 168, 235, 9, 193, 0, 0, 0,
	}
	expected := `{
  "error": {
    "code": 404,
    "message": "com.google.apps.framework.request.CanonicalCodeException: No row found for id invalid-deadbeef Code: NOT_FOUND",
    "status": "NOT_FOUND"
  }
}
`
	headers := http.Header{}
	headers.Set("Content-Encoding", "gzip")
	headers.Set("Content-Type", "application/json")
	content := request.ContentWithData(data, len(data), "application/json", headers)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("application/json", content.Type)
	suite.Assert().Equalf(len(expected), int(content.Length), "Content length should be %d", len(expected))
	suite.Assert().Equal(expected, string(content.Data))
}

func (suite *ContentSuite) TestCanCreateFromBogusCompressedData() {
	data := []byte{
		31, 139, 8, 0, 0, 0, 0, 0, 2, 255, 69, 142, 61, 15, 130, 64, 12, 134, 119, 126, 69, 195, 172,
		23, 7, 38, 86, 212, 17, 22, 157, 77, 229, 10, 185, 8, 215, 179, 119, 136, 9, 225, 191, 91, 162,
		36, 199, 190, 132, 154, 65, 120, 134, 142, 39, 111, 181, 10, 56, 11, 206, 191, 112, 112, 118,
		111, 9, 237, 157, 168, 131, 109, 70, 91, 155, 203, 237, 220, 92, 235, 99, 254, 123, 32, 38,
		76, 83, 220, 238, 255, 149, 154, 53, 91, 179, 15, 240, 168, 235, 9, 193, 0, 0, 0,
	}
	headers := http.Header{}
	headers.Set("Content-Encoding", "gzip")
	headers.Set("Content-Type", "application/json")
	content := request.ContentWithData(data, len(data), "application/json", headers)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("application/json", content.Type)
	suite.Assert().Equalf(len(data), int(content.Length), "Content length should be %d", len(data))
}

func (suite *ContentSuite) TestCanCreateFromBogusCompressedDataHeader() {
	data := []byte{
		23, 7, 38, 86, 212, 17, 22, 157, 77, 229, 10, 185, 8, 215, 179, 119, 136, 9, 225, 191, 91, 162,
		137, 75, 147, 231, 125, 250, 181, 100, 0, 57, 137, 176, 228, 37, 44, 10, 138, 45, 91, 82, 42,
		14, 197, 238, 27, 140, 20, 35, 246, 91, 166, 110, 52, 61, 115, 63, 144, 193, 16, 162, 233, 4,
		71, 154, 89, 30, 70, 232, 57, 81, 76, 166, 66, 207, 222, 181, 56, 84, 186, 230, 244, 110, 41,
		36, 199, 190, 132, 154, 65, 120, 134, 142, 39, 111, 181, 10, 56, 11, 206, 191, 112, 112, 118,
		111, 9, 237, 157, 168, 131, 109, 70, 91, 155, 203, 237, 220, 92, 235, 99, 254, 123, 32, 38,
		76, 83, 220, 238, 255, 149, 154, 53, 91, 179, 15, 240, 168, 235, 9, 193, 0, 0, 0,
	}
	headers := http.Header{}
	headers.Set("Content-Encoding", "gzip")
	headers.Set("Content-Type", "application/json")
	content := request.ContentWithData(data, len(data), "application/json", headers)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("application/json", content.Type)
	suite.Assert().Equalf(len(data), int(content.Length), "Content length should be %d", len(data))
}

func (suite *ContentSuite) TestCanMarshal() {
	data := []byte{1, 2, 3, 4, 5}
	url, _ := url.Parse("https://www.acme.com")
	header := http.Header{}
	header.Set("Custom-Header", "custom-value")
	cookie := http.Cookie{
		Name:    "name",
		Value:   "value",
		Path:    "/",
		Expires: time.Date(2022, 3, 4, 10, 12, 30, 0, time.UTC),
	}
	content := request.ContentWithData(data, url, header, "image/png", []*http.Cookie{&cookie})
	suite.Require().NotNil(content, "Content should not be nil")

	expected := `
	{
		"Type":    "image/png",
		"url":     "https://www.acme.com",
		"Length":  5,
		"headers": {"Custom-Header":["custom-value"]},
		"cookies": [{
			"name":  "name",
			"value": "value",
			"path":  "/",
			"expires": "2022-03-04T10:12:30Z"
		}],
		"Data":    "AQIDBAU="
	}`
	payload, err := json.Marshal(content)
	suite.Require().NoErrorf(err, "Failed to marshal content, error: %s", err)
	suite.JSONEq(expected, string(payload))
}

func (suite *ContentSuite) TestCanUmMarshal() {
	payload := `
	{
		"Type":    "image/png",
		"url":     "https://www.acme.com",
		"Length":  5,
		"headers": {"Custom-Header":["custom-value"]},
		"cookies": [{
			"name":  "name",
			"value": "value",
			"path":  "/",
			"expires": "2022-03-04T10:12:30Z"
		}],
		"Data":    "AQIDBAU="
	}`

	content := request.Content{}
	err := json.Unmarshal([]byte(payload), &content)
	suite.Require().NoErrorf(err, "Failed to unmarshal content, error: %s", err)
	suite.Assert().Equal("image/png", content.Type)
	suite.Require().NotNil(content.URL)
	suite.Assert().Equal("https://www.acme.com", content.URL.String())
	suite.Assert().Equal("custom-value", content.Headers.Get("Custom-Header"))
	suite.Require().Len(content.Cookies, 1, "There should be 1 cookie")
	suite.Require().NotNil(content.Cookies[0])
	suite.Assert().Equal("name", content.Cookies[0].Name)
	suite.Assert().Equal("value", content.Cookies[0].Value)
	suite.Assert().Equal("/", content.Cookies[0].Path)
	suite.Assert().Equal(time.Date(2022, 3, 4, 10, 12, 30, 0, time.UTC), content.Cookies[0].Expires)
	suite.Assert().False(content.Cookies[0].Secure)
	suite.Assert().Equal(uint64(5), content.Length)
	suite.Assert().Equal([]byte{1, 2, 3, 4, 5}, content.Data)
}

func (suite *ContentSuite) TestShouldFailUnmarshallWithBogusData() {
	payload := `
	{
		"Type":    1,
		"Length":  5,
		"Data":    "AQIDBAU="
	}`

	content := request.Content{}
	err := json.Unmarshal([]byte(payload), &content)
	suite.Require().Error(err)
	suite.Assert().ErrorIs(err, errors.JSONUnmarshalError, "Error should be a JSON Unmarshal Error")
}

func (suite *ContentSuite) TestCanMarshalCryptoAlgorithm() {
	algorithm := request.NONE
	payload, err := json.Marshal(algorithm)
	suite.Require().NoErrorf(err, "Failed to marshal algorithm, error: %s", err)
	suite.Assert().Equal(`"NONE"`, string(payload))

	algorithm = request.AESCTR
	payload, err = json.Marshal(algorithm)
	suite.Require().NoErrorf(err, "Failed to marshal algorithm, error: %s", err)
	suite.Assert().Equal(`"AESCTR"`, string(payload))
}

func (suite *ContentSuite) TestCanUnmarshalCryptoAlgorithm() {
	payload := `"AESCTR"`
	algorithm := request.CryptoAlgorithm(0)
	err := json.Unmarshal([]byte(payload), &algorithm)
	suite.Require().NoErrorf(err, "Failed to unmarshal algorithm, error: %s", err)
	suite.Assert().Equal(request.AESCTR, algorithm)

	payload = `"NONE"`
	algorithm = request.CryptoAlgorithm(0)
	err = json.Unmarshal([]byte(payload), &algorithm)
	suite.Require().NoErrorf(err, "Failed to unmarshal algorithm, error: %s", err)
	suite.Assert().Equal(request.NONE, algorithm)

	payload = `1`
	algorithm = request.CryptoAlgorithm(0)
	err = json.Unmarshal([]byte(payload), &algorithm)
	suite.Require().Errorf(err, "Should have failed to unmarshal algorithm")
	suite.Assert().ErrorIs(err, errors.JSONUnmarshalError, "Error should be a JSON Unmarshal Error")

	payload = `"INVALID"`
	algorithm = request.CryptoAlgorithm(0)
	err = json.Unmarshal([]byte(payload), &algorithm)
	suite.Require().Errorf(err, "Should have failed to unmarshal algorithm")
	suite.Assert().ErrorIs(err, errors.JSONUnmarshalError, "Error should be a JSON Unmarshal Error")
	suite.Assert().ErrorIs(err, errors.ArgumentInvalid, "Error should be an Argument Invalid Error")
	details := errors.ArgumentInvalid.Clone()
	suite.Require().ErrorAs(err, &details, "Error should contain an Invalid Error")
	suite.Assert().Equal("algorithm", details.What)
	suite.Assert().Equal("INVALID", details.Value)
}

func (suite *ContentSuite) TestCanDecryptWithNONE() {
	encrypted := []byte{0x17, 0x07, 0x26, 0x56, 0xd4, 0x11, 0x16, 0x9d, 0x4d, 0xe5, 0x0a, 0xb9, 0x08, 0xd7, 0xb3, 0x3b}
	decrypted := []byte{0x17, 0x07, 0x26, 0x56, 0xd4, 0x11, 0x16, 0x9d, 0x4d, 0xe5, 0x0a, 0xb9, 0x08, 0xd7, 0xb3, 0x3b}

	key := []byte{}
	content := request.Content{
		Type:   "image/jpeg",
		Length: uint64(len(encrypted)),
		Data:   encrypted,
	}

	suite.Assert().Equal("NONE", request.NONE.String())
	decryptedContent, err := content.Decrypt(request.NONE, key)
	suite.Require().NoError(err, "Failed to decrypt content")
	suite.Require().Lenf(decryptedContent.Data, len(decrypted), "Decrypted content should be %d", len(decrypted))
	suite.Require().Equal(decryptedContent.Length, uint64(len(decrypted)), "Decrypted content should be %d", len(decrypted))
	suite.Assert().Equal(decrypted, decryptedContent.Data, "Decrypted content is incorrect")
}

func (suite *ContentSuite) TestCanDecryptWithAESCTR() {
	encrypted := []byte{0xa9, 0x09, 0x20, 0xf8, 0x77, 0x58, 0x30, 0xee, 0x91, 0x22, 0x18, 0x5c, 0x1a, 0xfd, 0x2d, 0xf2}
	decrypted := []byte{0x17, 0x07, 0x26, 0x56, 0xd4, 0x11, 0x16, 0x9d, 0x4d, 0xe5, 0x0a, 0xb9, 0x08, 0xd7, 0xb3, 0x3b}
	key, _ := hex.DecodeString("FE400803E0BA6A4B3D611305C1A2EDE263A4D599C96F2BA49C837FD4E193C76D")
	content := request.Content{
		Type:   "image/jpeg",
		Length: uint64(len(encrypted)),
		Data:   encrypted,
	}

	suite.Assert().Equal("AESCTR", request.AESCTR.String())
	decryptedContent, err := content.Decrypt(request.AESCTR, key)
	suite.Require().NoError(err, "Failed to decrypt content")
	suite.Require().Lenf(decryptedContent.Data, len(decrypted), "Decrypted content should be %d", len(decrypted))
	suite.Require().Equal(decryptedContent.Length, uint64(len(decrypted)), "Decrypted content should be %d", len(decrypted))
	suite.Assert().Equal(decrypted, decryptedContent.Data, "Decrypted content is incorrect")
}

func (suite *ContentSuite) TestCanEncryptWithNONE() {
	encrypted := []byte{0x17, 0x07, 0x26, 0x56, 0xd4, 0x11, 0x16, 0x9d, 0x4d, 0xe5, 0x0a, 0xb9, 0x08, 0xd7, 0xb3, 0x3b}
	decrypted := []byte{0x17, 0x07, 0x26, 0x56, 0xd4, 0x11, 0x16, 0x9d, 0x4d, 0xe5, 0x0a, 0xb9, 0x08, 0xd7, 0xb3, 0x3b}

	key := []byte{}
	content := request.Content{
		Type:   "image/jpeg",
		Length: uint64(len(encrypted)),
		Data:   decrypted,
	}

	suite.Assert().Equal("NONE", request.NONE.String())
	encryptedContent, err := content.Encrypt(request.NONE, key)
	suite.Require().NoError(err, "Failed to decrypt content")
	suite.Require().Lenf(encryptedContent.Data, len(decrypted), "Encrypted content should be %d", len(encrypted))
	suite.Require().Equal(encryptedContent.Length, uint64(len(decrypted)), "Encrypted content should be %d", len(encrypted))
	suite.Assert().Equal(encrypted, encryptedContent.Data, "Encrypted content is incorrect")
}

func (suite *ContentSuite) TestCanEncryptWithAESCTR() {
	encrypted := []byte{0xa9, 0x09, 0x20, 0xf8, 0x77, 0x58, 0x30, 0xee, 0x91, 0x22, 0x18, 0x5c, 0x1a, 0xfd, 0x2d, 0xf2}
	decrypted := []byte{0x17, 0x07, 0x26, 0x56, 0xd4, 0x11, 0x16, 0x9d, 0x4d, 0xe5, 0x0a, 0xb9, 0x08, 0xd7, 0xb3, 0x3b}
	key, _ := hex.DecodeString("FE400803E0BA6A4B3D611305C1A2EDE263A4D599C96F2BA49C837FD4E193C76D")
	content := request.Content{
		Type:   "image/jpeg",
		Length: uint64(len(decrypted)),
		Data:   decrypted,
	}

	suite.Assert().Equal("AESCTR", request.AESCTR.String())
	encryptedContent, err := content.Encrypt(request.AESCTR, key)
	suite.Require().NoError(err, "Failed to encrypt content")
	suite.Require().Lenf(encryptedContent.Data, len(encrypted), "Encrypted content should be %d", len(encrypted))
	suite.Require().Equal(encryptedContent.Length, uint64(len(encrypted)), "Encrypted content should be %d", len(encrypted))
	suite.Assert().Equal(encrypted, encryptedContent.Data, "Encrypted content is incorrect")
}

func (suite *ContentSuite) TestShouldFailDecryptWithInvalidAlgorithm() {
	_, err := request.Content{}.Decrypt(request.CryptoAlgorithm(10), []byte{})
	suite.Require().Error(err)
	suite.Logger.Errorf("Expected Error!", err)
	suite.Assert().ErrorIs(err, errors.InvalidType, "Error should be an Invalid Type Error")
	var details errors.Error
	suite.Require().ErrorAs(err, &details)
	suite.Assert().Equal("Unknown 10", details.What)
}

func (suite *ContentSuite) TestShouldFailDecryptWithInvalidKey() {
	key := []byte{1, 2, 3, 4, 5}
	_, err := request.Content{}.Decrypt(request.AESCTR, key)
	suite.Require().Error(err)
	suite.Logger.Errorf("Expected Error!", err)
	suite.Assert().ErrorIs(err, errors.ArgumentInvalid, "Error should be an Argument Invalid Error")
	var details errors.Error
	suite.Require().ErrorAs(err, &details)
	suite.Assert().Equal("key", details.What)
	suite.Assert().Equal(key, details.Value.([]byte))
	suite.Assert().ErrorIs(err, aes.KeySizeError(len(key)), "Error should be a KeySizeError")
	suite.Require().NotNil(details.HasCauses(), "Error should have a cause")
	suite.Assert().Equal("crypto/aes: invalid key size 5", details.Causes[0].Error())
}

func (suite *ContentSuite) TestShouldFailEncryptWithInvalidAlgorithm() {
	_, err := request.Content{}.Encrypt(request.CryptoAlgorithm(10), []byte{})
	suite.Require().Error(err)
	suite.Logger.Errorf("Expected Error!", err)
	suite.Assert().ErrorIs(err, errors.InvalidType, "Error should be an Invalid Type Error")
	var details errors.Error
	suite.Require().ErrorAs(err, &details)
	suite.Assert().Equal("Unknown 10", details.What)
}

func (suite *ContentSuite) TestShouldFailEncryptWithInvalidKey() {
	key := []byte{1, 2, 3, 4, 5}
	_, err := request.Content{}.Encrypt(request.AESCTR, key)
	suite.Require().Error(err)
	suite.Logger.Errorf("Expected Error!", err)
	suite.Assert().ErrorIs(err, errors.ArgumentInvalid, "Error should be an Argument Invalid Error")
	var details errors.Error
	suite.Require().ErrorAs(err, &details)
	suite.Assert().Equal("key", details.What)
	suite.Assert().Equal(key, details.Value.([]byte))
	suite.Assert().ErrorIs(err, aes.KeySizeError(len(key)), "Error should be a KeySizeError")
	suite.Require().NotNil(details.HasCauses(), "Error should have a cause")
	suite.Assert().Equal("crypto/aes: invalid key size 5", details.Causes[0].Error())
}
