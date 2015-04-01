package acceptance

import (
	"os/exec"
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
		})
	})

	Context("for users", func() {
		var (
			user      warrant.User
			userToken string
		)

		AfterEach(func() {
			err := warrantClient.Users.Delete(user.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		It("allows a token to be retrieved", func() {
			By("creating a new user", func() {
				var err error
				user, err = warrantClient.Users.Create("username", "user@example.com", UAAToken)
				Expect(err).NotTo(HaveOccurred())
			})

			By("setting the user password", func() {
				err := warrantClient.Users.SetPassword(user.ID, "password", UAAToken)
				Expect(err).NotTo(HaveOccurred())
			})

			By("retrieving a user token", func() {
				var err error
				userToken, err = warrantClient.Users.GetToken("username", "password", "cf", "https://uaa.cloudfoundry.com/redirect/cf")
				Expect(err).NotTo(HaveOccurred())
			})

			By("checking that the token belongs to the user", func() {
				decodedToken, err := warrantClient.Tokens.Decode(userToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(decodedToken.UserID).To(Equal(user.ID))
			})
		})
	})

	Context("for clients", func() {
		var (
			client      warrant.Client
			clientToken string
		)

		AfterEach(func() {
			// TODO: replace with implementation that does not call out to UAAC
			cmd := exec.Command("uaac", "client", "delete", client.ID)
			output, err := cmd.Output()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("client registration deleted"))
		})

		It("allows a token for a client to be retrieved", func() {
			By("creating a new client", func() {
				client = warrant.Client{
					ID:                   "client-id",
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
				clientToken, err = warrantClient.Clients.GetToken("client-id", "client-secret")
				Expect(err).NotTo(HaveOccurred())
			})

			By("checking that the token belongs to the client", func() {
				decodedToken, err := warrantClient.Tokens.Decode(clientToken)
				Expect(err).NotTo(HaveOccurred())
				Expect(decodedToken.ClientID).To(Equal(client.ID))
			})
		})
	})
})
