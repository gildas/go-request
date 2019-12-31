package request_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-logger"
	"github.com/gildas/go-request"
	"github.com/stretchr/testify/suite"
)

type RequestSuite struct {
	suite.Suite
	Name   string
	Server *httptest.Server
	Logger *logger.Logger
}

func TestRequestSuite(t *testing.T) {
	suite.Run(t, new(RequestSuite))
}

func (suite *RequestSuite) SetupSuite() {
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
	suite.Logger = logger.Create("test", &logger.FileStream{Path: "./test-request.log", Unbuffered: true})
	suite.Server = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		suite.Logger.Infof("Request: %s %s", req.Method, req.URL)
		switch req.Method {
		case http.MethodPost:
			if _, err := res.Write([]byte("body")); err != nil {
				suite.Logger.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
			}
		case http.MethodGet:
			switch req.URL.Path {
			case "/":
				if _, err := res.Write([]byte("body")); err != nil {
					suite.Logger.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			case "/results":
				if _, err := res.Write([]byte(`{"code": 1234}`)); err != nil {
					suite.Logger.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			default:
				res.WriteHeader(http.StatusNotFound)
				if _, err := res.Write([]byte("{}")); err != nil {
					suite.Logger.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
				return
			}
		case http.MethodDelete:
		default:
			res.WriteHeader(http.StatusNotFound)
			if _, err := res.Write([]byte("{}")); err != nil {
				suite.Logger.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
			}
			return
		}
	}))
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

func (suite *RequestSuite) TestCanSendRequestWithPayload() {
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

func (suite *RequestSuite) TestCanSendRequestWithResults() {
	serverURL, _ := url.Parse(suite.Server.URL)
	serverURL, _ = serverURL.Parse("/results")
	results := struct {
		Code int `json:"code"`
	}{}
	reader, err := request.Send(&request.Options{
		URL:     serverURL,
		Logger:  suite.Logger,
	}, &results)
	suite.Require().Nil(err, "Failed sending request, err=%+v", err)
	suite.Require().NotNil(reader, "Content Reader should not be nil")
	suite.Assert().Equal(1234, results.Code, "Results should have been received and decoded")
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
		URL: serverURL,
		Logger: suite.Logger,
	}, nil)
	suite.Require().NotNil(err, "Should have failed sending request")
	suite.Assert().True(errors.Is(err, errors.HTTPNotFoundError), "error should be an HTTP Not Found error, error: %+v", err)
}
