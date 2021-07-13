package request_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"time"

	"github.com/gildas/go-core"
	"github.com/gildas/go-errors"
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
			log.Debugf("Checking Path: %s", req.URL.Path)
			switch strings.ToLower(req.URL.Path) {
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
						core.RespondWithError(res, http.StatusBadRequest, errors.ArgumentMissing.With("items").WithStack())
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
					core.RespondWithError(res, http.StatusBadRequest, errors.ArgumentMissing.With("body").WithStack())
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
			case "/item-with-timeout":
				_log := log.Record("attempt", req.Header.Get("X-Attempt"))
				attempt, err := strconv.Atoi(req.Header.Get("X-Attempt"))
				if err != nil {
					_log.Errorf("Request Header X-Attempt does not contain a valid number", err)
					attempt = 0
				}
				if attempt <= 1 { // in that case, we should timeout
					responseTimeout := 600 * time.Millisecond
					_log.Infof("Path: %s, first attempt, timing out (%s)", req.URL.Path, responseTimeout)
					time.Sleep(responseTimeout)
					_log.Infof("Path: %s, waited long enough", req.URL.Path)
					return
				}
				_log.Infof("Path: %s, attempt %d, processing expecting %s bytes", req.URL.Path, attempt, req.Header.Get("Content-Length"))
				reqContent, err := request.ContentFromReader(req.Body, req.Header.Get("Content-Type"))
				if err != nil {
					_log.Errorf("Failed to read request content", err)
					core.RespondWithError(res, http.StatusBadRequest, err)
					return
				}
				_log.Infof("Request body: %s, %d bytes: \n%s", reqContent.Type, reqContent.Length, string(reqContent.Data))
				if reqContent.Length == 0 {
					_log.Errorf("Content is empty")
					core.RespondWithError(res, http.StatusBadRequest, errors.ArgumentMissing.With("body").WithStack())
					return
				}
				item := struct{ ID string }{}
				if err = json.Unmarshal(reqContent.Data, &item); err != nil {
					_log.Errorf("Failed to read request content", err)
					core.RespondWithError(res, http.StatusBadRequest, err)
					return
				}
				if _, err := res.Write([]byte(item.ID)); err != nil {
					_log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
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
						core.RespondWithError(res, http.StatusBadRequest, errors.ArgumentInvalid.With("Content-Type", fileHeader.Header.Get("Content-Type")).WithStack())
						return
					}
					if fileHeader.Size == 0 {
						log.Errorf("Attachment is empty")
						core.RespondWithError(res, http.StatusBadRequest, errors.ArgumentMissing.With("body").WithStack())
						return
					}
				}
				if _, err := res.Write([]byte(fmt.Sprintf("%d", len(items)))); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			default:
				if _, err := res.Write([]byte("body")); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			}
		case http.MethodGet:
			log.Debugf("Checking Path: %s, raw: %s", req.URL.Path, req.URL.EscapedPath())
			switch strings.ToLower(req.URL.Path) {
			case "/":
				if _, err := res.Write([]byte("body")); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			case "/audio.mp3":
				res.Header().Add("Content-Type", "application/octet-stream")
				if _, err := res.Write([]byte(`body`)); err != nil {
					log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
				}
			// case "/bo%C3%AEte.png":
			case "/boîte.png":
				signature := req.URL.Query().Get("X-Amz-Signature")
				if signature != "853f1611536e57902b0fcabd36e7fbe77fe2278f40a0aeed4116953ea6ef4873" {
					log.Errorf("Request is missing signature in the query")
					res.Header().Add("Content-Type", "application/json")
					res.WriteHeader(http.StatusForbidden)
					if _, err := res.Write([]byte(`{"error": "error.signature.missing"}`)); err != nil {
						log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
					}
					return
				}
				/*
				id := req.Header.Get("X-Amz-Cf-Id")
				if id != "1233453567abcdef" {
					log.Errorf("Request is missing Header X-Amz-Cf-Id")
					res.Header().Add("Content-Type", "application/json")
					res.WriteHeader(http.StatusForbidden)
					if _, err := res.Write([]byte(`{"error": "error.id.missing"}`)); err != nil {
						log.Errorf("Failed to Write response to %s %s, error: %s", req.Method, req.URL, err)
					}
					return
				}
				*/
				res.Header().Add("Content-Type", "image/png")
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
				queryString := "response-content-disposition=attachment%3Bfilename%3D%22Bo%C3%AEte-Vitamine-C.jpg%22&X-Amz-Security-Token=IQoJb3JpZ2luX2VjEAQaDmFwLW5vcnRoZWFzdC0xIkYwRAIgXwjLAgEAHaXF5ADwxHrr%2BGzy2g5P69h5y5e68hHwjzECIFGV5tHl%2FS8VSIVjsWkxwE5Ks9QlkmM2MIVnB5XD5OvUKo0ECO3%2F%2F%2F%2F%2F%2F%2F%2F%2F%2FwEQABoMNzY1NjI4OTg1NDcxIgz%2Bjt73UFcVYMCGvYQq4QO02iQHzW%2FH%2Fs8El%2BTPF0ekwyOpSN0rxQ6482INGSqGiW%2FELVFRayVTR9T1nbuZ74FVuDe2VYWHYOYPyfmaZiKAnh0cR1kQdSE2A6SUhhkG%2BW0KF1Sw0O5J33f43fvhQYqcj%2FIAMTUB8FuVAN03hNYPQck83F%2FjuGYepPJ7AZGHix%2FYtUAB18Wq%2B3idKOS1abya0wV5pS9PSYK2hnt2pDMu82U2rjhNciQpAwBYIt%2FgGmQ1KQ3YGa8hpp5%2BBMC%2FDHletUEo257cAhZzwMOO3uyhK%2FVC1%2Fc3vthmA9EuWXpnbMXXGykZh7Ya26ookMSRXj10Fsz%2BGSe%2Fyan4vePUuwvTS6aozvL7KxoSs6wxD8pAzQRKkn1lf0i6Xzip461xAy8X0YowwxGE8XPzGcLztxgi7L5ef7NI3IzMphWkwH4QMiBD4D7ptGoE16j2zflmXkXNBPH9IBA1KJ%2Bqv1g6Olmrav19oWMDmVoQWD8%2FW%2BYBEobNOAPUJ6hfppxpinuAPMf1uIueFQrgwY58y3vy6WuMQmvjaIIu2u2QqSqH%2BK3SA3AIzcrEmoEub6OxO5Kge8LroqKr18LK2MNj4ZGzgPRrG%2F3aEIT7y0OHqBMF4TUCBFpwPwpuQz6f9yjKhqlZxypBMJ%2BDtocGOqYBoNa1X%2B6yOM2qg7T%2BtP8UNOFw4vZN73svJWhiqXe9lQhdwPvfRFIhrgIdlbvym3eSBD4C%2FyxtxWr9E4lxyFPfW3a6fV%2B7kl1aPjaE85LgVBJ2EGQ4bm1jNwQ1WGIQo%2FYlKBRC5f%2BekV64pv0ol6BQRwOF%2B9mXHRPigu64LVRTgRQBhgrneLEa00GS%2Bur3Nt1yPmTVWvMIoO0%2FBNM0VbBXsbtePl5ozQ%3D%3D&X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Date=20210713T123747Z&X-Amz-SignedHeaders=host&X-Amz-Expires=3600&X-Amz-Credential=ASIA3EQYLGB7Q6I4XLIG%2F20210713%2Fap-northeast-1%2Fs3%2Faws4_request&X-Amz-Signature=853f1611536e57902b0fcabd36e7fbe77fe2278f40a0aeed4116953ea6ef4873"
				res.Header().Add("Cache-Control", "no-cache")
				res.Header().Add("Cache-Control", "no-store")
				res.Header().Add("Cache-Control", "must-revalidate")
				res.Header().Add("Connection", "keep-alive")
				res.Header().Add("Content-Length", "0")
				res.Header().Add("Expires", "0")
				res.Header().Add("Strict-Transport-Security", "max-age=600; includeSubDomains")
				res.Header().Add("Location", "/Bo%C3%AEte.png?" + queryString)
				res.Header().Add("Via", "1.1 d4ecead8ac7dbeef7cdfe1233455668f.cloudfront.net (CloudFront)")
				res.Header().Add("X-Amz-Cf-Id", "1233453567abcdef")
				res.Header().Add("X-Amz-Cf-Pop", "NRT20-C2")
				res.Header().Add("X-Cache", "Miss form cloudfront")
				res.WriteHeader(http.StatusSeeOther)
				log.Infof("Redirecting to /Bo%%C3%%AEte.png")
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
