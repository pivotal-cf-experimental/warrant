package network

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"strings"
)

type requestBody interface {
	Encode() (requestBody io.Reader, contentType string, err error)
}

type jsonRequestBody struct {
	body interface{}
}

// NewJSONRequestBody returns an object capable of being encoded
// as JSON within a request body. It will indicate that the
// consuming request has a Content-Type header value of
// 'application/json'.
func NewJSONRequestBody(body interface{}) jsonRequestBody {
	return jsonRequestBody{
		body: body,
	}
}

func (j jsonRequestBody) Encode() (requestBody io.Reader, contentType string, err error) {
	bodyJSON, err := json.Marshal(j.body)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewReader(bodyJSON), "application/json", nil
}

// NewFormRequestBody returns an object capable of being form
// encoded into a request body. It will indicate that the
// consuming request has a Content-Type header value of
// 'application/x-www-form-urlencoded'.
func NewFormRequestBody(values url.Values) formRequestBody {
	return formRequestBody(values)
}

type formRequestBody url.Values

func (f formRequestBody) Encode() (requestBody io.Reader, contentType string, err error) {
	return strings.NewReader(url.Values(f).Encode()), "application/x-www-form-urlencoded", nil
}
