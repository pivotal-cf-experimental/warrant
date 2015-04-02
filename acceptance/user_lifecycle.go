package acceptance

import (
	"time"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Lifecycle", func() {
	var client warrant.Warrant

	BeforeEach(func() {
		client = warrant.New(warrant.Config{
			Host:          UAAHost,
			SkipVerifySSL: true,
			TraceWriter:   TraceWriter,
		})
	})

	It("creates, retrieves, and deletes a user", func() {
		var user warrant.User

		By("creating a new user", func() {
			var err error
			user, err = client.Users.Create("user-name", "user@example.com", UAAToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(user.UserName).To(Equal("user-name"))
			Expect(user.Emails).To(ConsistOf([]string{"user@example.com"}))
			Expect(user.CreatedAt).To(BeTemporally("~", time.Now().UTC(), 10*time.Minute))
			Expect(user.UpdatedAt).To(BeTemporally("~", time.Now().UTC(), 10*time.Minute))
			Expect(user.Version).To(Equal(0))
			Expect(user.Emails).To(ConsistOf([]string{"user@example.com"}))
			Expect(user.Active).To(BeTrue())
			Expect(user.Verified).To(BeFalse())
			Expect(user.Origin).To(Equal("uaa"))
			//Expect(user.Groups).To(ConsistOf([]warrant.Group{})) TODO: finish up groups implementation
		})

		By("finding the user", func() {
			fetchedUser, err := client.Users.Get(user.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedUser).To(Equal(user))
		})

		By("updating the user", func() {
			updatedUser, err := client.Users.Update(user, UAAToken)
			Expect(err).NotTo(HaveOccurred())

			fetchedUser, err := client.Users.Get(user.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedUser).To(Equal(updatedUser))
		})

		By("deleting the user", func() {
			err := client.Users.Delete(user.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.Users.Get(user.ID, UAAToken)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})
	})

	It("does not allow a user to be created without an email address", func() {
		_, err := client.Users.Create("invalid-email-user", "", UAAToken)
		Expect(err).To(BeAssignableToTypeOf(warrant.UnexpectedStatusError{}))
	})

	It("does not allow non-existant users to be updated", func() {
		user, err := client.Users.Create("user-name", "user@example.com", UAAToken)
		Expect(err).NotTo(HaveOccurred())

		originalUserID := user.ID
		user.ID = "non-existant-user-guid"
		_, err = client.Users.Update(user, UAAToken)
		Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))

		err = client.Users.Delete(originalUserID, UAAToken)
		Expect(err).NotTo(HaveOccurred())
	})
})
