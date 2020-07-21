package request_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gildas/go-core"
	"github.com/gildas/go-request"
)

func CreateTestServer(suite *RequestSuite) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		log := suite.Logger.Child("server", "handler")
		headers := map[string]string{}
		for key, values := range req.Header {
			headers[key] = strings.Join(values, ", ")
		}
		log.Record("headers", headers).Infof("Request: %s %s", req.Method, req.URL)

		switch req.Method {
		case http.MethodPost:
			switch req.URL.Path {
			case "/items":
				items := []stuff{}

				if req.Header.Get("Content-type") == "application/json" {
					reqContent, err := request.ContentFromReader(req.Body, req.Header.Get("Content-Type"))
					if err != nil {
						log.Errorf("Failed to read request content", err)
						core.RespondWithError(res, http.StatusBadRequest, err)
						return
					}
					log.Infof("Request body: %s, %d bytes: \n%s", reqContent.Type, reqContent.Length, string(reqContent.Data))
					if err = json.Unmarshal(reqContent.Data, &items); err != nil {
						log.Errorf("Failed to read request content", err)
						core.RespondWithError(res, http.StatusBadRequest, err)
						return
					}
					if len(items) < 1 {
						log.Errorf(("Not enough items to add"))
						core.RespondWithError(res, http.StatusBadRequest, err)
						return
					}
					log.Infof("Adding #%d items", len(items))
				} else if req.Header.Get("Content-type") == "application/x-www-form-urlencoded" {
					err := req.ParseForm()
					if err != nil {
						log.Errorf("Failed to parse request as a form", err)
						core.RespondWithError(res, http.StatusBadRequest, err)
						return
					}
					log.Infof("POST Form: %+#v", req.PostForm)
					for key, values := range req.PostForm {
						log.Infof("Form: key=%s, values=%+#v", key, values)
					}
					id := req.PostFormValue("ID")
					if len(id) > 0 {
						item := stuff{id}
						log.Infof("Adding %#+v", item)
						items = append(items, item)

					}
				}

				if _, err := res.Write([]byte(fmt.Sprintf("%d", len(items)))); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			case "/item":
				reqContent, err := request.ContentFromReader(req.Body, req.Header.Get("Content-Type"))
				if err != nil {
					log.Errorf("Failed to read request content", err)
					core.RespondWithError(res, http.StatusBadRequest, err)
					return
				}
				log.Infof("Request body: %s, %d bytes: \n%s", reqContent.Type, reqContent.Length, string(reqContent.Data))
				if reqContent.Length == 0 {
					log.Errorf("Content is empty")
					core.RespondWithError(res, http.StatusBadRequest, err)
					return
				}
				item := struct{ ID string }{}
				if err = json.Unmarshal(reqContent.Data, &item); err != nil {
					log.Errorf("Failed to read request content", err)
					core.RespondWithError(res, http.StatusBadRequest, err)
					return
				}
				if _, err := res.Write([]byte(item.ID)); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			case "/image":
				items := []stuff{}
				if strings.HasPrefix(req.Header.Get("Content-Type"), "multipart/form-data") {
					err := req.ParseMultipartForm(int64(1024))
					if err != nil {
						log.Errorf("Failed to parse request as a multipart form", err)
						core.RespondWithError(res, http.StatusBadRequest, err)
						return
					}
					log.Infof("POST Form: %+#v", req.PostForm)
					for key, values := range req.PostForm {
						log.Infof("Form: key=%s, values=%+#v", key, values)
					}
					id := req.PostFormValue("ID")
					if len(id) > 0 {
						item := stuff{id}
						log.Infof("Adding %#+v", item)
						items = append(items, item)

					}
					_, fileHeader, err := req.FormFile("file")
					if err != nil {
						log.Errorf("Failed to read file from multipart form", err)
						core.RespondWithError(res, http.StatusBadRequest, err)
						return
					}
					log.Infof("File %s: mime=%s, size=%d", fileHeader.Filename, fileHeader.Header.Get("Content-Type"), fileHeader.Size)
					log.Debugf("File header=%s", fileHeader.Header)
					if fileHeader.Header.Get("Content-Type") != "image/png" {
						log.Errorf("Attachment is not a PNG image")
						core.RespondWithError(res, http.StatusBadRequest, err)
						return
					}
					if fileHeader.Size == 0 {
						log.Errorf("Attachment is empty")
						core.RespondWithError(res, http.StatusBadRequest, err)
						return
					}
				}
				if _, err := res.Write([]byte(fmt.Sprintf("%d", len(items)))); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
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
			case "/retry":
				attempt := req.Header.Get("X-Attempt")
				if attempt != "5" { // On the 5th attempt, we want to return 200
					res.WriteHeader(http.StatusServiceUnavailable)
					return
				}
			case "/redirect":
				res.Header().Add("Location", "/")
				res.WriteHeader(http.StatusFound)
				log.Infof("Redirecting to /")
			case "/bad_redirect":
				res.Header().Add("Location", "") // This is on purpose to check if the client handles this error well
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
		case http.MethodDelete:
			defer req.Body.Close()
			reqContent, err := request.ContentFromReader(req.Body, req.Header.Get("Content-Type"))
			if err != nil {
				log.Errorf("Failed to read request content", err)
				core.RespondWithError(res, http.StatusBadRequest, err)
				return
			}
			log.Infof("Request body: %s, %d bytes: \n%s", reqContent.Type, reqContent.Length, string(reqContent.Data))
			switch req.URL.Path {
			case "/items":
				items := []struct{ ID string }{}
				if err = json.Unmarshal(reqContent.Data, &items); err != nil {
					log.Errorf("Failed to read request content", err)
					core.RespondWithError(res, http.StatusBadRequest, err)
					return
				}
				log.Infof("Deleting #%d items", len(items))
				if _, err := res.Write([]byte(fmt.Sprintf("%d", len(items)))); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
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

}
