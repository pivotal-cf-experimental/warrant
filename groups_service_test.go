package warrant_test

import (
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GroupsService", func() {
	var (
		service warrant.GroupsService
		token   string
		config  warrant.Config
	)

	BeforeEach(func() {
		config = warrant.Config{
			Host:          fakeUAAServer.URL(),
			SkipVerifySSL: true,
			TraceWriter:   TraceWriter,
		}
		service = warrant.NewGroupsService(config)
		token = fakeUAAServer.ClientTokenFor("admin", []string{"scim.write"}, []string{"scim"})
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

		It("requires the scim.write scope", func() {
			token = fakeUAAServer.ClientTokenFor("admin", []string{"scim.banana"}, []string{"scim"})
			_, err := service.Create("banana.write", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})

		It("requires the scim audience", func() {
			token = fakeUAAServer.ClientTokenFor("admin", []string{"scim.write"}, []string{"banana"})
			_, err := service.Create("banana.write", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})

		Context("failure cases", func() {
			It("returns an error when a group with the given name already exists", func() {
				_, err := service.Create("banana.write", token)
				Expect(err).NotTo(HaveOccurred())

				_, err = service.Create("banana.write", token)
				Expect(err).To(BeAssignableToTypeOf(warrant.DuplicateResourceError{}))
				Expect(err.Error()).To(Equal("duplicate resource: {\"message\":\"A group with displayName: banana.write already exists.\",\"error\":\"scim_resource_already_exists\"}"))
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
})
