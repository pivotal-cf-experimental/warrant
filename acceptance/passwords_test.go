package acceptance

import (
	"time"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Passwords", func() {
	var (
		warrantClient warrant.Warrant
		user          warrant.User
		client        warrant.Client
	)

	BeforeEach(func() {
		warrantClient = warrant.New(warrant.Config{
			Host:          UAAHost,
			SkipVerifySSL: true,
			TraceWriter:   TraceWriter,
		})

		client = warrant.Client{
			ID:                   UAADefaultClientID,
			Scope:                []string{"openid", "password.write"},
			ResourceIDs:          []string{},
			Authorities:          []string{"scim.read", "scim.write"},
			AuthorizedGrantTypes: []string{"password"},
			AccessTokenValidity:  24 * time.Hour,
			RedirectURI:          []string{"https://redirect.example.com"},
			Autoapprove:          []string{"openid", "password.write"},
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

	It("allows a user password to be set/updated", func() {
		var (
			userToken string
		)

		By("creating a new user", func() {
			var err error
			user, err = warrantClient.Users.Create(UAADefaultUsername, "warrant-user@example.com", UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		By("setting the user password using a valid client", func() {
			err := warrantClient.Users.SetPassword(user.ID, "password", UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		By("retrieving the user token using the new password", func() {
			var err error
			userToken, err = warrantClient.Users.GetToken(user.UserName, "password", client)
			Expect(err).NotTo(HaveOccurred())

			decodedToken, err := warrantClient.Tokens.Decode(userToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(decodedToken.UserID).To(Equal(user.ID))
		})

		By("changing a user's own password", func() {
			err := warrantClient.Users.ChangePassword(user.ID, "password", "new-password", userToken)
			Expect(err).NotTo(HaveOccurred())
		})

		By("retrieving the user token using the new password", func() {
			var err error
			userToken, err = warrantClient.Users.GetToken(user.UserName, "new-password", client)
			Expect(err).NotTo(HaveOccurred())

			decodedToken, err := warrantClient.Tokens.Decode(userToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(decodedToken.UserID).To(Equal(user.ID))
		})
	})
})
