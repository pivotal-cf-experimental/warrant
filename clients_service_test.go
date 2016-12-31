package warrant_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ClientsService", func() {
	var (
		service warrant.ClientsService
		token   string
		config  warrant.Config
	)

	BeforeEach(func() {
		config = warrant.Config{
			Host:          fakeUAA.URL(),
			SkipVerifySSL: true,
			TraceWriter:   TraceWriter,
		}
		service = warrant.NewClientsService(config)
		token = fakeUAA.ClientTokenFor("admin", []string{"clients.write", "clients.read"}, []string{"clients"})
	})

	Describe("Create/Get/Update", func() {
		var client warrant.Client

		BeforeEach(func() {
			client = warrant.Client{
				ID:                   "client-id",
				Scope:                []string{"notification_preferences.read", "openid"},
				ResourceIDs:          []string{"none"},
				Authorities:          []string{"scim.read", "scim.write"},
				AuthorizedGrantTypes: []string{"authorization_code"},
				AccessTokenValidity:  5000 * time.Second,
				RedirectURI:          []string{"https://redirect.example.com"},
				Autoapprove:          []string{"openid"},
			}
		})

		It("can create a new client, retrieve it and update it", func() {
			err := service.Create(client, "client-secret", token)
			Expect(err).NotTo(HaveOccurred())

			foundClient, err := service.Get(client.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(foundClient).To(Equal(client))

			client.Scope = []string{"bananas.pick", "openid"}
			client.Authorities = []string{"scim.write"}
			client.AuthorizedGrantTypes = []string{"authorization_code", "client_credentials"}
			client.RedirectURI = []string{"https://redirect.example.com/sessions/create"}
			client.Autoapprove = []string{"notification_preferences.read", "openid"}
			err = service.Update(client, token)
			Expect(err).NotTo(HaveOccurred())

			updatedClient, err := service.Get(client.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedClient).To(Equal(client))
		})

		It("can create a new client without a secret", func() {
			client = warrant.Client{
				ID:                   "client-id",
				Scope:                []string{"openid"},
				ResourceIDs:          []string{"none"},
				Authorities:          []string{"scim.read", "scim.write"},
				AuthorizedGrantTypes: []string{"implicit"},
				AccessTokenValidity:  5000 * time.Second,
				RedirectURI:          []string{"https://redirect.example.com"},
				Autoapprove:          []string{},
			}

			err := service.Create(client, "", token)
			Expect(err).NotTo(HaveOccurred())

			foundClient, err := service.Get(client.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(foundClient).To(Equal(client))
		})

		It("responds with an error when the client includes a redirect_uri without the correct grant types", func() {
			client.AuthorizedGrantTypes = []string{"client_credentials"}
			client.RedirectURI = []string{"https://redirect.example.com"}

			err := service.Create(client, "client-secret", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.BadRequestError{}))
			Expect(err.Error()).To(Equal(`bad request: {"error_description":"A redirect_uri can only be used by implicit or authorization_code grant types.","error":"invalid_client"}`))
		})

		It("responds with an error when the client cannot be created", func() {
			client.AuthorizedGrantTypes = []string{"invalid-grant-type"}
			err := service.Create(client, "client-secret", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.BadRequestError{}))
			Expect(err.Error()).To(Equal(`bad request: {"error_description":"invalid-grant-type is not an allowed grant type. Must be one of: [implicit refresh_token authorization_code client_credentials password]","error":"invalid_client"}`))
		})

		It("responds with an error when the client cannot be found", func() {
			_, err := service.Get("unknown-client", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})

		Context("failure cases", func() {
			It("returns an error if the json response is malformed", func() {
				malformedJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Write([]byte("this is not JSON"))
				}))
				service = warrant.NewClientsService(warrant.Config{
					Host:          malformedJSONServer.URL,
					SkipVerifySSL: true,
					TraceWriter:   TraceWriter,
				})

				_, err := service.Get("some-client", "some-token")
				Expect(err).To(BeAssignableToTypeOf(warrant.MalformedResponseError{}))
				Expect(err).To(MatchError("malformed response: invalid character 'h' in literal true (expecting 'r')"))
			})
		})
	})

	Describe("GetToken", func() {
		var (
			client       warrant.Client
			clientSecret string
		)

		BeforeEach(func() {
			client = warrant.Client{
				ID:                   "client-id",
				Scope:                []string{"openid", "bananas.eat"},
				ResourceIDs:          []string{"none"},
				Authorities:          []string{"scim.read", "scim.write"},
				AuthorizedGrantTypes: []string{"client_credentials"},
				AccessTokenValidity:  5000 * time.Second,
			}
			clientSecret = "client-secret"

			err := service.Create(client, clientSecret, token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("retrieves a token for the client given a valid secret", func() {
			clientToken, err := service.GetToken(client.ID, clientSecret)
			Expect(err).NotTo(HaveOccurred())

			tokensService := warrant.NewTokensService(config)
			decodedToken, err := tokensService.Decode(clientToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(decodedToken).To(Equal(warrant.Token{
				ClientID: client.ID,
				Scopes:   []string{"openid", "bananas.eat"},
				Issuer:   fmt.Sprintf("%s/oauth/token", fakeUAA.URL()),
			}))
		})

		Context("failure cases", func() {
			It("returns an error if the json response is malformed", func() {
				malformedJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Write([]byte("this is not JSON"))
				}))
				service = warrant.NewClientsService(warrant.Config{
					Host:          malformedJSONServer.URL,
					SkipVerifySSL: true,
					TraceWriter:   TraceWriter,
				})

				_, err := service.GetToken("some-client", "some-secret")
				Expect(err).To(BeAssignableToTypeOf(warrant.MalformedResponseError{}))
				Expect(err).To(MatchError("malformed response: invalid character 'h' in literal true (expecting 'r')"))
			})
		})
	})

	Describe("Delete", func() {
		var client warrant.Client

		BeforeEach(func() {
			client = warrant.Client{
				ID:                   "client-id",
				Scope:                []string{"openid"},
				ResourceIDs:          []string{"none"},
				Authorities:          []string{"scim.read", "scim.write"},
				AuthorizedGrantTypes: []string{"client_credentials"},
				AccessTokenValidity:  5000 * time.Second,
			}

			err := service.Create(client, "secret", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("deletes the client", func() {
			err := service.Delete(client.ID, token)
			Expect(err).NotTo(HaveOccurred())

			_, err = service.Get(client.ID, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})

		It("errors when the token is unauthorized", func() {
			token = fakeUAA.ClientTokenFor("admin", []string{"clients.foo", "clients.boo"}, []string{"clients"})
			err := service.Delete(client.ID, token)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})
	})

	Describe("List", func() {
		var client warrant.Client
		var otherClient warrant.Client

		BeforeEach(func() {
			client = warrant.Client{
				ID:                   "xyz-client",
				Name:                 "client",
				Scope:                []string{"openid"},
				ResourceIDs:          []string{"none"},
				Authorities:          []string{"scim.read", "scim.write"},
				AuthorizedGrantTypes: []string{"client_credentials"},
				AccessTokenValidity:  5000 * time.Second,
			}

			otherClient = warrant.Client{
				ID:                   "abc-client",
				Name:                 "other-client",
				Scope:                []string{"openid"},
				ResourceIDs:          []string{"none"},
				Authorities:          []string{"scim.read", "scim.write"},
				AuthorizedGrantTypes: []string{"client_credentials"},
				AccessTokenValidity:  5000 * time.Second,
			}

			err := service.Create(client, "secret", token)
			Expect(err).NotTo(HaveOccurred())
			err = service.Create(otherClient, "secret", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("finds clients that match a filter", func() {
			clients, err := service.List(warrant.Query{
				Filter: fmt.Sprintf("id EQ '%s'", client.ID),
			}, token)
			Expect(err).NotTo(HaveOccurred())

			Expect(clients).To(HaveLen(1))
			Expect(clients[0].ID).To(Equal("xyz-client"))
		})

		It("returns an empty list of clients if nothing matches the filter", func() {
			clients, err := service.List(warrant.Query{
				Filter: "id eq 'not-a-real-id'",
			}, token)
			Expect(err).NotTo(HaveOccurred())

			Expect(clients).To(HaveLen(0))
		})

		It("returns a list of clients sorted by id", func() {
			clients, err := service.List(warrant.Query{}, token)
			Expect(err).NotTo(HaveOccurred())

			Expect(clients).To(HaveLen(2))
			Expect(clients[0].ID).To(Equal("abc-client"))
			Expect(clients[0].Name).To(Equal("other-client"))
			Expect(clients[1].ID).To(Equal("xyz-client"))
			Expect(clients[1].Name).To(Equal("client"))
		})

		It("returns a list of clients sorted by name", func() {
			clients, err := service.List(warrant.Query{
				SortBy: "name",
			}, token)
			Expect(err).NotTo(HaveOccurred())

			Expect(clients).To(HaveLen(2))
			Expect(clients[0].Name).To(Equal("client"))
			Expect(clients[1].Name).To(Equal("other-client"))
		})

		It("errors when the token is unauthorized", func() {
			token = fakeUAA.ClientTokenFor("admin", []string{"clients.foo", "clients.boo"}, []string{"clients"})
			err := service.Delete(client.ID, token)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})
	})
})
