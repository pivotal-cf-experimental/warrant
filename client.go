package warrant

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

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
	Body                  interface{}
	AcceptableStatusCodes []int
}

type Client struct {
	config Config
	Users  UsersService
}

func NewClient(config Config) Client {
	return Client{
		config: config,
		Users:  NewUsersService(config),
	}
}

func (c Client) makeRequest(requestArgs requestArguments) (int, []byte, error) {
	if requestArgs.AcceptableStatusCodes == nil {
		panic("acceptable status codes for this request were not set")
	}

	var bodyReader io.Reader
	if requestArgs.Body != nil {
		requestBody, err := json.Marshal(requestArgs.Body)
		if err != nil {
			return 0, []byte{}, newRequestBodyMarshalError(err)
		}
		bodyReader = bytes.NewReader(requestBody)
	}

	requestURL := c.config.Host + requestArgs.Path
	request, err := http.NewRequest(requestArgs.Method, requestURL, bodyReader)
	if err != nil {
		return 0, []byte{}, newRequestConfigurationError(err)
	}
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", requestArgs.Token))

	if requestArgs.Body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if requestArgs.IfMatch != "" {
		request.Header.Set("If-Match", requestArgs.IfMatch)
	}

	c.printRequest(request)

	response, err := network.GetClient(network.Config{
		SkipVerifySSL: c.config.SkipVerifySSL,
	}).Do(request)
	if err != nil {
		return 0, []byte{}, newRequestHTTPError(err)
	}

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, []byte{}, newResponseReadError(err)
	}

	c.printResponse(response.StatusCode, responseBody)

	if response.StatusCode == http.StatusNotFound {
		return response.StatusCode, responseBody, newNotFoundError(responseBody)
	}

	if response.StatusCode == http.StatusUnauthorized {
		return response.StatusCode, responseBody, newUnauthorizedError(responseBody)
	}

	for _, acceptableCode := range requestArgs.AcceptableStatusCodes {
		if response.StatusCode == acceptableCode {
			return response.StatusCode, responseBody, nil
		}
	}

	return response.StatusCode, responseBody, newUnexpectedStatusError(response.StatusCode, responseBody)
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

func (c Client) printResponse(status int, body []byte) {
	if os.Getenv("TRACE") != "" {
		fmt.Printf("RESPONSE: %d %s\n", status, body)
	}
}
