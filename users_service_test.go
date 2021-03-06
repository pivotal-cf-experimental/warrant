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

var _ = Describe("UsersService", func() {
	var (
		service        warrant.UsersService
		clientsService warrant.ClientsService
		token          string
		config         warrant.Config
	)

	BeforeEach(func() {
		config = warrant.Config{
			Host:          fakeUAA.URL(),
			SkipVerifySSL: true,
			TraceWriter:   TraceWriter,
		}
		service = warrant.NewUsersService(config)

		clientsService = warrant.NewClientsService(config)

		var err error
		token, err = clientsService.GetToken("admin", "admin")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Create", func() {
		It("creates a new user", func() {
			user, err := service.Create("created-user", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())
			Expect(user.ID).NotTo(BeEmpty())
			Expect(user.UserName).To(Equal("created-user"))
			Expect(user.CreatedAt).To(BeTemporally("~", time.Now().UTC(), 2*time.Millisecond))
			Expect(user.UpdatedAt).To(BeTemporally("~", time.Now().UTC(), 2*time.Millisecond))
			Expect(user.Version).To(Equal(0))
			Expect(user.Emails).To(ConsistOf([]string{"user@example.com"}))
			Expect(user.Groups).To(ConsistOf([]warrant.Group{}))
			Expect(user.Active).To(BeTrue())
			Expect(user.Verified).To(BeFalse())
			Expect(user.Origin).To(Equal("uaa"))
			Expect(user.ExternalID).To(Equal(""))
			Expect(user.FormattedName).To(Equal(""))
			Expect(user.FamilyName).To(Equal(""))
			Expect(user.GivenName).To(Equal(""))
			Expect(user.MiddleName).To(Equal(""))

			fetchedUser, err := service.Get(user.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedUser).To(Equal(user))
		})

		Context("when the client does not have the scim.write scope", func() {
			It("returns an unauthorized error", func() {
				c := warrant.Client{
					ID:          "unauthorized",
					ResourceIDs: []string{"scim"},
					Authorities: []string{"scim.read"},
				}

				err := clientsService.Create(c, "secret", token)
				Expect(err).NotTo(HaveOccurred())

				t, err := clientsService.GetToken(c.ID, "secret")
				Expect(err).NotTo(HaveOccurred())

				_, err = service.Create("created-user", "user@example.com", t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		Context("when the client does not have the scim audience", func() {
			It("returns an unauthorized error", func() {
				c := warrant.Client{
					ID:          "unauthorized",
					ResourceIDs: []string{"banana"},
					Authorities: []string{"scim.write"},
				}

				err := clientsService.Create(c, "secret", token)
				Expect(err).NotTo(HaveOccurred())

				t, err := clientsService.GetToken(c.ID, "secret")
				Expect(err).NotTo(HaveOccurred())

				_, err = service.Create("created-user", "user@example.com", t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		It("requires an email address", func() {
			_, err := service.Create("created-user", "", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.BadRequestError{}))
			Expect(err.Error()).To(Equal(`bad request: {"error_description":"[Assertion failed] - this String argument must have text; it must not be null, empty, or blank","error":"invalid_scim_resource"}`))
		})

		Context("failure cases", func() {
			It("returns an error when a user with the given username already exists", func() {
				_, err := service.Create("username", "user@example.com", token)
				Expect(err).NotTo(HaveOccurred())

				_, err = service.Create("username", "user@example.com", token)
				Expect(err).To(BeAssignableToTypeOf(warrant.DuplicateResourceError{}))
				Expect(err.Error()).To(Equal("duplicate resource: {\"error_description\":\"Username already in use: username\",\"error\":\"scim_resource_already_exists\"}"))
			})

			It("returns an error when the json response is malformed", func() {
				malformedJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("this is not JSON"))
				}))
				service = warrant.NewUsersService(warrant.Config{
					Host:          malformedJSONServer.URL,
					SkipVerifySSL: true,
					TraceWriter:   TraceWriter,
				})

				_, err := service.Create("some-user", "some-user@example.com", "some-token")
				Expect(err).To(BeAssignableToTypeOf(warrant.MalformedResponseError{}))
				Expect(err).To(MatchError("malformed response: invalid character 'h' in literal true (expecting 'r')"))
			})
		})
	})

	Describe("Get", func() {
		var createdUser warrant.User

		BeforeEach(func() {
			var err error
			createdUser, err = service.Create("created-user", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the found user", func() {
			user, err := service.Get(createdUser.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(user).To(Equal(createdUser))
		})

		Context("when the client does not have the scim.read scope", func() {
			It("returns an unauthorized error", func() {
				c := warrant.Client{
					ID:          "unauthorized",
					ResourceIDs: []string{"scim"},
					Authorities: []string{"scim.write"},
				}

				err := clientsService.Create(c, "secret", token)
				Expect(err).NotTo(HaveOccurred())

				t, err := clientsService.GetToken(c.ID, "secret")
				Expect(err).NotTo(HaveOccurred())

				_, err = service.Get(createdUser.ID, t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		Context("when the client does not have the scim audience", func() {
			It("returns an unauthorized error", func() {
				c := warrant.Client{
					ID:          "unauthorized",
					ResourceIDs: []string{"banana"},
					Authorities: []string{"scim.read"},
				}

				err := clientsService.Create(c, "secret", token)
				Expect(err).NotTo(HaveOccurred())

				t, err := clientsService.GetToken(c.ID, "secret")
				Expect(err).NotTo(HaveOccurred())

				_, err = service.Get(createdUser.ID, t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the user cannot be found", func() {
				_, err := service.Get("non-existent-user-id", token)
				Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
			})

			It("returns an error when the json response is malformed", func() {
				malformedJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Write([]byte("this is not JSON"))
				}))
				service = warrant.NewUsersService(warrant.Config{
					Host:          malformedJSONServer.URL,
					SkipVerifySSL: true,
					TraceWriter:   TraceWriter,
				})

				_, err := service.Get("some-user-id", "some-token")
				Expect(err).To(BeAssignableToTypeOf(warrant.MalformedResponseError{}))
				Expect(err).To(MatchError("malformed response: invalid character 'h' in literal true (expecting 'r')"))
			})
		})
	})

	Describe("Delete", func() {
		var user warrant.User

		BeforeEach(func() {
			var err error
			user, err = service.Create("deleted-user", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("deletes the user", func() {
			err := service.Delete(user.ID, token)
			Expect(err).NotTo(HaveOccurred())

			_, err = service.Get(user.ID, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})

		Context("when the client does not have the scim.write scope", func() {
			It("returns an unauthorized error", func() {
				c := warrant.Client{
					ID:          "unauthorized",
					ResourceIDs: []string{"scim"},
					Authorities: []string{"scim.read"},
				}

				err := clientsService.Create(c, "secret", token)
				Expect(err).NotTo(HaveOccurred())

				t, err := clientsService.GetToken(c.ID, "secret")
				Expect(err).NotTo(HaveOccurred())

				err = service.Delete(user.ID, t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		Context("when the client does not have the scim audience", func() {
			It("returns an unauthorized error", func() {
				c := warrant.Client{
					ID:          "unauthorized",
					ResourceIDs: []string{"banana"},
					Authorities: []string{"scim.write"},
				}

				err := clientsService.Create(c, "secret", token)
				Expect(err).NotTo(HaveOccurred())

				t, err := clientsService.GetToken(c.ID, "secret")
				Expect(err).NotTo(HaveOccurred())

				err = service.Delete(user.ID, t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		It("returns an error when the user does not exist", func() {
			err := service.Delete("non-existant-user-guid", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})
	})

	Describe("Update", func() {
		var user warrant.User

		BeforeEach(func() {
			var err error
			user, err = service.Create("new-user", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("updates an existing user", func() {
			user.UserName = "updated-user"
			updatedUser, err := service.Update(user, token)
			Expect(err).NotTo(HaveOccurred())

			fetchedUser, err := service.Get(user.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedUser).To(Equal(updatedUser))
		})

		It("allows fields to be updated", func() {
			user.ExternalID = "external-id"
			user.FormattedName = "James Tiberius Kirk"
			user.FamilyName = "Kirk"
			user.GivenName = "James"
			user.MiddleName = "Tiberius"

			updatedUser, err := service.Update(user, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedUser.ExternalID).To(Equal(user.ExternalID))
			Expect(updatedUser.FormattedName).To(Equal(user.FormattedName))
			Expect(updatedUser.FamilyName).To(Equal(user.FamilyName))
			Expect(updatedUser.GivenName).To(Equal(user.GivenName))
			Expect(updatedUser.MiddleName).To(Equal(user.MiddleName))

			fetchedUser, err := service.Get(user.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedUser).To(Equal(updatedUser))
		})

		Context("when the client does not have the scim.write scope", func() {
			It("returns an unauthorized error", func() {
				c := warrant.Client{
					ID:          "unauthorized",
					ResourceIDs: []string{"scim"},
					Authorities: []string{"scim.read"},
				}

				err := clientsService.Create(c, "secret", token)
				Expect(err).NotTo(HaveOccurred())

				t, err := clientsService.GetToken(c.ID, "secret")
				Expect(err).NotTo(HaveOccurred())

				_, err = service.Update(user, t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		Context("when the client does not have the scim audience", func() {
			It("returns an unauthorized error", func() {
				c := warrant.Client{
					ID:          "unauthorized",
					ResourceIDs: []string{"banana"},
					Authorities: []string{"scim.write"},
				}

				err := clientsService.Create(c, "secret", token)
				Expect(err).NotTo(HaveOccurred())

				t, err := clientsService.GetToken(c.ID, "secret")
				Expect(err).NotTo(HaveOccurred())

				_, err = service.Update(user, t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		It("must match the 'If-Match' header value", func() {
			user.Version = 24
			_, err := service.Update(user, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.BadRequestError{}))
			Expect(err).To(MatchError(`bad request: {"error_description":"Missing If-Match for PUT","error":"scim"}`))
		})

		It("returns an error if the user does not exist", func() {
			user.ID = "non-existant-guid"
			_, err := service.Update(user, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})

		Context("failure cases", func() {
			It("returns an error when the json response is malformed", func() {
				malformedJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Write([]byte("this is not JSON"))
				}))
				service = warrant.NewUsersService(warrant.Config{
					Host:          malformedJSONServer.URL,
					SkipVerifySSL: true,
					TraceWriter:   TraceWriter,
				})

				_, err := service.Update(warrant.User{ID: "some-user-id"}, "some-token")
				Expect(err).To(BeAssignableToTypeOf(warrant.MalformedResponseError{}))
				Expect(err).To(MatchError("malformed response: invalid character 'h' in literal true (expecting 'r')"))
			})
		})
	})

	Describe("SetPassword", func() {
		var user warrant.User

		BeforeEach(func() {
			var err error
			user, err = service.Create("password-user", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("sets the password belonging to the given user guid", func() {
			err := service.SetPassword(user.ID, "password", token)
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the client does not have the password.write scope", func() {
			It("returns an unauthorized error", func() {
				c := warrant.Client{
					ID:          "unauthorized",
					ResourceIDs: []string{"password"},
					Authorities: []string{"password.read"},
				}

				err := clientsService.Create(c, "secret", token)
				Expect(err).NotTo(HaveOccurred())

				t, err := clientsService.GetToken(c.ID, "secret")
				Expect(err).NotTo(HaveOccurred())

				err = service.SetPassword(user.ID, "password", t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		Context("when the client does not have the password audience", func() {
			It("returns an unauthorized error", func() {
				c := warrant.Client{
					ID:          "unauthorized",
					ResourceIDs: []string{"banana"},
					Authorities: []string{"password.write"},
				}

				err := clientsService.Create(c, "secret", token)
				Expect(err).NotTo(HaveOccurred())

				t, err := clientsService.GetToken(c.ID, "secret")
				Expect(err).NotTo(HaveOccurred())

				err = service.SetPassword(user.ID, "password", t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		It("returns an error if the user does not exist", func() {
			err := service.SetPassword("non-existant-guid", "password", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})
	})

	Describe("ChangePassword", func() {
		var (
			user      warrant.User
			userToken string
		)

		BeforeEach(func() {
			var err error
			user, err = service.Create("change-password-user", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())

			err = service.SetPassword(user.ID, "old-password", token)
			Expect(err).NotTo(HaveOccurred())

			userToken = fakeUAA.UserTokenFor(user.ID, []string{}, []string{})
		})

		Context("with a user token", func() {
			It("changes the password given the old password", func() {
				err := service.ChangePassword(user.ID, "old-password", "new-password", userToken)
				Expect(err).NotTo(HaveOccurred())
			})

			It("does not change password if the old password does not match", func() {
				err := service.ChangePassword(user.ID, "bad-password", "new-password", userToken)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		Context("with a client token", func() {
			Context("when it has the password.write scope and password audience", func() {
				It("changes the password regardless of the old password", func() {
					c := warrant.Client{
						ID:          "authorized",
						ResourceIDs: []string{"password"},
						Authorities: []string{"password.write"},
					}

					err := clientsService.Create(c, "secret", token)
					Expect(err).NotTo(HaveOccurred())

					t, err := clientsService.GetToken(c.ID, "secret")
					Expect(err).NotTo(HaveOccurred())

					err = service.ChangePassword(user.ID, "bad-password", "new-password", t)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the client does not have the password.write scope", func() {
				It("returns an unauthorized error", func() {
					c := warrant.Client{
						ID:          "authorized",
						ResourceIDs: []string{"password"},
						Authorities: []string{"password.read"},
					}

					err := clientsService.Create(c, "secret", token)
					Expect(err).NotTo(HaveOccurred())

					t, err := clientsService.GetToken(c.ID, "secret")
					Expect(err).NotTo(HaveOccurred())

					err = service.ChangePassword(user.ID, "old-password", "new-password", t)
					Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
				})
			})

			Context("when the client does not have the password audience", func() {
				It("returns an unauthorized error", func() {
					c := warrant.Client{
						ID:          "authorized",
						ResourceIDs: []string{"banana"},
						Authorities: []string{"password.write"},
					}

					err := clientsService.Create(c, "secret", token)
					Expect(err).NotTo(HaveOccurred())

					t, err := clientsService.GetToken(c.ID, "secret")
					Expect(err).NotTo(HaveOccurred())

					err = service.ChangePassword(user.ID, "old-password", "new-password", t)
					Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
				})
			})
		})
	})

	Describe("GetToken", func() {
		var (
			user   warrant.User
			client warrant.Client
			scopes []string
		)

		BeforeEach(func() {
			var err error
			user, err = service.Create("username", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())

			err = service.SetPassword(user.ID, "password", token)
			Expect(err).NotTo(HaveOccurred())

			clientsService := warrant.NewClientsService(config)

			scopes = []string{"notification_preferences.read", "notification_preferences.write"}
			client = warrant.Client{
				ID:                   "some-client-id",
				Scope:                scopes,
				ResourceIDs:          []string{""},
				Authorities:          []string{"scim.read", "scim.write"},
				AuthorizedGrantTypes: []string{"implicit"},
				AccessTokenValidity:  24 * time.Hour,
				RedirectURI:          []string{"https://redirect.example.com"},
				Autoapprove:          scopes,
			}
			err = clientsService.Create(client, "", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns a valid token given a username and password", func() {
			token, err := service.GetToken("username", "password", client)
			Expect(err).NotTo(HaveOccurred())
			Expect(token).NotTo(BeEmpty())

			tokensService := warrant.NewTokensService(config)
			decodedToken, err := tokensService.Decode(token)
			Expect(err).NotTo(HaveOccurred())
			Expect(decodedToken.UserID).To(Equal(user.ID))
			Expect(decodedToken.Scopes).To(Equal(scopes))
		})

		Context("failure cases", func() {
			It("returns an error when the request does not succeed", func() {
				_, err := service.GetToken("unknown-user", "password", client)
				Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
			})

			It("returns an error when the response is not parsable", func() {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`%%%%%`))
				}))

				config.Host = server.URL
				service = warrant.NewUsersService(config)

				_, err := service.GetToken("username", "password", client)
				Expect(err).To(BeAssignableToTypeOf(warrant.MalformedResponseError{}))
				Expect(err).To(MatchError(ContainSubstring(`invalid character '%' looking for beginning of value`)))
			})

			It("returns an error when the client requesting the token does not exist", func() {
				client.ID = "missing-client"

				_, err := service.GetToken("username", "password", client)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
				Expect(err).To(MatchError(`Warrant UnauthorizedError: {"error_description":"No client with requested id: missing-client","error":"invalid_client"}`))
			})
		})
	})

	Describe("List", func() {
		var (
			user      warrant.User
			otherUser warrant.User
		)

		BeforeEach(func() {
			var err error
			user, err = service.Create("username", "xyz@example.com", token)
			Expect(err).NotTo(HaveOccurred())

			otherUser, err = service.Create("other", "abc@example.com", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("finds users that match a filter", func() {
			users, err := service.List(warrant.Query{
				Filter: fmt.Sprintf("id eq '%s'", user.ID),
			}, token)
			Expect(err).NotTo(HaveOccurred())

			Expect(users).To(HaveLen(1))
			Expect(users[0].ID).To(Equal(user.ID))
		})

		It("returns an empty list of users if nothing matches the filter", func() {
			users, err := service.List(warrant.Query{
				Filter: "id eq ''",
			}, token)
			Expect(err).NotTo(HaveOccurred())

			Expect(users).To(HaveLen(0))
		})

		It("ignores the case of the parameter in the filter", func() {
			users, err := service.List(warrant.Query{
				Filter: fmt.Sprintf("ID eq '%s'", user.ID),
			}, token)
			Expect(err).NotTo(HaveOccurred())

			Expect(users).To(HaveLen(1))
			Expect(users[0].ID).To(Equal(user.ID))
		})

		It("ignores the case of the operator in the filter", func() {
			users, err := service.List(warrant.Query{
				Filter: fmt.Sprintf("id EQ '%s'", user.ID),
			}, token)
			Expect(err).NotTo(HaveOccurred())

			Expect(users).To(HaveLen(1))
			Expect(users[0].ID).To(Equal(user.ID))
		})

		It("defaults to sorting users by date created", func() {
			users, err := service.List(warrant.Query{}, token)
			Expect(err).NotTo(HaveOccurred())

			Expect(users).To(HaveLen(2))
			Expect(users[0].ID).To(Equal(user.ID))
			Expect(users[1].ID).To(Equal(otherUser.ID))
		})

		It("returns a list of users sorted by email when sortBy is given", func() {
			users, err := service.List(warrant.Query{
				SortBy: "email",
			}, token)
			Expect(err).NotTo(HaveOccurred())

			Expect(users).To(HaveLen(2))
			Expect(users[0].Emails[0]).To(Equal(otherUser.Emails[0]))
			Expect(users[1].Emails[0]).To(Equal(user.Emails[0]))
		})

		Context("failure cases", func() {
			It("returns an error when the query is malformed", func() {
				_, err := service.List(warrant.Query{
					Filter: fmt.Sprintf("invalid-parameter eq '%s'", user.ID),
				}, token)
				Expect(err).To(BeAssignableToTypeOf(warrant.BadRequestError{}))
				Expect(err.Error()).To(Equal(`bad request: {"error_description":"Invalid filter expression: [invalid-parameter eq '` + user.ID + `']","error":"scim"}`))
			})

			It("returns an error when the JSON is malformed", func() {
				malformedJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Write([]byte("this is not JSON"))
				}))
				service = warrant.NewUsersService(warrant.Config{
					Host:          malformedJSONServer.URL,
					SkipVerifySSL: true,
					TraceWriter:   TraceWriter,
				})

				_, err := service.List(warrant.Query{
					Filter: fmt.Sprintf("id eq '%s'", user.ID),
				}, token)
				Expect(err).To(BeAssignableToTypeOf(warrant.MalformedResponseError{}))
				Expect(err).To(MatchError("malformed response: invalid character 'h' in literal true (expecting 'r')"))
			})
		})
	})
})
