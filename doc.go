/*
Package go-request sends requests to HTTP/REST services

Usage


The main func allows to send HTTP request to REST servers and takes care of payloads, JSON, result collection.

Examples:

	res, err := request.Send(&request.Options{
		URL: myURL,
	}, nil)
	if err != nil {
		return err
	}
	data := struct{Data string}{}
	err := res.UnmarshalContentJSON(&data)

Here we send an HTTP GET request and unmarshal the response (a `ContentReader`).

It is also possible to let `request.Send` do the unmarshal for us:

	data := struct{Data string}{}
	_, err := request.Send(&request.Options{
		URL: myURL,
	}, &data)
	if err != nil {
		return err
	}

Authorization can be stored in the `Options.Authorization`:

	payload := struct{Key string}{}
	data := struct{Data string}{}
	_, err := request.Send(&request.Options{
		URL:           myURL,
		Authorization: request.BasicAuthorization("user", "password"),
	}, &data)
	if err != nil {
		return err
	}

or, with a Bearer Token:

	payload := struct{Key string}{}
	data := struct{Data string}{}
	_, err := request.Send(&request.Options{
		URL:           myURL,
		Authorization: request.BearerAuthorization("myTokenABCD"),
	}, &data)
	if err != nil {
		return err
	}

Objects can be sent as payloads:

	payload := struct{Key string}{}
	data := struct{Data string}{}
	_, err := request.Send(&request.Options{
		URL:     myURL,
		Payload: payload,
	}, &data)
	if err != nil {
		return err
	}

A payload will induce an HTTP POST unless mentioned.

So, to send an `HTTP PUT`, simply write:

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

To send an x-www-form, use a `map` in the payload:  

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

To send a multipart form with an attachment, use a `map`, an attachment, and one of the key must start with `>`:  

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

Notes

- if the PayloadType is not mentioned, it is calculated when processing the Payload.

- if the payload is a `ContentReader` or a `Content`, it is used directly.

- if the payload is a `map[string]xxx` where *xxx* is not `string`, the `fmt.Stringer` is used whenever possible to get the string version of the values.

- if the payload is a struct or a pointer to struct, the body is sent as `application/json` and marshaled.

- if the payload is an array or a slice, the body is sent as `application/json` and marshaled.

- The option `Logger` can be used to let the `request` library log to a `gildas/go-logger`. By default, it logs to a `NilStream` (see github.com/gildas/go-logger).

- When using a logger, you can control how much of the Request/Response Body is logged with the options `RequestBodyLogSize`/`ResponseBodyLogSize`. By default they are set to 2048 bytes. If you do not want to log them, set the options to *-1*.

TODO

- Support other kinds of `map` in the payload, like `map[string]int`, etc.

- Maybe have an interface for the Payload to allow users to provide the logic of building the payload themselves.

*/
package request