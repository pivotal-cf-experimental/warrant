package acceptance

import (
	"encoding/pem"
	"time"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tokens", func() {
	var warrantClient warrant.Warrant

	BeforeEach(func() {
		warrantClient = warrant.New(warrant.Config{
			Host:          UAAHost,
			SkipVerifySSL: true,
			TraceWriter:   TraceWriter,
		})
	})

	Context("for users", func() {
		var (
			client    warrant.Client
			user      warrant.User
			userToken string
			scopes    []string
		)

		BeforeEach(func() {
			scopes = []string{"scim.me"}
			client = warrant.Client{
				ID:                   UAADefaultClientID,
				Scope:                scopes,
				ResourceIDs:          []string{""},
				Authorities:          []string{"scim.read", "scim.write"},
				AuthorizedGrantTypes: []string{"implicit"},
				AccessTokenValidity:  24 * time.Hour,
				RedirectURI:          []string{"https://redirect.example.com"},
				Autoapprove:          scopes,
			}

			err := warrantClient.Clients.Create(client, "", UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			err := warrantClient.Users.Delete(user.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())

			err = warrantClient.Clients.Delete(client.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		It("allows a token to be retrieved", func() {
			By("creating a new user", func() {
				var err error
				user, err = warrantClient.Users.Create(UAADefaultUsername, "warrant-user@example.com", UAAToken)
				Expect(err).NotTo(HaveOccurred())
			})

			By("setting the user password", func() {
				err := warrantClient.Users.SetPassword(user.ID, "password", UAAToken)
				Expect(err).NotTo(HaveOccurred())
			})

			By("retrieving a user token", func() {
				var err error
				userToken, err = warrantClient.Users.GetToken(UAADefaultUsername, "password", client)
				Expect(err).NotTo(HaveOccurred())
			})

			By("checking that the token belongs to the user and has correct scopes", func() {
				decodedToken, err := warrantClient.Tokens.Decode(userToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(decodedToken.UserID).To(Equal(user.ID))
				Expect(decodedToken.Scopes).To(Equal(scopes))
			})
		})
	})

	Context("for clients", func() {
		var (
			client      warrant.Client
			clientToken string
		)

		AfterEach(func() {
			err := warrantClient.Clients.Delete(client.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		It("allows a token for a client to be retrieved", func() {
			By("creating a new client", func() {
				client = warrant.Client{
					ID:                   UAADefaultClientID,
					Scope:                []string{},
					ResourceIDs:          []string{""},
					Authorities:          []string{"scim.read", "scim.write"},
					AuthorizedGrantTypes: []string{"client_credentials"},
					AccessTokenValidity:  24 * time.Hour,
				}

				err := warrantClient.Clients.Create(client, "client-secret", UAAToken)
				Expect(err).NotTo(HaveOccurred())
			})

			By("retrieving the client token", func() {
				var err error
				clientToken, err = warrantClient.Clients.GetToken(UAADefaultClientID, "client-secret")
				Expect(err).NotTo(HaveOccurred())
			})

			By("checking that the token belongs to the client", func() {
				decodedToken, err := warrantClient.Tokens.Decode(clientToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(decodedToken.ClientID).To(Equal(client.ID))
			})
		})
	})

	Context("fetching the signing key", func() {
		It("can fetch a valid signing key from the server", func() {
			signingKey, err := warrantClient.Tokens.GetSigningKey(UAAAdminClient, UAAAdminSecret)
			Expect(err).NotTo(HaveOccurred())
			Expect(signingKey.Algorithm).To(Equal("HMACSHA256"))

			block, _ := pem.Decode([]byte(signingKey.Value))
			Expect(block).NotTo(BeNil())
			Expect(block.Type).To(Equal("PUBLIC KEY"))
		})
	})
})
