package acceptance

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tokens", func() {
	var (
		client    warrant.Client
		token     string
		user      warrant.User
		userToken string
	)

	BeforeEach(func() {
		token = os.Getenv("UAA_TOKEN")

		client = warrant.NewClient(warrant.Config{
			Host:          os.Getenv("UAA_HOST"),
			SkipVerifySSL: true,
		})
	})

	AfterEach(func() {
		var err error
		err = client.Users.Delete(user.ID, token)
		Expect(err).NotTo(HaveOccurred())
	})

	It("allows a token for a user to be retrieved", func() {
		By("creating a new user", func() {
			var err error
			user, err = client.Users.Create("username", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())
		})

		By("setting the user password", func() {
			err := client.Users.SetPassword(user.ID, "password", token)
			Expect(err).NotTo(HaveOccurred())
		})

		By("retrieving a user token", func() {
			var err error
			userToken, err = client.OAuth.GetUserToken("username", "password", "cf", "https://uaa.cloudfoundry.com/redirect/cf")
			Expect(err).NotTo(HaveOccurred())
		})

		By("checking that the token belongs to the user", func() {
			// TODO: replace with implementation that does not call out to UAAC
			cmd := exec.Command("uaac", "token", "decode", userToken)
			output, err := cmd.Output()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring(fmt.Sprintf("user_id: %s", user.ID)))
		})
	})
})
