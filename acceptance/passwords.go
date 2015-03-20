package acceptance

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Passwords", func() {
	var (
		client warrant.Warrant
		token  string
	)

	BeforeEach(func() {
		token = os.Getenv("UAA_TOKEN")

		client = warrant.New(warrant.Config{
			Host:          os.Getenv("UAA_HOST"),
			SkipVerifySSL: true,
		})
	})

	AfterEach(func() {
		// TODO: replace with implementation that does not call out to UAAC
		cmd := exec.Command("uaac", "token", "client", "get", os.Getenv("UAA_ADMIN_CLIENT"), "-s", os.Getenv("UAA_ADMIN_SECRET"))
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
			user, err = client.Users.Create("user-name", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())
		})

		By("setting the user password using a valid client", func() {
			err := client.Users.SetPassword(user.ID, "password", token)
			Expect(err).NotTo(HaveOccurred())
		})

		By("retrieving the user token using the new password", func() {
			// TODO: replace with implementation that does not call out to UAAC
			cmd := exec.Command("uaac", "token", "get", user.UserName, "password")
			output, err := cmd.Output()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("Successfully fetched token via implicit (with posted credentials) grant."))

			cmd = exec.Command("uaac", "token", "decode")
			output, err = cmd.Output()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring(fmt.Sprintf("user_id: %s", user.ID)))

			cmd = exec.Command("uaac", "context")
			output, err = cmd.Output()
			Expect(err).NotTo(HaveOccurred())
			matches := regexp.MustCompile(`access_token: (.*)\n`).FindStringSubmatch(string(output))
			Expect(matches).To(HaveLen(2))

			userToken = matches[1]
			Expect(userToken).NotTo(BeEmpty())
		})

		By("changing a user's own password", func() {
			err := client.Users.ChangePassword(user.ID, "password", "new-password", userToken)
			Expect(err).NotTo(HaveOccurred())
		})

		By("retrieving the user token using the new password", func() {
			// TODO: replace with implementation that does not call out to UAAC
			cmd := exec.Command("uaac", "token", "get", user.UserName, "new-password")
			output, err := cmd.Output()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring("Successfully fetched token via implicit (with posted credentials) grant."))

			cmd = exec.Command("uaac", "token", "decode")
			output, err = cmd.Output()
			Expect(err).NotTo(HaveOccurred())
			Expect(output).To(ContainSubstring(fmt.Sprintf("user_id: %s", user.ID)))

			cmd = exec.Command("uaac", "context")
			output, err = cmd.Output()
			Expect(err).NotTo(HaveOccurred())
			matches := regexp.MustCompile(`access_token: (.*)\n`).FindStringSubmatch(string(output))
			Expect(matches).To(HaveLen(2))

			userToken = matches[1]
			Expect(userToken).NotTo(BeEmpty())
		})

		By("deleting the user", func() {
			err := client.Users.Delete(user.ID, token)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.Users.Get(user.ID, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})
	})
})
