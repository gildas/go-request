# go-request

[![GoDoc](https://godoc.org/github.com/gildas/go-requests?status.svg)](https://godoc.org/github.com/gildas/go-requests)
A Package to send requests to HTTP/REST services

|  |   |   |   |
---|---|---|---|
master | [![Build Status](https://dev.azure.com/keltiek/gildas/_apis/build/status/gildas.go-request?branchName=master)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=2&branchName=master) | [![Tests](https://img.shields.io/azure-devops/tests/keltiek/gildas/3/master)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=3&branchName=master) | [![coverage](https://img.shields.io/azure-devops/coverage/keltiek/gildas/3/master)](https://dev.azure.com/keltiek/gildas/_build/latest?definitionId=3&branchName=master&view=codecoverage-tab)  
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
    URL:     myURL,
    Authorization: "Basic sdfgsdfgsdfgdsfgw42agoi0s9ix"
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

So, to send an `HTTP UPDATE`, simply:

```go
payload := struct{Key string}{}
data := struct{Data string}{}
_, err := request.Send(&request.Options{
    Method:  http.MethodUPDAE,
    URL:     myURL,
    Payload: payload,
}, &data)
if err != nil {
    return err
}
```

if the PayloadType is not mentioned, it is calculated when processing the Payload.

if the payload is a `ContentReader` or a `Content`, it is used directly.

if the payload is a `map[string]string`

if the payload is a struct{}, this func will send the body as `application/json` and will marshal it.