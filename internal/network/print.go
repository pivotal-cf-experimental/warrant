package network

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

// TODO: do not use os.Getenv

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

func (c Client) printResponse(resp Response) {
	if os.Getenv("TRACE") != "" {
		fmt.Printf("RESPONSE: %d %s %+v\n", resp.Code, resp.Body, resp.Headers)
	}
}
