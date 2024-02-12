# go-request

![GoVersion](https://img.shields.io/github/go-mod/go-version/gildas/go-request)
[![GoDoc](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/gildas/go-request)
[![License](https://img.shields.io/github/license/gildas/go-request)](https://github.com/gildas/go-request/blob/master/LICENSE)
[![Report](https://goreportcard.com/badge/github.com/gildas/go-request)](https://goreportcard.com/report/github.com/gildas/go-request)  

![master](https://img.shields.io/badge/branch-master-informational)
[![Test](https://github.com/gildas/go-request/actions/workflows/test.yml/badge.svg?branch=master)](https://github.com/gildas/go-request/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/gildas/go-request/branch/master/graph/badge.svg?token=gFCzS9b7Mu)](https://codecov.io/gh/gildas/go-request/branch/master)

![dev](https://img.shields.io/badge/branch-dev-informational)
[![Test](https://github.com/gildas/go-request/actions/workflows/test.yml/badge.svg?branch=dev)](https://github.com/gildas/go-request/actions/workflows/test.yml)
[![codecov](https://codecov.io/gh/gildas/go-request/branch/dev/graph/badge.svg?token=gFCzS9b7Mu)](https://codecov.io/gh/gildas/go-request/branch/dev)

A Package to send requests to HTTP/REST services.

## Usage

The main func allows to send HTTP request to REST servers and takes care of payloads, JSON, result collection.

Examples:

```go
res, err := request.Send(&request.Options{
    URL: myURL,
}, nil)
if err != nil {
    return err
}
data := struct{Data string}{}
err := res.UnmarshalContentJSON(&data)
```

Here we send an HTTP GET request and unmarshal the response.

It is also possible to let `request.Send` do the unmarshal for us:

```go
data := struct{Data string}{}
_, err := request.Send(&request.Options{
    URL: myURL,
}, &data)
```

In that case, the returned `Content`'s data is an empty byte array. Its other properties are valid (like the size, mime type, etc)

You can also download data directly to an `io.Writer`:

```go
writer, err := os.Create(filepath.Join("tmp", "data"))
defer writer.Close()
res, err := request.Send(&request.Options{
  URL: serverURL,
}, writer)
log.Infof("Downloaded %d bytes", res.Length)
```

In that case, the returned `Content`'s data is an empty byte array. Its other properties are valid (like the size, mime type, etc)

Authorization can be stored in the `Options.Authorization`:

```go
payload := struct{Key string}{}
res, err := request.Send(&request.Options{
    URL:           myURL,
    Authorization: request.BasicAuthorization("user", "password"),
}, nil)
```

or, with a Bearer Token:  

```go
payload := struct{Key string}{}
res, err := request.Send(&request.Options{
    URL:           myURL,
    Authorization: request.BearerAuthorization("myTokenABCD"),
}, nil)
```

Objects can be sent as payloads:

```go
payload := struct{Key string}{}
res, err := request.Send(&request.Options{
    URL:     myURL,
    Payload: payload,
}, nil)
```

A payload will induce an HTTP POST unless mentioned.

So, to send an `HTTP PUT`, simply write:

```go
payload := struct{Key string}{}
res, err := request.Send(&request.Options{
    Method:  http.MethodPut,
    URL:     myURL,
    Payload: payload,
}, nil)
```

To send an x-www-form, use a `map` in the payload:  

```go
res, err := request.Send(&request.Options{
    Method:  http.MethodPut,
    URL:     myURL,
    Payload: map[string]string{
        "ID":   "1234",
        "Kind": "stuff,"
    },
}, nil)
```

To send a multipart form with an attachment, use a `map`, an attachment, and one of the key must start with `>`:  

```go
attachment := io.Open("/path/to/file")
res, err := request.Send(&request.Options{
    Method:  http.MethodPut,
    URL:     myURL,
    Payload: map[string]string{
        "ID":    "1234",
        "Kind":  "stuff,"
        ">file": "image.png",
    },
    Attachment: attachment,
}, nil)
```

The file name and its key will be written in the `multipart/form-data`'s `Content-Disposition` header as: `form-data; name="file"; filename="image.png"`.

To send the request again when receiving a Service Unavailable (`Attempts` and `Timeout` are optional):  

```go
res, err := request.Send(&request.Options{
    URL:                  myURL,
    RetryableStatusCodes: []int{http.StatusServiceUnavailable},
    Attempts:             10,
    Timeout:              2 * time.Second,
}, nil)
```

By default, upon receiving a retryable status code, `Send` will use am exponential backoff algorithm to retry the request. By default, it will wait for 3 seconds before retrying for 5 minutes, then 9 seconds between 5 and 10 minutes, then 27 seconds between 10 and 15 minutes, etc.

You can change the delay and the backoff factor like this:

```go
res, err := request.Send(&request.Options{
    URL:                         myURL,
    InterAttemptDelay:           5 * time.Second,
    InterAttemptBackoffInterval: 2 * time.Minute,
}, nil)
```

You can also not use the backoff algorithm and use the `Retry-After` header instead:

```go
res, err := request.Send(&request.Options{
    URL:                       myURL,
    // ...
    InterAttemptUseRetryAfter: true,
}, nil)
```

When sending requests to upload data streams, you can provide an `io.Writer` to write the progress to:

```go
import "github.com/schollz/progressbar/v3"

reader, err := os.Open(pathToFile)
defer reader.Close()
stat, err := reader.Stat()
bar := progressbar.DefaultBytes(stat.Size(), "Uploading")
res, err := request.Send(&request.Options{
  Method:         http.MethodPost,
  URL:            serverURL,
  Payload:        reader,
  ProgressWriter: bar,
}, reader)
```

If the progress `io.Writer` is also an `io.Closer`, it will be closed at the end of the `request.Send()`.

When sending requests to download data streams, you can provide an `io.Writer` to write the progress to:

```go
import "github.com/schollz/progressbar/v3"

writer, err := os.Create(pathToFile)
defer writer.Close()
bar := progressbar.DefaultBytes(-1, "Downloading") // will use a spinner
res, err := request.Send(&request.Options{
  URL:            serverURL,
  ProgressWriter: bar,
}, writer)
```

Again, if the progress `io.Writer` is also an `io.Closer`, it will be closed at the end of the `request.Send()`.

if you provide a `request.Options.ProgressSetMax` func or if the `io.Writer` is a `request.ProgressBarMaxSetter` or a `request.ProgressBarMaxChanger`, `request.Send` will call it to set the maximum value of the progress bar from the response `Content-Length`:

```go
import "github.com/schollz/progressbar/v3"

writer, err := os.Create(pathToFile)
defer writer.Close()
bar := progressbar.DefaultBytes(1, "Downloading") // use a temporary max value
res, err := request.Send(&request.Options{
  URL:            serverURL,
  ProgressWriter: bar,
}, writer)
```

```go
import "github.com/cheggaaa/pb/v3"

writer, err := os.Create(pathToFile)
defer writer.Close()
bar := pb.StartNew(1) // use a temporary max value
res, err := request.Send(&request.Options{
  URL:                serverURL,
  ProgressWriter:     bar,
  ProgressSetMaxFunc: func(max int64) { bar.SetTotal64(max) },
}, writer)
```

**Notes:**  

- if the PayloadType is not mentioned, it is calculated when processing the Payload.
- if the payload is a `ContentReader` or a `Content`, it is used directly.
- if the payload is a `map[string]xxx` where *xxx* is not `string`, the `fmt.Stringer` is used whenever possible to get the string version of the values.
- if the payload is a struct or a pointer to struct, the body is sent as `application/json` and marshaled.
- if the payload is an array or a slice, the body is sent as `application/json` and marshaled.
- The option `Logger` can be used to let the `request` library log to a `gildas/go-logger`. By default, it logs to a `NilStream` (see github.com/gildas/go-logger).
- When using a logger, you can control how much of the Request/Response Body is logged with the options `RequestBodyLogSize`/`ResponseBodyLogSize`. By default they are set to 2048 bytes. If you do not want to log them, set the options to *-1*.
- `Send()` makes 5 attempts by default to reach the given URL. If option `RetryableStatusCodes` is given, it will attempt the request again when it receives an HTTP Status Code in the given list. If it is not given, the default list is `[]int{http.StatusServiceUnavailable, http.StatusGatewayTimeout, http.StatusBadGateway, http.StatusRequestTimeout, http.StatusTooManyRequests}`.
- The default timeout for `Send()` is 1 second.

**TODO:**  

- Support other kinds of `map` in the payload, like `map[string]int`, etc.
- Maybe have an interface for the Payload to allow users to provide the logic of building the payload themselves. (`type PayloadBuilder interface { BuildPayload() *ContentReader}`?!?)
