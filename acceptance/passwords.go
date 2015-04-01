package acceptance

import (
	"os/exec"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Passwords", func() {
	var client warrant.Warrant

	BeforeEach(func() {
		client = warrant.New(warrant.Config{
			Host:          UAAHost,
			SkipVerifySSL: true,
		})

	})

	AfterEach(func() {
		// TODO: replace with implementation that does not call out to UAAC
		cmd := exec.Command("uaac", "token", "client", "get", UAAAdminClient, "-s", UAAAdminSecret)
		output, err := cmd.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring("Successfully fetched token via client credentials grant."))
	})

	It("allows a user password to be set/updated", func() {
		var (
			user      warrant.User
			userToken string
		)

		By("creating a new user", func() {
			var err error
			user, err = client.Users.Create("user-name", "user@example.com", UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		By("setting the user password using a valid client", func() {
			err := client.Users.SetPassword(user.ID, "password", UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		By("retrieving the user token using the new password", func() {
			var err error
			userToken, err = client.Users.GetToken(user.UserName, "password", "cf", "https://uaa.cloudfoundry.com/redirect/cf")
			Expect(err).NotTo(HaveOccurred())

			decodedToken, err := client.Tokens.Decode(userToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(decodedToken.UserID).To(Equal(user.ID))
		})

		By("changing a user's own password", func() {
			err := client.Users.ChangePassword(user.ID, "password", "new-password", userToken)
			Expect(err).NotTo(HaveOccurred())
		})

		By("retrieving the user token using the new password", func() {
			var err error
			userToken, err = client.Users.GetToken(user.UserName, "new-password", "cf", "https://uaa.cloudfoundry.com/redirect/cf")
			Expect(err).NotTo(HaveOccurred())

			decodedToken, err := client.Tokens.Decode(userToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(decodedToken.UserID).To(Equal(user.ID))
		})

		By("deleting the user", func() {
			err := client.Users.Delete(user.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.Users.Get(user.ID, UAAToken)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})
	})
})
