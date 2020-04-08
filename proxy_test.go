package request_test

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
)

func CreateTestProxy(suite *RequestSuite) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
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
