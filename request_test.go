package request_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
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

type RequestSuite struct {
	suite.Suite
	Name   string
	Server *httptest.Server
	Proxy  *httptest.Server
	Logger *logger.Logger
	Start  time.Time
}

func TestRequestSuite(t *testing.T) {
	suite.Run(t, new(RequestSuite))
}

func (suite *RequestSuite) TestCheckTestServer() {
	suite.Require().NotNil(suite.Server)
	serverURL, err := url.Parse(suite.Server.URL)
	suite.Require().Nil(err)
	suite.Require().NotNil(serverURL)
	suite.T().Logf("Server URL: %s", serverURL.String())
}

func (suite *RequestSuite) TestCheckTestProxy() {
	suite.Require().NotNil(suite.Proxy)
	proxyURL, err := url.Parse(suite.Proxy.URL)
	suite.Require().Nil(err)
	suite.Require().NotNil(proxyURL)
	suite.T().Logf("Proxy URL: %s", proxyURL.String())
}

func (suite *RequestSuite) TestCanSendRequestWithURL() {
	serverURL, _ := url.Parse(suite.Server.URL)
	reader, err := request.Send(&request.Options{
		URL:    serverURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithProxy() {
	serverURL, _ := url.Parse(suite.Server.URL)
	proxyURL, _ := url.Parse(suite.Proxy.URL)
	reader, err := request.Send(&request.Options{
		URL:      serverURL,
		Proxy:    proxyURL,
		Attempts: 1,
		Logger:   suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithLogSizeOptions() {
	serverURL, _ := url.Parse(suite.Server.URL)
	reader, err := request.Send(&request.Options{
		URL:                 serverURL,
		ResponseBodyLogSize: 4096,
		RequestBodyLogSize:  4096,
		Logger:              suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithLogSizeOffOptions() {
	serverURL, _ := url.Parse(suite.Server.URL)
	reader, err := request.Send(&request.Options{
		URL:                 serverURL,
		ResponseBodyLogSize: -1,
		RequestBodyLogSize:  -1,
		Logger:              suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithHeaders() {
	serverURL, _ := url.Parse(suite.Server.URL)
	reader, err := request.Send(&request.Options{
		URL: serverURL,
		Headers: map[string]string{
			"X-Signature": "123456789",
		},
		Logger: suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithCookies() {
	serverURL, _ := url.Parse(suite.Server.URL)
	reader, err := request.Send(&request.Options{
		URL: serverURL,
		Cookies: []*http.Cookie{
			{Name: "Test", Value: "1234", Secure: true, HttpOnly: true},
		},
		Logger: suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithQueryParameters() {
	serverURL, _ := url.Parse(suite.Server.URL)
	reader, err := request.Send(&request.Options{
		URL: serverURL,
		Parameters: map[string]string{
			"page": "25",
		},
		Logger: suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithContentReaderPayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/item")
	data := struct{ ID string }{ID: "1234"}
	payload, _ := json.Marshal(data)
	payloadReader := request.ContentReader{
		Type:   "application/json",
		Length: int64(len(payload)),
		Reader: ioutil.NopCloser(bytes.NewBuffer(payload)),
	}
	reader, err := request.Send(&request.Options{
		URL:     serverURL,
		Payload: payloadReader,
		Logger:  suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1234", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithContentReaderPayloadAndNoLength() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/item")
	data := struct{ ID string }{ID: "1234"}
	payload, _ := json.Marshal(data)
	payloadReader := request.ContentReader{
		Type:   "application/json",
		Reader: ioutil.NopCloser(bytes.NewBuffer(payload)),
	}
	reader, err := request.Send(&request.Options{
		URL:     serverURL,
		Payload: payloadReader,
		Logger:  suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1234", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithContentReaderPayloadAndNoType() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/item")
	data := struct{ ID string }{ID: "1234"}
	payload, _ := json.Marshal(data)
	payloadReader := request.ContentReader{
		Length: int64(len(payload)),
		Reader: ioutil.NopCloser(bytes.NewBuffer(payload)),
	}
	reader, err := request.Send(&request.Options{
		URL:     serverURL,
		Payload: payloadReader,
		Logger:  suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1234", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithContentReaderPointerPayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/item")
	data := struct{ ID string }{ID: "1234"}
	payload, _ := json.Marshal(data)
	payloadReader := &request.ContentReader{
		Type:   "application/json",
		Length: int64(len(payload)),
		Reader: ioutil.NopCloser(bytes.NewBuffer(payload)),
	}
	reader, err := request.Send(&request.Options{
		URL:     serverURL,
		Payload: payloadReader,
		Logger:  suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1234", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithContentReaderPointerPayloadAndNoLength() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/item")
	data := struct{ ID string }{ID: "1234"}
	payload, _ := json.Marshal(data)
	payloadReader := &request.ContentReader{
		Type:   "application/json",
		Reader: ioutil.NopCloser(bytes.NewBuffer(payload)),
	}
	reader, err := request.Send(&request.Options{
		URL:     serverURL,
		Payload: payloadReader,
		Logger:  suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1234", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithContentReaderPointerPayloadAndNoType() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/item")
	data := struct{ ID string }{ID: "1234"}
	payload, _ := json.Marshal(data)
	payloadReader := &request.ContentReader{
		Length: int64(len(payload)),
		Reader: ioutil.NopCloser(bytes.NewBuffer(payload)),
	}
	reader, err := request.Send(&request.Options{
		URL:     serverURL,
		Payload: payloadReader,
		Logger:  suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1234", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithContentPayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/item")
	data := struct{ ID string }{ID: "1234"}
	payload, _ := json.Marshal(data)
	payloadContent := request.ContentWithData(payload, "application/json")
	reader, err := request.Send(&request.Options{
		URL:     serverURL,
		Payload: *payloadContent,
		Logger:  suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1234", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithContentPointerPayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/item")
	data := struct{ ID string }{ID: "1234"}
	payload, _ := json.Marshal(data)
	payloadContent := request.ContentWithData(payload, "application/json")
	reader, err := request.Send(&request.Options{
		URL:     serverURL,
		Payload: payloadContent,
		Logger:  suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1234", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithAttachmentAsPayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/item")
	data := struct{ ID string }{ID: "1234"}
	payload, _ := json.Marshal(data)
	payloadContent := request.ContentWithData(payload, "application/json")
	reader, err := request.Send(&request.Options{
		URL:        serverURL,
		Attachment: payloadContent.Reader(),
		Logger:     suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1234", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithStructPayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/item")
	reader, err := request.Send(&request.Options{
		URL:     serverURL,
		Payload: struct{ ID string }{ID: "1234"},
		Logger:  suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1234", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithStructPayloadAndNoReqLogSize() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/item")
	reader, err := request.Send(&request.Options{
		URL:                serverURL,
		Payload:            struct{ ID string }{ID: "1234"},
		RequestBodyLogSize: -1,
		Logger:             suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1234", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithStringMapPayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/items")
	reader, err := request.Send(&request.Options{
		URL:     serverURL,
		Payload: map[string]string{"ID": "1234"},
		Logger:  suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithMapPayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/items")
	reader, err := request.Send(&request.Options{
		URL:                serverURL,
		Payload:            map[string]stuff{"ID": {"1234"}},
		RequestBodyLogSize: -1,
		Logger:             suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithStringMapPayloadAndAttachment() {
	attachment := request.ContentWithData(smallPNG(), "image/png")
	suite.Require().Equal("image/png", attachment.Type, "Attachment type is wrong")
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/image")
	reader, err := request.Send(&request.Options{
		URL:        serverURL,
		Payload:    map[string]string{"ID": "1234", ">file": "image.png"},
		Attachment: attachment.Reader(),
		Logger:     suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithSlicePayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/items")
	reader, err := request.Send(&request.Options{
		Method: http.MethodDelete,
		URL:    serverURL,
		Payload: []struct{ ID string }{
			{ID: "1234"},
			{ID: "5678"},
		},
		Logger: suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("2", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithSlicePayloadAndNoReqLogSize() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/items")
	reader, err := request.Send(&request.Options{
		Method: http.MethodDelete,
		URL:    serverURL,
		Payload: []struct{ ID string }{
			{ID: "1234"},
			{ID: "5678"},
		},
		RequestBodyLogSize: -1,
		Logger:             suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("2", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithResults() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/results")
	results := struct {
		Code int `json:"code"`
	}{}
	reader, err := request.Send(&request.Options{
		URL:    serverURL,
		Logger: suite.Logger,
	}, &results)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	suite.Assert().Equal(1234, results.Code, "Results should have been received and decoded")
}

func (suite *RequestSuite) TestCanSendRequestWithResultsAndInvalidData() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/")
	results := struct {
		Code int `json:"code"`
	}{}
	reader, err := request.Send(&request.Options{
		URL:    serverURL,
		Logger: suite.Logger,
	}, &results)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	suite.Assert().Equal(0, results.Code, "Results should not have been decoded")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithToken() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/token")
	reader, err := request.Send(&request.Options{
		URL:           serverURL,
		Authorization: "Bearer ThisIsAToken",
		Logger:        suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestShouldFailSendingWithMissingURL() {
	_, err := request.Send(&request.Options{}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Assert().True(errors.Is(err, errors.ArgumentMissing), "error should be an Argument Missing error, error: %+v", err)
	var details errors.Error
	suite.Require().True(errors.As(err, &details), "Error chain should contain an errors.Error")
	suite.Assert().Equal("URL", details.What, "Error's What is wrong")
}

func (suite *RequestSuite) TestShouldFailSendingWithWrongURL() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/these_are_not_the_droids_you_are_looking_for")
	_, err := request.Send(&request.Options{
		URL:    serverURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Assert().True(errors.Is(err, errors.HTTPNotFound), "error should be an HTTP Not Found error, error: %+v", err)
}

func (suite *RequestSuite) TestShouldFailSendingWithInvalidMethod() {
	serverURL, _ := url.Parse(suite.Server.URL)
	_, err := request.Send(&request.Options{
		Method: "HOCUS POCUS",
		URL:    serverURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Assert().Contains(err.Error(), "invalid method")
}

func (suite *RequestSuite) TestShouldFailSendingWithUnsupportedPayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/items")
	_, err := request.Send(&request.Options{
		URL:     serverURL,
		Payload: 1234,
		Logger:  suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Assert().Contains(err.Error(), "Unsupported Payload: int")
}

func (suite *RequestSuite) TestShouldFailSendingWithEmptyAttachment() {
	attachment := request.ContentWithData([]byte{}, "image/png")
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/items")
	_, err := request.Send(&request.Options{
		URL:        serverURL,
		Payload:    map[string]string{"ID": "1234", ">file": "image.png"},
		Attachment: attachment.Reader(),
		Logger:     suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Assert().Contains(err.Error(), "Missing/Empty Attachment")
}

func (suite *RequestSuite) TestShouldFailSendingWithMissingAttachmentName() {
	attachment := request.ContentWithData(smallPNG(), "image/png")
	suite.Require().Equal("image/png", attachment.Type, "Attachment type is wrong")
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/image")
	_, err := request.Send(&request.Options{
		URL:        serverURL,
		Payload:    map[string]string{"ID": "1234", ">file": ""},
		Attachment: attachment.Reader(),
		Logger:     suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Assert().Contains(err.Error(), "Empty value for multipart form field")
}

func (suite *RequestSuite) TestShouldFailSendingWithMissingAttachmentKey() {
	attachment := request.ContentWithData(smallPNG(), "image/png")
	suite.Require().Equal("image/png", attachment.Type, "Attachment type is wrong")
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/image")
	_, err := request.Send(&request.Options{
		URL:        serverURL,
		Payload:    map[string]string{"ID": "1234", ">": "image.png"},
		Attachment: attachment.Reader(),
		Logger:     suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Assert().Contains(err.Error(), "Empty key for multipart form field")
}

func (suite *RequestSuite) TestShouldFailSendingWitBogusAttachmentReader() {
	reader := &request.ContentReader{
		Type:   "image/png",
		Length: int64(67),
		Reader: failingReader(0),
	}
	suite.Require().Equal("image/png", reader.Type, "Attachment type is wrong")
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/image")
	_, err := request.Send(&request.Options{
		URL:        serverURL,
		Payload:    map[string]string{"ID": "1234", ">file": "image.png"},
		Attachment: reader,
		Logger:     suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Assert().Contains(err.Error(), "Failed to write attachment to multipart form field file")
}

func (suite *RequestSuite) TestCanReceiveJPGType() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/bad_jpg_type")
	reader, err := request.Send(&request.Options{
		URL:    serverURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
	suite.Assert().Equal("image/jpeg", reader.Type, "Type was not converted correctly")
	suite.Assert().Equal("image/jpeg", content.Type, "Type was not converted correctly")
}

func (suite *RequestSuite) TestCanReceiveTypeFromAccept() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/data")
	reader, err := request.Send(&request.Options{
		URL:    serverURL,
		Accept: "text/html",
		Logger: suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
	suite.Assert().Equal("text/html", reader.Type, "Type was not converted correctly")
	suite.Assert().Equal("text/html", content.Type, "Type was not converted correctly")
}

func (suite *RequestSuite) TestCanReceiveTypeFromURL() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/audio.mp3")
	reader, err := request.Send(&request.Options{
		URL:    serverURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
	suite.Assert().Equal("audio/mpeg3", reader.Type, "Type was not converted correctly")
	suite.Assert().Equal("audio/mpeg3", content.Type, "Type was not converted correctly")
}

func (suite *RequestSuite) TestCanRetryReceivingRequest() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/retry")
	_, err := request.Send(&request.Options{
		URL:                  serverURL,
		RetryableStatusCodes: []int{http.StatusServiceUnavailable},
		Attempts:             5,
		Logger:               suite.Logger,
		Timeout:              1 * time.Second,
	}, nil)
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
}

func (suite *RequestSuite) TestCanRetryPostingRequest() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/retry")
	_, err := request.Send(&request.Options{
		URL:                  serverURL,
		Payload:              struct {
			ID string `json:"id"`
		}{
			ID: "1234",
		},
		RetryableStatusCodes: []int{http.StatusServiceUnavailable},
		Attempts:             5,
		Logger:               suite.Logger,
		Timeout:              1 * time.Second,
	}, nil)
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
}

func (suite *RequestSuite) TestShouldFailWithBadRedirectLocation() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/bad_redirect")
	_, err := request.Send(&request.Options{
		URL:    serverURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	var details *url.Error
	suite.Require().True(errors.As(err, &details), "Error chain should contain an URL Error")
	suite.Assert().Contains(details.Error(), "response missing Location header")
}

func (suite *RequestSuite) TestShouldFailReceivingWhenTimeoutAnd1Attempt() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/timeout")
	start := time.Now()
	_, err := request.Send(&request.Options{
		URL:      serverURL,
		Attempts: 1,
		Logger:   suite.Logger,
		Timeout:  1 * time.Second,
	}, nil)
	end := time.Since(start)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Infof("Expected error: %s", err.Error())
	suite.Assert().True(errors.Is(err, errors.HTTPStatusRequestTimeout), "error should be an HTTP Request Timeout error, error: %+v", err)
	suite.Assert().LessOrEqual(int64(end), int64(2*time.Second), "The request lasted more than 2 second (%s)", end)
}

func (suite *RequestSuite) TestShouldFailReceivingWhenTimeoutAnd2Attempts() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/timeout")
	start := time.Now()
	_, err := request.Send(&request.Options{
		URL:      serverURL,
		Attempts: 2,
		Logger:   suite.Logger,
		Timeout:  1 * time.Second,
	}, nil)
	end := time.Since(start)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Infof("Expected error: %s", err.Error())
	suite.Assert().True(errors.Is(err, errors.HTTPStatusRequestTimeout), "error should be an HTTP Request Timeout error, error: %+v", err)
	suite.Assert().LessOrEqual(int64(end), int64(4*time.Second), "The request lasted more than 4 second (%s)", end)
}

func (suite *RequestSuite) TestShouldFailPostingWhenTimeoutAnd1Attempt() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/timeout")
	start := time.Now()
	_, err := request.Send(&request.Options{
		URL:      serverURL,
		Payload:  struct{
			ID string `json:"id"`
		}{
			ID: "1",
		},
		Attempts: 1,
		Logger:   suite.Logger,
		Timeout:  1 * time.Second,
	}, nil)
	end := time.Since(start)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Infof("Expected error: %s", err.Error())
	suite.Assert().True(errors.Is(err, errors.HTTPStatusRequestTimeout), "error should be an HTTP Request Timeout error, error: %+v", err)
	suite.Assert().LessOrEqual(int64(end), int64(2*time.Second), "The request lasted more than 2 second (%s)", end)
}

func (suite *RequestSuite) TestShouldFailPostingWhenTimeoutAnd2Attempts() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/timeout")
	start := time.Now()
	_, err := request.Send(&request.Options{
		URL:      serverURL,
		Payload:  struct{
			ID string `json:"id"`
		}{
			ID: "1",
		},
		Attempts: 2,
		Logger:   suite.Logger,
		Timeout:  1 * time.Second,
	}, nil)
	end := time.Since(start)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Infof("Expected error: %s", err.Error())
	suite.Assert().True(errors.Is(err, errors.HTTPStatusRequestTimeout), "error should be an HTTP Request Timeout error, error: %+v", err)
	suite.Assert().LessOrEqual(int64(end), int64(4*time.Second), "The request lasted more than 4 second (%s)", end)
}

func (suite *RequestSuite) TestShouldFailReceivingWithTooManyRetries() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/retry")
	_, err := request.Send(&request.Options{
		URL:                  serverURL,
		RetryableStatusCodes: []int{http.StatusServiceUnavailable},
		Attempts:             2,
		Logger:               suite.Logger,
		Timeout:              1 * time.Second,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Infof("Expected error: %s", err.Error())
	suite.Assert().True(errors.Is(err, errors.HTTPServiceUnavailable), "error should be an HTTP Service Unavailable error, error: %+v", err)
}

func (suite *RequestSuite) TestShouldFailPostingWithTooManyRetries() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/retry")
	_, err := request.Send(&request.Options{
		URL:                  serverURL,
		Payload:              struct{
			ID string `json:"id"`
		}{
				ID: "1",
		},
		RetryableStatusCodes: []int{http.StatusServiceUnavailable},
		Attempts:             2,
		Logger:               suite.Logger,
		Timeout:              1 * time.Second,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Logger.Infof("Expected error: %s", err.Error())
	suite.Assert().True(errors.Is(err, errors.HTTPServiceUnavailable), "error should be an HTTP Service Unavailable error, error: %+v", err)
}

func (suite *RequestSuite) TestShouldFailReceivingBadResponse() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/bad_response")
	_, err := request.Send(&request.Options{
		URL:    serverURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Assert().Contains(err.Error(), "unexpected EOF")
}

func (suite *RequestSuite) TestCanSendPostRequestWithRedirect() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/redirect")
	reader, err := request.Send(&request.Options{
		Method: http.MethodPost,
		URL:    serverURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestCanSendGetRequestWithRedirect() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/redirect")
	reader, err := request.Send(&request.Options{
		URL:    serverURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithContentPayloadAndOneTimeout() {
	requestTimeout := 500 * time.Millisecond
	suite.Logger.Infof("Request Timeout: %s", requestTimeout)
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/item-with-timeout")
	data := struct{ ID string }{ID: "1234"}
	payload, _ := json.Marshal(data)
	payloadContent := request.ContentWithData(payload, "application/json")
	reader, err := request.Send(&request.Options{
		Method:  http.MethodPost,
		URL:     serverURL,
		Payload: *payloadContent,
		Timeout: requestTimeout,
		Logger:  suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("1234", string(content.Data))
	suite.Logger.Infof("Test finished")
}

// Suite Tools

func (suite *RequestSuite) SetupSuite() {
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

	suite.Server = CreateTestServer(suite)
	suite.Proxy = CreateTestProxy(suite)
}

func (suite *RequestSuite) TearDownSuite() {
	suite.Logger.Debugf("Tearing down")
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
	}
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))

	suite.Server.Close()
	suite.Logger.Infof("Closed the Test WEB Server")
	suite.Logger.Close()
}

func (suite *RequestSuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
}

func (suite *RequestSuite) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	if suite.T().Failed() {
		suite.Logger.Errorf("Test %s failed", testName)
	}
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}
