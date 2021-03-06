# go-request

![GoVersion](https://img.shields.io/github/go-mod/go-version/gildas/go-request)
[![GoDoc](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/gildas/go-request) 
[![License](https://img.shields.io/github/license/gildas/go-request)](https://github.com/gildas/go-request/blob/master/LICENSE) 
[![Report](https://goreportcard.com/badge/github.com/gildas/go-request)](https://goreportcard.com/report/github.com/gildas/go-request)  

A Package to send requests to HTTP/REST services.

|  |   |   |
---|---|---|
master | [![Build Status](https://github.com/gildas/go-request/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/gildas/go-request/actions) | [![coverage](https://codecov.io/gh/gildas/go-request/branch/master/graph/badge.svg)](https://codecov.io/gh/gildas/go-request)  
dev | [![Build Status](https://github.com/gildas/go-request/actions/workflows/go.yml/badge.svg?branch=dev)](https://github.com/gildas/go-request/actions) | [![coverage](https://codecov.io/gh/gildas/go-request/branch/dev/graph/badge.svg)](https://codecov.io/gh/gildas/go-request)  

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
Here we send an HTTP GET request and unmarshal the response (a `ContentReader`).

It is also possible to let `request.Send` do the unmarshal for us:

```go
data := struct{Data string}{}
_, err := request.Send(&request.Options{
    URL: myURL,
}, &data)
if err != nil {
    return err
}
```

Authorization can be stored in the `Options.Authorization`:

```go
payload := struct{Key string}{}
data := struct{Data string}{}
_, err := request.Send(&request.Options{
    URL:           myURL,
    Authorization: request.BasicAuthorization("user", "password"),
}, &data)
if err != nil {
    return err
}
```

or, with a Bearer Token:  

```go
payload := struct{Key string}{}
data := struct{Data string}{}
_, err := request.Send(&request.Options{
    URL:           myURL,
    Authorization: request.BearerAuthorization("myTokenABCD"),
}, &data)
if err != nil {
    return err
}
```

Objects can be sent as payloads:

```go
payload := struct{Key string}{}
data := struct{Data string}{}
_, err := request.Send(&request.Options{
    URL:     myURL,
    Payload: payload,
}, &data)
if err != nil {
    return err
}
```

A payload will induce an HTTP POST unless mentioned.

So, to send an `HTTP PUT`, simply write:

```go
payload := struct{Key string}{}
data := struct{Data string}{}
_, err := request.Send(&request.Options{
    Method:  http.MethodPut,
    URL:     myURL,
    Payload: payload,
}, &data)
if err != nil {
    return err
}
```

To send an x-www-form, use a `map` in the payload:  

```go
data := struct{Data string}{}
_, err := request.Send(&request.Options{
    Method:  http.MethodPut,
    URL:     myURL,
    Payload: map[string]string{
        "ID":   "1234",
        "Kind": "stuff,"
    },
}, &data)
if err != nil {
    return err
}
```

To send a multipart form with an attachment, use a `map`, an attachment, and one of the key must start with `>`:  

```go
attachment := request.ContentWithData(myReadFile(), "image/png")
data := struct{Data string}{}
_, err := request.Send(&request.Options{
    Method:  http.MethodPut,
    URL:     myURL,
    Payload: map[string]string{
        "ID":    "1234",
        "Kind":  "stuff,"
        ">file": "image.png",
    },
    Attachment: attachment.Reader(),
}, &data)
if err != nil {
    return err
}
```
To send the request again when receiving a Service Unavailable (`Attempts` and `Timeout` are optional):  
```go
_, err := request.Send(&request.Options{
    URL:                  myURL,
    RetryableStatusCodes: []int{http.StatusServiceUnavailable},
    Attempts:             10,
    Timeout:              2 * time.Second,
}, nil)
if err != nil {
    return err
}
```

**Notes:**  
- if the PayloadType is not mentioned, it is calculated when processing the Payload.
- if the payload is a `ContentReader` or a `Content`, it is used directly.
- if the payload is a `map[string]xxx` where *xxx* is not `string`, the `fmt.Stringer` is used whenever possible to get the string version of the values.
- if the payload is a struct or a pointer to struct, the body is sent as `application/json` and marshaled.
- if the payload is an array or a slice, the body is sent as `application/json` and marshaled.
- The option `Logger` can be used to let the `request` library log to a `gildas/go-logger`. By default, it logs to a `NilStream` (see github.com/gildas/go-logger).
- When using a logger, you can control how much of the Request/Response Body is logged with the options `RequestBodyLogSize`/`ResponseBodyLogSize`. By default they are set to 2048 bytes. If you do not want to log them, set the options to *-1*.
- `Send()` makes 5 attempts by default to reach the given URL. If option `RetryableStatusCodes` is given, it will attempt the request again when it receives an HTTP Status Code in the given list.
- The default timeout for `Send()` is 1 second.

**TODO**  
- Support other kinds of `map` in the payload, like `map[string]int`, etc.
- Maybe have an interface for the Payload to allow users to provide the logic of building the payload themselves. (`type PayloadBuilder interface { BuildPayload() *ContentReader}`?!?)