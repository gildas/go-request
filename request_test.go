package request_test

import (
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
	"github.com/stretchr/testify/suite"
)

type RequestSuite struct {
	suite.Suite
	Name   string
	Server *httptest.Server
	Proxy  *httptest.Server
	Logger *logger.Logger
}

func TestRequestSuite(t *testing.T) {
	suite.Run(t, new(RequestSuite))
}

func (suite *RequestSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
	suite.Logger = logger.Create("test", &logger.FileStream{Path: "./test-request.log", Unbuffered: true}).Child("test", "test")
	suite.Server = CreateTestServer(suite)
	suite.Proxy = CreateTestProxy(suite)
}

func (suite *RequestSuite) TearDownSuite() {
	suite.Server.Close()
	suite.Logger.Close()
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

func (suite *RequestSuite) TestCanSendRequestWithStructPayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
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
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithStructPayloadAndNoReqLogSize() {
	serverURL, _ := url.Parse(suite.Server.URL)
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
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithStringMapPayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
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
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestCanSendRequestWithStringerMapPayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
	reader, err := request.Send(&request.Options{
		URL:                serverURL,
		Payload:            map[string]stuff{"ID": stuff{"1234"}},
		RequestBodyLogSize: -1,
		Logger:             suite.Logger,
	}, nil)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	content, err := reader.ReadContent()
	suite.Require().Nil(err, "Failed reading response content, err=%+v", err)
	suite.Require().NotNil(content, "Content should not be nil")
	suite.Assert().Equal("body", string(content.Data))
}

func (suite *RequestSuite) TestShouldFailSendingWithUnsupportedMapPayload() {
	serverURL, _ := url.Parse(suite.Server.URL)
	_, err := request.Send(&request.Options{
		URL:                serverURL,
		Payload:            map[string]int{"ID": 1234},
		RequestBodyLogSize: -1,
		Logger:             suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Assert().True(errors.Is(err, errors.ArgumentInvalidError), "error should be an Argument Invalid error, error: %+v", err)
	var details *errors.Error
	suite.Require().True(errors.As(err, &details), "Error chain should contain an errors.Error")
	suite.Assert().Equal("Payload Type", details.What, "Error's What is wrong")
	suite.Assert().Equal("map[string]int", details.Value.(string), "Error's What is wrong")
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
	suite.Assert().True(errors.Is(err, errors.ArgumentMissingError), "error should be an Argument Missing error, error: %+v", err)
	var details *errors.Error
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
	suite.Assert().True(errors.Is(err, errors.HTTPNotFoundError), "error should be an HTTP Not Found error, error: %+v", err)
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
	suite.Assert().True(errors.Is(err, errors.HTTPStatusRequestTimeoutError), "error should be an HTTP Request Timeout error, error: %+v", err)
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
	suite.Assert().True(errors.Is(err, errors.HTTPStatusRequestTimeoutError), "error should be an HTTP Request Timeout error, error: %+v", err)
	suite.Assert().LessOrEqual(int64(end), int64(4*time.Second), "The request lasted more than 4 second (%s)", end)
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
