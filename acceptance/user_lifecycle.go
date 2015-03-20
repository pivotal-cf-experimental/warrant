package acceptance

import (
	"os"
	"time"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Lifecycle", func() {
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

	It("creates, retrieves, and deletes a user", func() {
		var user warrant.User

		By("creating a new user", func() {
			var err error
			user, err = client.Users.Create("user-name", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())
			Expect(user.UserName).To(Equal("user-name"))
			Expect(user.Emails).To(ConsistOf([]string{"user@example.com"}))
			Expect(user.CreatedAt).To(BeTemporally("~", time.Now().UTC(), 10*time.Minute)) // TODO: this is weird, but server time could be divergent from local time
			Expect(user.UpdatedAt).To(BeTemporally("~", time.Now().UTC(), 10*time.Minute))
			Expect(user.Version).To(Equal(0))
			Expect(user.Emails).To(ConsistOf([]string{"user@example.com"}))
			Expect(user.Active).To(BeTrue())
			Expect(user.Verified).To(BeFalse())
			Expect(user.Origin).To(Equal("uaa"))
			//Expect(user.Groups).To(ConsistOf([]warrant.Group{})) TODO: finish up groups implementation
		})

		By("finding the user", func() {
			fetchedUser, err := client.Users.Get(user.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedUser).To(Equal(user))
		})

		By("updating the user", func() {
			updatedUser, err := client.Users.Update(user, token)
			Expect(err).NotTo(HaveOccurred())

			fetchedUser, err := client.Users.Get(user.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedUser).To(Equal(updatedUser))
		})

		By("deleting the user", func() {
			err := client.Users.Delete(user.ID, token)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.Users.Get(user.ID, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})
	})

	It("does not allow a user to be created without an email address", func() {
		_, err := client.Users.Create("invalid-email-user", "", token)
		Expect(err).To(BeAssignableToTypeOf(warrant.UnexpectedStatusError{}))
	})

	It("does not allow non-existant users to be updated", func() {
		user, err := client.Users.Create("user-name", "user@example.com", token)
		Expect(err).NotTo(HaveOccurred())

		originalUserID := user.ID
		user.ID = "non-existant-user-guid"
		_, err = client.Users.Update(user, token)
		Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))

		err = client.Users.Delete(originalUserID, token)
		Expect(err).NotTo(HaveOccurred())
	})
})
