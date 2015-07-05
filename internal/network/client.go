package network

import (
	"io"
	"io/ioutil"
	"net/http"
)

func NewClient(config Config) Client {
	return Client{
		config: config,
	}
}

type Client struct {
	config Config
}

type Config struct {
	Host          string
	SkipVerifySSL bool
	TraceWriter   io.Writer
}

type Request struct {
	Method                string
	Path                  string
	Authorization         authorization
	IfMatch               string
	Body                  requestBody
	AcceptableStatusCodes []int
	DoNotFollowRedirects  bool
}

type Response struct {
	Code    int
	Body    []byte
	Headers http.Header
}

func (c Client) MakeRequest(req Request) (Response, error) {
	if req.AcceptableStatusCodes == nil {
		panic("acceptable status codes for this request were not set")
	}

	request, err := c.buildRequest(req)
	if err != nil {
		return Response{}, err
	}

	var resp *http.Response
	transport := buildTransport(c.config.SkipVerifySSL)
	if req.DoNotFollowRedirects {
		resp, err = transport.RoundTrip(request)
	} else {
		client := &http.Client{Transport: transport}
		resp, err = client.Do(request)
	}
	if err != nil {
		return Response{}, newRequestHTTPError(err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Response{}, newResponseReadError(err)
	}

	return c.handleResponse(req, Response{
		Code:    resp.StatusCode,
		Body:    responseBody,
		Headers: resp.Header,
	})
}

func (c Client) buildRequest(req Request) (*http.Request, error) {
	var bodyReader io.Reader
	var contentType string
	if req.Body != nil {
		var err error
		bodyReader, contentType, err = req.Body.Encode()
		if err != nil {
			return &http.Request{}, newRequestBodyMarshalError(err)
		}
	}

	requestURL := c.config.Host + req.Path
	request, err := http.NewRequest(req.Method, requestURL, bodyReader)
	if err != nil {
		return &http.Request{}, newRequestConfigurationError(err)
	}

	if req.Authorization != nil {
		request.Header.Set("Authorization", req.Authorization.Authorization())
	}

	request.Header.Set("Accept", "application/json")

	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if req.IfMatch != "" {
		request.Header.Set("If-Match", req.IfMatch)
	}

	c.printRequest(request)

	return request, nil
}

func (c Client) handleResponse(request Request, response Response) (Response, error) {
	c.printResponse(response)

	if response.Code == http.StatusNotFound {
		return Response{}, newNotFoundError(response.Body)
	}

	if response.Code == http.StatusUnauthorized || response.Code == http.StatusForbidden {
		return Response{}, newUnauthorizedError(response.Body)
	}

	for _, acceptableCode := range request.AcceptableStatusCodes {
		if response.Code == acceptableCode {
			return response, nil
		}
	}

	return Response{}, newUnexpectedStatusError(response.Code, response.Body)
}
