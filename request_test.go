package request_test

import (
	"net/http"
	"net"
	"io"
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
	suite.Server = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		log := suite.Logger.Child("server", "handler")
		headers := map[string]string{}
		for key, values := range req.Header {
			headers[key] = strings.Join(values, ", ")
		}
		log.Record("headers", headers).Infof("Request: %s %s", req.Method, req.URL)

		switch req.Method {
		case http.MethodPost:
			switch req.URL.Path {
			case "/redirect":
				res.Header().Add("Location", "/")
				res.WriteHeader(http.StatusSeeOther)
				// res.WriteHeader(http.StatusFound)
				log.Infof("Redirecting to /")
			default:
				if _, err := res.Write([]byte("body")); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			}
		case http.MethodGet:
			switch req.URL.Path {
			case "/":
				if _, err := res.Write([]byte("body")); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			case "/audio.mp3":
				res.Header().Add("Content-Type", "application/octet-stream")
				if _, err := res.Write([]byte(`body`)); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			case "/bad_jpg_type":
				res.Header().Add("Content-Type", "image/jpg")
				if _, err := res.Write([]byte(`body`)); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			case "/bad_response":
				res.Header().Add("Content-Length", "1")
				if _, err := res.Write([]byte(``)); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			case "/data":
				res.Header().Add("Content-Type", "application/octet-stream")
				if _, err := res.Write([]byte(`body`)); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			case "/token":
				auth := req.Header.Get("Authorization")
				if strings.Compare(auth, "Bearer ThisIsAToken") != 0 {
					res.WriteHeader(http.StatusUnauthorized)
					return
				}
				if _, err := res.Write([]byte("body")); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			case "/redirect":
				res.Header().Add("Location", "/")
				res.WriteHeader(http.StatusFound)
				log.Infof("Redirecting to /")
			case "/results":
				if _, err := res.Write([]byte(`{"code": 1234}`)); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			case "/timeout":
				time.Sleep(5 * time.Second)
			default:
				res.WriteHeader(http.StatusNotFound)
				if _, err := res.Write([]byte("{}")); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
				return
			}
		default:
			res.WriteHeader(http.StatusMethodNotAllowed)
			if _, err := res.Write([]byte("{}")); err != nil {
				log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
			}
			return
		}
	}))
	suite.Proxy = httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		log := suite.Logger.Child("proxy", "handler")
		headers := map[string]string{}
		for key, values := range req.Header {
			headers[key] = strings.Join(values, ", ")
		}
		log.Record("headers", headers).Infof("Request: %s %s", req.Method, req.URL)

		log.Infof("Proxying to %s", req.URL)
		client := &http.Client{}
		req.RequestURI = ""
		if remoteIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
			log.Infof("Proxying from %s", remoteIP)
			req.Header.Set("X-Forwarded-For", remoteIP)
		}
		proxyRes, err := client.Do(req)
		if err != nil {
			log.Errorf("Failed to proxy", err)
			http.Error(res, "Proxy Error", http.StatusBadGateway)
		}
		defer proxyRes.Body.Close()
		for key, values := range proxyRes.Header {
			for _, value := range values {
				res.Header().Add(key, value)
			}
		}
		log.Infof("Replying to client (HTTP %s)", proxyRes.Status)
		res.WriteHeader(proxyRes.StatusCode)
		_, _ = io.Copy(res, proxyRes.Body)
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

func (suite *RequestSuite) TestCanSendRequestWithProxy() {
	serverURL, _ := url.Parse(suite.Server.URL)
	proxyURL, _  := url.Parse(suite.Proxy.URL)
	reader, err := request.Send(&request.Options{
		URL:    serverURL,
		Proxy:  proxyURL,
		Attempts: 1,
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
		URL:      serverURL,
		Logger:   suite.Logger,
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
