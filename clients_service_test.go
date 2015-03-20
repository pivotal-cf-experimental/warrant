package warrant_test

import (
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
			Host:          fakeUAAServer.URL(),
			SkipVerifySSL: true,
		}
		service = warrant.NewClientsService(config)
		token = fakeUAAServer.TokenFor([]string{"clients.write", "clients.read"}, []string{"clients"})
	})

	Describe("Create/Get", func() {
		It("an error does not occur and the new client can be fetched", func() {
			client := warrant.Client{
				ID:                   "client-id",
				Scope:                []string{"openid"},
				ResourceIDs:          []string{"none"},
				Authorities:          []string{"scim.read", "scim.write"},
				AuthorizedGrantTypes: []string{"client_credentials"},
				AccessTokenValidity:  5000 * time.Second,
			}

			err := service.Create(client, "client-secret", token)
			Expect(err).NotTo(HaveOccurred())

			foundClient, err := service.Get(client.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(foundClient).To(Equal(client))
		})

		It("responds with an error when the client cannot be found", func() {
			_, err := service.Get("unknown-client", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})
	})
})
