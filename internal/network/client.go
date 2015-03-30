package network

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var _transports map[bool]http.RoundTripper

func init() {
	_transports = map[bool]http.RoundTripper{
		true:  buildTransport(true),
		false: buildTransport(false),
	}
}

func GetClient(config Config) *http.Client {
	return &http.Client{
		Transport: _transports[config.SkipVerifySSL],
	}
}

func buildTransport(skipVerifySSL bool) http.RoundTripper {
	return &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: skipVerifySSL,
		},
		Proxy: http.ProxyFromEnvironment,
		Dial: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 10 * time.Second,
	}
}

type Config struct {
	Host          string
	SkipVerifySSL bool
}

type RequestArguments struct {
	Method                string
	Path                  string
	Authorization         authorization
	IfMatch               string
	Body                  requestBody
	AcceptableStatusCodes []int
	DoNotFollowRedirects  bool
}

type requestBody interface {
	Encode() (requestBody io.Reader, contentType string, err error)
}

type jsonRequestBody struct {
	body interface{}
}

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

func NewFormRequestBody(values url.Values) formRequestBody {
	return formRequestBody(values)
}

type formRequestBody url.Values

func (f formRequestBody) Encode() (requestBody io.Reader, contentType string, err error) {
	return strings.NewReader(url.Values(f).Encode()), "application/x-www-form-urlencoded", nil
}

type authorization interface {
	Authorization() string
}

func NewTokenAuthorization(token string) tokenAuthorization {
	return tokenAuthorization(token)
}

type tokenAuthorization string

func (a tokenAuthorization) Authorization() string {
	return fmt.Sprintf("Bearer %s", a)
}

func NewBasicAuthorization(username, password string) basicAuthorization {
	return basicAuthorization{
		Username: username,
		Password: password,
	}
}

type basicAuthorization struct {
	Username string
	Password string
}

func (b basicAuthorization) Authorization() string {
	auth := b.Username + ":" + b.Password
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth)))
}

type response struct {
	Code    int
	Body    []byte
	Headers http.Header
}

type Client struct {
	config Config
}

func NewClient(config Config) Client {
	return Client{
		config: config,
	}
}

func (c Client) MakeRequest(args RequestArguments) (response, error) {
	if args.AcceptableStatusCodes == nil {
		panic("acceptable status codes for this request were not set")
	}

	var bodyReader io.Reader
	var contentType string
	if args.Body != nil {
		var err error
		bodyReader, contentType, err = args.Body.Encode()
		if err != nil {
			return response{}, newRequestBodyMarshalError(err)
		}
	}

	requestURL := c.config.Host + args.Path
	request, err := http.NewRequest(args.Method, requestURL, bodyReader)
	if err != nil {
		return response{}, newRequestConfigurationError(err)
	}

	if args.Authorization != nil {
		request.Header.Set("Authorization", args.Authorization.Authorization())
	}

	request.Header.Set("Accept", "application/json")

	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if args.IfMatch != "" {
		request.Header.Set("If-Match", args.IfMatch)
	}

	c.printRequest(request)

	var resp *http.Response
	if args.DoNotFollowRedirects {
		resp, err = GetClient(Config{
			SkipVerifySSL: c.config.SkipVerifySSL,
		}).Transport.RoundTrip(request)
	} else {
		resp, err = GetClient(Config{
			SkipVerifySSL: c.config.SkipVerifySSL,
		}).Do(request)
	}
	if err != nil {
		return response{}, newRequestHTTPError(err)
	}
	defer resp.Body.Close()

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response{}, newResponseReadError(err)
	}

	parsedResponse := response{
		Code:    resp.StatusCode,
		Body:    responseBody,
		Headers: resp.Header,
	}
	c.printResponse(parsedResponse)

	if resp.StatusCode == http.StatusNotFound {
		return response{}, newNotFoundError(responseBody)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return response{}, newUnauthorizedError(responseBody)
	}

	for _, acceptableCode := range args.AcceptableStatusCodes {
		if resp.StatusCode == acceptableCode {
			return parsedResponse, nil
		}
	}

	return response{}, newUnexpectedStatusError(resp.StatusCode, responseBody)
}

func (c Client) printRequest(request *http.Request) {
	if os.Getenv("TRACE") != "" {
		bodyCopy := bytes.NewBuffer([]byte{})
		if request.Body != nil {
			body := bytes.NewBuffer([]byte{})
			_, err := io.Copy(io.MultiWriter(body, bodyCopy), request.Body)
			if err != nil {
				panic(err)
			}

			request.Body = ioutil.NopCloser(body)
		}

		fmt.Printf("REQUEST: %s %s %s %v\n", request.Method, request.URL, bodyCopy.String(), request.Header)
	}
}

func (c Client) printResponse(resp response) {
	if os.Getenv("TRACE") != "" {
		fmt.Printf("RESPONSE: %d %s %+v\n", resp.Code, resp.Body, resp.Headers)
	}
}
