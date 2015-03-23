package warrant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/pivotal-cf-experimental/warrant/internal/network"
)

type Config struct {
	Host          string
	SkipVerifySSL bool
}

type requestArguments struct {
	Method                string
	Path                  string
	Token                 string
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

func (j jsonRequestBody) Encode() (requestBody io.Reader, contentType string, err error) {
	bodyJSON, err := json.Marshal(j.body)
	if err != nil {
		return nil, "", err
	}
	return bytes.NewReader(bodyJSON), "application/json", nil
}

type formRequestBody url.Values

func (f formRequestBody) Encode() (requestBody io.Reader, contentType string, err error) {
	return strings.NewReader(url.Values(f).Encode()), "application/x-www-form-urlencoded", nil
}

type response struct {
	Code    int
	Body    []byte
	Headers http.Header
}

type Warrant struct {
	config  Config
	Users   UsersService
	Clients ClientsService
}

func New(config Config) Warrant {
	return Warrant{
		config:  config,
		Users:   NewUsersService(config),
		Clients: NewClientsService(config),
	}
}

func (w Warrant) makeRequest(requestArgs requestArguments) (response, error) {
	if requestArgs.AcceptableStatusCodes == nil {
		panic("acceptable status codes for this request were not set")
	}

	var bodyReader io.Reader
	var contentType string
	if requestArgs.Body != nil {
		var err error
		bodyReader, contentType, err = requestArgs.Body.Encode()
		if err != nil {
			return response{}, newRequestBodyMarshalError(err)
		}
	}

	requestURL := w.config.Host + requestArgs.Path
	request, err := http.NewRequest(requestArgs.Method, requestURL, bodyReader)
	if err != nil {
		return response{}, newRequestConfigurationError(err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", requestArgs.Token))
	request.Header.Set("Accept", "application/json")

	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	if requestArgs.IfMatch != "" {
		request.Header.Set("If-Match", requestArgs.IfMatch)
	}

	w.printRequest(request)

	var resp *http.Response
	if requestArgs.DoNotFollowRedirects {
		resp, err = network.GetClient(network.Config{
			SkipVerifySSL: w.config.SkipVerifySSL,
		}).Transport.RoundTrip(request)
	} else {
		resp, err = network.GetClient(network.Config{
			SkipVerifySSL: w.config.SkipVerifySSL,
		}).Do(request)
	}
	if err != nil {
		return response{}, newRequestHTTPError(err)
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return response{}, newResponseReadError(err)
	}

	parsedResponse := response{
		Code:    resp.StatusCode,
		Body:    responseBody,
		Headers: resp.Header,
	}
	w.printResponse(parsedResponse)

	if resp.StatusCode == http.StatusNotFound {
		return response{}, newNotFoundError(responseBody)
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return response{}, newUnauthorizedError(responseBody)
	}

	for _, acceptableCode := range requestArgs.AcceptableStatusCodes {
		if resp.StatusCode == acceptableCode {
			return parsedResponse, nil
		}
	}

	return response{}, newUnexpectedStatusError(resp.StatusCode, responseBody)
}

func (w Warrant) printRequest(request *http.Request) {
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

func (w Warrant) printResponse(resp response) {
	if os.Getenv("TRACE") != "" {
		fmt.Printf("RESPONSE: %d %s %+v\n", resp.Code, resp.Body, resp.Headers)
	}
}
