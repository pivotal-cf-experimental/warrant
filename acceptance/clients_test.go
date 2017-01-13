package acceptance

import (
	"time"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client Lifecycle", func() {
	var (
		warrantClient warrant.Warrant
		client        warrant.Client
	)

	BeforeEach(func() {
		client = warrant.Client{
			ID:                   UAADefaultClientID,
			Scope:                []string{"openid"},
			ResourceIDs:          []string{"none"},
			Authorities:          []string{"scim.read", "scim.write"},
			AuthorizedGrantTypes: []string{"client_credentials"},
			AccessTokenValidity:  5000 * time.Second,
			RedirectURI:          []string{"https://redirect.example.com"},
			Autoapprove:          []string{},
		}

		warrantClient = warrant.New(warrant.Config{
			Host:          UAAHost,
			SkipVerifySSL: true,
			TraceWriter:   TraceWriter,
		})
	})

	AfterEach(func() {
		_, err := warrantClient.Clients.Get(client.ID, UAAToken)
		if err == nil {
			err := warrantClient.Clients.Delete(client.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())
		}
	})

	It("creates, and retrieves a client", func() {
		By("creating a client", func() {
			err := warrantClient.Clients.Create(client, "secret", UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		By("finding the client", func() {
			fetchedClient, err := warrantClient.Clients.Get(client.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedClient).To(Equal(client))
		})

		By("updating the client", func() {
			client.Scope = []string{"bananas.eat", "openid"}
			client.Authorities = []string{"scim.read"}
			client.AuthorizedGrantTypes = []string{"client_credentials", "implicit"}
			client.RedirectURI = []string{"https://redirect.example.com/sessions/create"}

			err := warrantClient.Clients.Update(client, UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		By("retrieving the updated client", func() {
			fetchedClient, err := warrantClient.Clients.Get(client.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedClient).To(Equal(client))
		})
	})

	Context("when there is more than one client", func() {
		var (
			client1, client2, client3 warrant.Client
		)

		BeforeEach(func() {
			client1 = warrant.Client{
				ID:                   "warrant-client-one",
				Scope:                []string{"openid"},
				ResourceIDs:          []string{"none"},
				Authorities:          []string{"scim.read", "scim.write"},
				AuthorizedGrantTypes: []string{"client_credentials"},
				AccessTokenValidity:  5000 * time.Second,
				RedirectURI:          []string{"https://redirect.example.com"},
				Autoapprove:          []string{},
			}
			err := warrantClient.Clients.Create(client1, "secret", UAAToken)
			Expect(err).NotTo(HaveOccurred())

			client2 = warrant.Client{
				ID:                   "warrant-client-two",
				Scope:                []string{"openid"},
				ResourceIDs:          []string{"none"},
				Authorities:          []string{"scim.read", "scim.write"},
				AuthorizedGrantTypes: []string{"client_credentials"},
				AccessTokenValidity:  5000 * time.Second,
				RedirectURI:          []string{"https://redirect.example.com"},
				Autoapprove:          []string{},
			}
			err = warrantClient.Clients.Create(client2, "secret", UAAToken)
			Expect(err).NotTo(HaveOccurred())

			client3 = warrant.Client{
				ID:                   "warrant-client-three",
				Scope:                []string{"openid"},
				ResourceIDs:          []string{"none"},
				Authorities:          []string{"scim.read", "scim.write"},
				AuthorizedGrantTypes: []string{"client_credentials"},
				AccessTokenValidity:  5000 * time.Second,
				RedirectURI:          []string{"https://redirect.example.com"},
				Autoapprove:          []string{},
			}
			err = warrantClient.Clients.Create(client3, "secret", UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := warrantClient.Clients.Delete(client1.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())

			err = warrantClient.Clients.Delete(client2.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())

			err = warrantClient.Clients.Delete(client3.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		It("lists clients", func() {
			clients, err := warrantClient.Clients.List(warrant.Query{
				Filter: "client_id co 'warrant-client'",
			}, UAAToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(clients).To(HaveLen(3))
		})
	})

	It("creates clients with implicit grants", func() {
		client = warrant.Client{
			ID:                   UAADefaultClientID,
			Scope:                []string{"openid"},
			ResourceIDs:          []string{"none"},
			Authorities:          []string{"scim.read", "scim.write"},
			AuthorizedGrantTypes: []string{"implicit"},
			AccessTokenValidity:  5000 * time.Second,
			RedirectURI:          []string{"https://redirect.example.com"},
			Autoapprove:          []string{},
		}

		By("creating a client", func() {
			err := warrantClient.Clients.Create(client, "", UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		By("finding the client", func() {
			fetchedClient, err := warrantClient.Clients.Get(client.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedClient).To(Equal(client))
		})
	})

	It("rejects requests from clients that do not have clients.write scope", func() {
		var token string

		By("creating a client", func() {
			err := warrantClient.Clients.Create(client, "secret", UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		By("fetching the new client token", func() {
			var err error

			token, err = warrantClient.Clients.GetToken(client.ID, "secret")
			Expect(err).NotTo(HaveOccurred())
		})

		By("using the new client token to delete the client", func() {
			err := warrantClient.Clients.Delete(client.ID, token)
			Expect(err).To(HaveOccurred())
			Expect(err).To(BeAssignableToTypeOf(warrant.ForbiddenError{}))
		})
	})
})
