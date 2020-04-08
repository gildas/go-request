# go-request

![GoVersion](https://img.shields.io/github/go-mod/go-version/gildas/go-request)
[![GoDoc](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/gildas/go-request) 
[![License](https://img.shields.io/github/license/gildas/go-request)](https://github.com/gildas/go-request/blob/master/LICENSE) 
[![Report](https://goreportcard.com/badge/github.com/gildas/go-request)](https://goreportcard.com/report/github.com/gildas/go-request)  

A Package to send requests to HTTP/REST services.

|  |   |   |   |
---|---|---|---|
master | [![Build Status](https://dev.azure.com/keltiek/gildas/_apis/build/status/gildas.go-request?branchName=master)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=3&branchName=master) | [![Tests](https://img.shields.io/azure-devops/tests/keltiek/gildas/3/master)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=3&branchName=master) | [![coverage](https://img.shields.io/azure-devops/coverage/keltiek/gildas/3/master)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=3&branchName=master&view=codecoverage-tab)  
dev | [![Build Status](https://dev.azure.com/keltiek/gildas/_apis/build/status/gildas.go-request?branchName=dev)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=3&branchName=dev) | [![Tests](https://img.shields.io/azure-devops/tests/keltiek/gildas/3/dev)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=3&branchName=dev) | [![coverage](https://img.shields.io/azure-devops/coverage/keltiek/gildas/3/dev)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=3&branchName=dev&view=codecoverage-tab)  

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

**Notes:**  
- if the PayloadType is not mentioned, it is calculated when processing the Payload.
- if the payload is a `ContentReader` or a `Content`, it is used directly.
- if the payload is a `map[string]xxx` where *xxx* is not `string`, the `fmt.Stringer` is used whenever possible to get the string version of the values.
- if the payload is a struct or a pointer to struct, the body is sent as `application/json` and marshaled.
- if the payload is an array or a slice, the body is sent as `application/json` and marshaled.
- The option `Logger` can be used to let the `request` library log to a `gildas/go-logger`. By default, it logs to a `NilStream` (see github.com/gildas/go-logger).
- When using a logger, you can control how much of the Request/Response Body is logged with the options `RequestBodyLogSize`/`ResponseBodyLogSize`. By default they are set to 2048 bytes. If you do not want to log them, set the options to *-1*.

**TODO**  
- Support other kinds of `map` in the payload, like `map[string]int`, etc.
- Maybe have an interface for the Payload to allow users to provide the logic of building the payload themselves. (`type PayloadBuilder interface { BuildPayload() *ContentReader}`?!?)