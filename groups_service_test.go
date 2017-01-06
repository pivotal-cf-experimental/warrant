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

var _ = Describe("GroupsService", func() {
	var (
		service        warrant.GroupsService
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
		service = warrant.NewGroupsService(config)

		clientsService = warrant.NewClientsService(config)

		var err error
		token, err = clientsService.GetToken("admin", "admin")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Create", func() {
		It("creates a group given a name", func() {
			group, err := service.Create("banana.write", token)
			Expect(err).NotTo(HaveOccurred())
			Expect(group.ID).NotTo(BeEmpty())
			Expect(group.DisplayName).To(Equal("banana.write"))
			Expect(group.Version).To(Equal(0))
			Expect(group.CreatedAt).To(BeTemporally("~", time.Now().UTC(), 2*time.Millisecond))
			Expect(group.UpdatedAt).To(BeTemporally("~", time.Now().UTC(), 2*time.Millisecond))
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

				_, err = service.Create("some-group", t)
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

				_, err = service.Create("some-group", t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		Context("failure cases", func() {
			It("returns an error when a group with the given name already exists", func() {
				_, err := service.Create("banana.write", token)
				Expect(err).NotTo(HaveOccurred())

				_, err = service.Create("banana.write", token)
				Expect(err).To(BeAssignableToTypeOf(warrant.DuplicateResourceError{}))
				Expect(err.Error()).To(Equal("duplicate resource: {\"error_description\":\"A group with displayName: banana.write already exists.\",\"error\":\"scim_resource_already_exists\"}"))
			})

			It("returns an error when the json response is malformed", func() {
				malformedJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("this is not JSON"))
				}))
				service = warrant.NewGroupsService(warrant.Config{
					Host:          malformedJSONServer.URL,
					SkipVerifySSL: true,
					TraceWriter:   TraceWriter,
				})

				_, err := service.Create("banana.read", "some-token")
				Expect(err).To(BeAssignableToTypeOf(warrant.MalformedResponseError{}))
				Expect(err).To(MatchError("malformed response: invalid character 'h' in literal true (expecting 'r')"))
			})
		})
	})

	Describe("Update", func() {
		var group warrant.Group

		BeforeEach(func() {
			var err error
			group, err = service.Create("banana.read", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("updates fields an existing group", func() {
			group.DisplayName = "banana.nope"
			group.Description = "bananas and such"

			updatedGroup, err := service.Update(group, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(updatedGroup.DisplayName).To(Equal(group.DisplayName))
			Expect(updatedGroup.Description).To(Equal(group.Description))

			fetchedGroup, err := service.Get(group.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedGroup).To(Equal(updatedGroup))
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

				_, err = service.Update(group, t)
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

				_, err = service.Update(group, t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		It("must match the 'If-Match' header value", func() {
			group.Version = 24
			_, err := service.Update(group, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.BadRequestError{}))
			Expect(err).To(MatchError(`bad request: {"error_description":"Missing If-Match for PUT","error":"scim"}`))
		})

		It("returns an error if the group does not exist", func() {
			group.ID = "not-a-real-guid"
			_, err := service.Update(group, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})

		Context("failure cases", func() {
			It("returns an error when the json response is malformed", func() {
				malformedJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Write([]byte("this is not JSON"))
				}))
				service = warrant.NewGroupsService(warrant.Config{
					Host:          malformedJSONServer.URL,
					SkipVerifySSL: true,
					TraceWriter:   TraceWriter,
				})

				_, err := service.Update(warrant.Group{ID: "some-group-id"}, "some-token")
				Expect(err).To(BeAssignableToTypeOf(warrant.MalformedResponseError{}))
				Expect(err).To(MatchError("malformed response: invalid character 'h' in literal true (expecting 'r')"))
			})
		})
	})

	Describe("Get", func() {
		var createdGroup warrant.Group

		BeforeEach(func() {
			var err error
			createdGroup, err = service.Create("created-group", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the found group", func() {
			group, err := service.Get(createdGroup.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(group).To(Equal(createdGroup))
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

				_, err = service.Get(createdGroup.ID, t)
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

				_, err = service.Get(createdGroup.ID, t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the group cannot be found", func() {
				_, err := service.Get("non-existent-group-id", token)
				Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
			})

			It("returns an error when the json response is malformed", func() {
				malformedJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Write([]byte("this is not JSON"))
				}))
				service = warrant.NewGroupsService(warrant.Config{
					Host:          malformedJSONServer.URL,
					SkipVerifySSL: true,
					TraceWriter:   TraceWriter,
				})

				_, err := service.Get("some-group-id", "some-token")
				Expect(err).To(BeAssignableToTypeOf(warrant.MalformedResponseError{}))

				Expect(err).To(MatchError("malformed response: invalid character 'h' in literal true (expecting 'r')"))
			})
		})
	})

	Describe("CheckMembership", func() {
		var group warrant.Group
		var member warrant.Member

		BeforeEach(func() {
			var err error
			group, err = service.Create("some-group", token)
			Expect(err).NotTo(HaveOccurred())

			member = warrant.Member{
				Value:  "some-member-id",
				Type:   "USER",
				Origin: "uaa",
			}
		})

		It("returns OK if the member belongs to the group", func() {
			members, err := service.ListMembers(group.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(0))

			_, err = service.AddMember(group.ID, member.Value, token)
			Expect(err).NotTo(HaveOccurred())

			members, err = service.ListMembers(group.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(1))

			foundMember, found, err := service.CheckMembership(group.ID, member.Value, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(foundMember).To(Equal(member))
			Expect(found).To(BeTrue())
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

				_, _, err = service.CheckMembership(group.ID, member.Value, t)
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

				_, _, err = service.CheckMembership(group.ID, member.Value, t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the group is not found or the user is not a member", func() {
				_, found, err := service.CheckMembership(group.ID, member.Value, token)
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
			})

			It("returns an error when the json response is malformed", func() {
				malformedJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Write([]byte("this is not JSON"))
				}))
				service = warrant.NewGroupsService(warrant.Config{
					Host:          malformedJSONServer.URL,
					SkipVerifySSL: true,
					TraceWriter:   TraceWriter,
				})

				_, _, err := service.CheckMembership("some-group-id", "some-member-id", "some-token")
				Expect(err).To(BeAssignableToTypeOf(warrant.MalformedResponseError{}))
				Expect(err).To(MatchError("malformed response: invalid character 'h' in literal true (expecting 'r')"))
			})
		})
	})

	Describe("Delete", func() {
		var group warrant.Group

		BeforeEach(func() {
			var err error
			group, err = service.Create("banana.read", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("deletes the group", func() {
			err := service.Delete(group.ID, token)
			Expect(err).NotTo(HaveOccurred())

			_, err = service.Create("banana.read", token)
			Expect(err).NotTo(HaveOccurred())
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

				err = service.Delete(group.ID, t)
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

				err = service.Delete(group.ID, t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		It("returns an error when the group does not exist", func() {
			err := service.Delete("non-existant-group-guid", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})
	})

	Describe("List", func() {
		It("retrieves a list of all the groups", func() {
			writeGroup, err := service.Create("banana.write", token)
			Expect(err).NotTo(HaveOccurred())

			readGroup, err := service.Create("banana.read", token)
			Expect(err).NotTo(HaveOccurred())

			groups, err := service.List(warrant.Query{}, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(groups).To(HaveLen(2))
			Expect(groups[0].DisplayName).To(Equal(writeGroup.DisplayName))
			Expect(groups[1].DisplayName).To(Equal(readGroup.DisplayName))
		})

		It("finds groups that match a filter", func() {
			group, err := service.Create("banana.write", token)
			Expect(err).NotTo(HaveOccurred())

			groups, err := service.List(warrant.Query{
				Filter: fmt.Sprintf("displayName eq %q", "banana.write"),
			}, token)
			Expect(err).NotTo(HaveOccurred())

			Expect(groups).To(HaveLen(1))
			Expect(groups[0].ID).To(Equal(group.ID))
		})

		Context("when a group does not match filter", func() {
			It("does not return that group", func() {
				_, err := service.Create("banana.something-else", token)
				Expect(err).NotTo(HaveOccurred())

				groups, err := service.List(warrant.Query{
					Filter: fmt.Sprintf(`displayName eq "%s"`, "banana.nope"),
				}, token)
				Expect(err).NotTo(HaveOccurred())

				Expect(groups).To(HaveLen(0))
			})
		})

		It("returns a list of groups sorted by display name", func() {
			_, err := service.Create("eggplant", token)
			Expect(err).NotTo(HaveOccurred())

			_, err = service.Create("banana", token)
			Expect(err).NotTo(HaveOccurred())

			groups, err := service.List(warrant.Query{
				SortBy: "displayname",
			}, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(groups).To(HaveLen(2))
			Expect(groups[0].DisplayName).To(Equal("banana"))
			Expect(groups[1].DisplayName).To(Equal("eggplant"))
		})

		Context("when nothing matches the filter", func() {
			It("returns an empty list", func() {
				groups, err := service.List(warrant.Query{
					Filter: `displayName eq "a-non-existant-thing"`,
				}, token)
				Expect(err).NotTo(HaveOccurred())

				Expect(groups).To(HaveLen(0))
			})
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

				_, err = service.List(warrant.Query{}, t)
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

				_, err = service.List(warrant.Query{}, t)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})

		Context("failure cases", func() {
			It("returns an error when the server does not respond validly", func() {
				erroringServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
				service = warrant.NewGroupsService(warrant.Config{
					Host:          erroringServer.URL,
					SkipVerifySSL: true,
					TraceWriter:   TraceWriter,
				})

				_, err := service.List(warrant.Query{}, token)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnexpectedStatusError{}))
			})

			It("returns an error when the JSON is malformed", func() {
				malformedJSONServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
					w.Write([]byte("this is not JSON"))
				}))
				service = warrant.NewGroupsService(warrant.Config{
					Host:          malformedJSONServer.URL,
					SkipVerifySSL: true,
					TraceWriter:   TraceWriter,
				})

				_, err := service.List(warrant.Query{}, token)
				Expect(err).To(BeAssignableToTypeOf(warrant.MalformedResponseError{}))
				Expect(err).To(MatchError("malformed response: invalid character 'h' in literal true (expecting 'r')"))
			})
		})
	})
})
