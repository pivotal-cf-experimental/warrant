package warrant_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var unsupportedJSONType = func() {}

var _ = Describe("Client", func() {
	var token string
	var fakeServer *httptest.Server
	var client warrant.Client

	BeforeEach(func() {
		token = "TOKEN"
		fakeServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			requestBody, err := ioutil.ReadAll(req.Body)
			if err != nil {
				panic(err)
			}

			var responseBody struct {
				Body    string      `json:"body"`
				Headers http.Header `json:"headers"`
			}
			responseBody.Body = string(requestBody)
			responseBody.Headers = req.Header

			response, err := json.Marshal(responseBody)
			if err != nil {
				panic(err)
			}

			w.WriteHeader(http.StatusOK)
			w.Write(response)
		}))

		client = warrant.NewClient(warrant.Config{
			Host:          fakeServer.URL,
			SkipVerifySSL: true,
		})
	})

	AfterEach(func() {
		fakeServer.Close()
	})

	It("has a users service", func() {
		Expect(client.Users).To(BeAssignableToTypeOf(warrant.UsersService{}))
	})

	It("has an oauth service", func() {
		Expect(client.OAuth).To(BeAssignableToTypeOf(warrant.OAuthService{}))
	})

	Describe("makeRequest", func() {
		It("can make requests", func() {
			jsonBody := map[string]interface{}{
				"hello": "goodbye",
			}

			requestArgs := warrant.NewRequestArguments(warrant.TestRequestArguments{
				Method: "GET",
				Path:   "/path",
				Token:  token,
				Body:   warrant.NewJSONRequestBody(jsonBody),
				AcceptableStatusCodes: []int{http.StatusOK},
			})
			resp, err := client.MakeRequest(requestArgs)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.Code).To(Equal(http.StatusOK))
			Expect(resp.Body).To(MatchJSON(`{
				"body": "{\"hello\":\"goodbye\"}",
				"headers": {
					"Accept":          ["application/json"],
					"Accept-Encoding": ["gzip"],
					"Authorization":   ["Bearer TOKEN"],
					"Content-Length":  ["19"],
					"Content-Type":    ["application/json"],
					"User-Agent":      ["Go 1.1 package http"]
				}
			}`))
		})

		Context("Following redirects", func() {
			var testRequestArgs warrant.TestRequestArguments

			BeforeEach(func() {
				redirectingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					if req.URL.Path == "/redirect" {
						w.Header().Set("Location", "/noredirect")
						w.WriteHeader(http.StatusFound)
						return
					}

					w.Write([]byte("did not redirect"))
				}))

				client = warrant.NewClient(warrant.Config{
					Host:          redirectingServer.URL,
					SkipVerifySSL: true,
				})

				testRequestArgs = warrant.TestRequestArguments{
					Method: "GET",
					Path:   "/redirect",
					Token:  token,
					AcceptableStatusCodes: []int{http.StatusFound, http.StatusOK},
				}
			})

			Context("when DoNotFollowRedirects is not set", func() {
				It("follows redirects to their location", func() {
					requestArgs := warrant.NewRequestArguments(testRequestArgs)
					resp, err := client.MakeRequest(requestArgs)
					Expect(err).NotTo(HaveOccurred())
					Expect(resp.Code).To(Equal(http.StatusOK))
					Expect(resp.Headers.Get("Location")).To(Equal(""))
					Expect(resp.Body).To(ContainSubstring("did not redirect"))
				})
			})

			Context("when DoNotFollowRedirects is set", func() {
				It("does not follow redirects to their location", func() {
					testRequestArgs.DoNotFollowRedirects = true
					requestArgs := warrant.NewRequestArguments(testRequestArgs)
					resp, err := client.MakeRequest(requestArgs)
					Expect(err).NotTo(HaveOccurred())
					Expect(resp.Code).To(Equal(http.StatusFound))
					Expect(resp.Headers.Get("Location")).To(Equal("/noredirect"))
				})
			})
		})

		Context("Headers", func() {
			Context("when there is a JSON body", func() {
				It("includes the Content-Type header in the request", func() {
					requestArgs := warrant.NewRequestArguments(warrant.TestRequestArguments{
						Method: "GET",
						Path:   "/path",
						Token:  token,
						Body:   warrant.NewJSONRequestBody(map[string]string{"hello": "world"}),
						AcceptableStatusCodes: []int{http.StatusOK},
					})
					resp, err := client.MakeRequest(requestArgs)
					Expect(err).NotTo(HaveOccurred())
					Expect(resp.Code).To(Equal(http.StatusOK))
					Expect(resp.Body).To(MatchJSON(`{
						"body": "{\"hello\":\"world\"}",
						"headers":{
							"Accept":          ["application/json"],
							"Accept-Encoding": ["gzip"],
							"Authorization":   ["Bearer TOKEN"],
							"Content-Length":  ["17"],
							"Content-Type":    ["application/json"],
							"User-Agent":      ["Go 1.1 package http"]
						}
					}`))
				})
			})

			Context("when there is no JSON body", func() {
				It("does not include the Content-Type header in the request", func() {
					requestArgs := warrant.NewRequestArguments(warrant.TestRequestArguments{
						Method: "GET",
						Path:   "/path",
						Token:  token,
						Body:   nil,
						AcceptableStatusCodes: []int{http.StatusOK},
					})
					resp, err := client.MakeRequest(requestArgs)
					Expect(err).NotTo(HaveOccurred())
					Expect(resp.Code).To(Equal(http.StatusOK))
					Expect(resp.Body).To(MatchJSON(`{
						"body": "",
						"headers": {
							"Accept":          ["application/json"],
							"Accept-Encoding": ["gzip"],
							"Authorization":   ["Bearer TOKEN"],
							"User-Agent":      ["Go 1.1 package http"]
						}
					}`))
				})
			})

			Context("when the If-Match argument is assigned", func() {
				It("includes the header in the request", func() {
					requestArgs := warrant.NewRequestArguments(warrant.TestRequestArguments{
						Method:  "GET",
						Path:    "/path",
						Token:   token,
						IfMatch: "45",
						Body:    nil,
						AcceptableStatusCodes: []int{http.StatusOK},
					})
					resp, err := client.MakeRequest(requestArgs)
					Expect(err).NotTo(HaveOccurred())
					Expect(resp.Code).To(Equal(http.StatusOK))
					Expect(resp.Body).To(MatchJSON(`{
						"body": "",
						"headers": {
							"Accept":          ["application/json"],
							"Accept-Encoding": ["gzip"],
							"Authorization":   ["Bearer TOKEN"],
							"If-Match":        ["45"],
							"User-Agent":      ["Go 1.1 package http"]
						}
					}`))
				})
			})

			Context("when the If-Match argument is not assigned", func() {
				It("does not include the header in the request", func() {
					requestArgs := warrant.NewRequestArguments(warrant.TestRequestArguments{
						Method: "GET",
						Path:   "/path",
						Token:  token,
						Body:   nil,
						AcceptableStatusCodes: []int{http.StatusOK},
					})
					resp, err := client.MakeRequest(requestArgs)
					Expect(err).NotTo(HaveOccurred())
					Expect(resp.Code).To(Equal(http.StatusOK))
					Expect(resp.Body).To(MatchJSON(`{
						"body": "",
						"headers": {
							"Accept":          ["application/json"],
							"Accept-Encoding": ["gzip"],
							"Authorization":   ["Bearer TOKEN"],
							"User-Agent":      ["Go 1.1 package http"]
						}
					}`))
				})
			})
		})

		Context("when errors occur", func() {
			It("returns a RequestBodyMarshalError when the request body cannot be marshalled", func() {
				requestArgs := warrant.NewRequestArguments(warrant.TestRequestArguments{
					Method: "GET",
					Path:   "/path",
					Token:  token,
					Body:   warrant.NewJSONRequestBody(unsupportedJSONType),
					AcceptableStatusCodes: []int{http.StatusOK},
				})

				_, err := client.MakeRequest(requestArgs)
				Expect(err).To(BeAssignableToTypeOf(warrant.RequestBodyMarshalError{}))
			})

			It("returns a RequestConfigurationError when the request params are bad", func() {
				client = warrant.NewClient(warrant.Config{
					Host: "://example.com",
				})

				requestArgs := warrant.NewRequestArguments(warrant.TestRequestArguments{
					Method: "GET",
					Path:   "/path",
					Token:  "token",
					Body:   nil,
					AcceptableStatusCodes: []int{http.StatusOK},
				})
				_, err := client.MakeRequest(requestArgs)
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(warrant.RequestConfigurationError{}))
			})

			It("returns a RequestHTTPError when the request fails", func() {
				client = warrant.NewClient(warrant.Config{
					Host: "banana://example.com",
				})

				requestArgs := warrant.NewRequestArguments(warrant.TestRequestArguments{
					Method: "GET",
					Path:   "/path",
					Token:  "token",
					Body:   nil,
					AcceptableStatusCodes: []int{http.StatusOK},
				})
				_, err := client.MakeRequest(requestArgs)
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(warrant.RequestHTTPError{}))
			})

			It("returns a ResponseReadError when the response cannot be read", func() {
				unintelligibleServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Header().Set("Content-Length", "100")
					w.Write([]byte("{}"))
				}))

				client = warrant.NewClient(warrant.Config{
					Host: unintelligibleServer.URL,
				})

				requestArgs := warrant.NewRequestArguments(warrant.TestRequestArguments{
					Method: "GET",
					Path:   "/path",
					Token:  "token",
					Body:   nil,
					AcceptableStatusCodes: []int{http.StatusOK},
				})
				_, err := client.MakeRequest(requestArgs)
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(warrant.ResponseReadError{}))

				unintelligibleServer.Close()
			})

			It("returns an UnexpectedStatusError when the response status is not an expected value", func() {
				requestArgs := warrant.NewRequestArguments(warrant.TestRequestArguments{
					Method: "GET",
					Path:   "/path",
					Token:  "token",
					Body:   nil,
					AcceptableStatusCodes: []int{http.StatusTeapot},
				})
				_, err := client.MakeRequest(requestArgs)
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(warrant.UnexpectedStatusError{}))
			})

			It("returns a NotFoundError when the response status is 404", func() {
				missingServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}))

				client = warrant.NewClient(warrant.Config{
					Host: missingServer.URL,
				})

				requestArgs := warrant.NewRequestArguments(warrant.TestRequestArguments{
					Method: "GET",
					Path:   "/path",
					Token:  "token",
					Body:   nil,
					AcceptableStatusCodes: []int{http.StatusOK},
				})
				_, err := client.MakeRequest(requestArgs)
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))

				missingServer.Close()
			})

			It("returns an UnauthorizedError when the response status is 401", func() {
				lockedServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
				}))

				client = warrant.NewClient(warrant.Config{
					Host: lockedServer.URL,
				})

				requestArgs := warrant.NewRequestArguments(warrant.TestRequestArguments{
					Method: "GET",
					Path:   "/path",
					Token:  "token",
					Body:   nil,
					AcceptableStatusCodes: []int{http.StatusOK},
				})
				_, err := client.MakeRequest(requestArgs)
				Expect(err).To(HaveOccurred())
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))

				lockedServer.Close()
			})
		})
	})
})

var _ = Describe("RequestBodyEncoder", func() {
	Describe("JSONRequestBody", func() {
		Describe("Encode", func() {
			It("returns a JSON encoded representation of the given object with proper content type", func() {
				var object struct {
					Hello string `json:"hello"`
				}
				object.Hello = "goodbye"

				body, contentType, err := warrant.NewJSONRequestBody(object).Encode()
				Expect(err).NotTo(HaveOccurred())
				Expect(ioutil.ReadAll(body)).To(MatchJSON(`{
					"hello": "goodbye"
				}`))
				Expect(contentType).To(Equal("application/json"))
			})

			It("returns an error when the JSON cannot be encoded", func() {
				_, _, err := warrant.NewJSONRequestBody(func() {}).Encode()
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("FormRequestBody", func() {
		Describe("Encode", func() {
			It("returns a form URL encoded representation of the given object with proper content type", func() {
				values := url.Values{
					"hello": []string{"goodbye"},
					"black": []string{"white"},
				}

				body, contentType, err := warrant.NewFormRequestBody(values).Encode()
				Expect(err).NotTo(HaveOccurred())
				Expect(ioutil.ReadAll(body)).To(BeEquivalentTo("black=white&hello=goodbye"))
				Expect(contentType).To(Equal("application/x-www-form-urlencoded"))
			})
		})
	})
})
