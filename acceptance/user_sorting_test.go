package acceptance

import (
	"time"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Sorting", func() {
	var (
		client    warrant.Warrant
		user      warrant.User
		otherUser warrant.User
	)

	BeforeEach(func() {
		client = warrant.New(warrant.Config{
			Host:          UAAHost,
			SkipVerifySSL: true,
			TraceWriter:   TraceWriter,
		})
	})

	AfterEach(func() {
		client.Users.Delete(user.ID, UAAToken)
		client.Users.Delete(otherUser.ID, UAAToken)

		_, err := client.Users.Get(user.ID, UAAToken)
		Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))

		_, err = client.Users.Get(otherUser.ID, UAAToken)
		Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
	})

	It("sorts users by field", func() {
		var err error
		user, err = client.Users.Create(UAADefaultUsername, "xyz@example.com", UAAToken)
		Expect(err).NotTo(HaveOccurred())
		Expect(user.UserName).To(Equal(UAADefaultUsername))
		Expect(user.Emails).To(ConsistOf([]string{"xyz@example.com"}))
		Expect(user.CreatedAt).To(BeTemporally("~", time.Now().UTC(), 10*time.Minute))
		Expect(user.UpdatedAt).To(BeTemporally("~", time.Now().UTC(), 10*time.Minute))
		Expect(user.Version).To(Equal(0))
		Expect(user.Active).To(BeTrue())
		Expect(user.Verified).To(BeTrue())
		Expect(user.Origin).To(Equal("uaa"))

		otherUser, err = client.Users.Create("warrant-user-2", "abc@example.com", UAAToken)
		Expect(err).NotTo(HaveOccurred())
		Expect(otherUser.UserName).To(Equal("warrant-user-2"))
		Expect(otherUser.Emails).To(ConsistOf([]string{"abc@example.com"}))
		Expect(otherUser.CreatedAt).To(BeTemporally("~", time.Now().UTC(), 10*time.Minute))
		Expect(otherUser.UpdatedAt).To(BeTemporally("~", time.Now().UTC(), 10*time.Minute))
		Expect(otherUser.Version).To(Equal(0))
		Expect(otherUser.Active).To(BeTrue())
		Expect(otherUser.Verified).To(BeTrue())
		Expect(otherUser.Origin).To(Equal("uaa"))

		allUsers, err := client.Users.List(warrant.Query{
			Filter: "userName co 'warrant-user'",
			SortBy: "email",
		}, UAAToken)
		Expect(err).NotTo(HaveOccurred())

		Expect(allUsers).To(HaveLen(2))
		Expect(allUsers[0].Emails[0]).To(Equal(otherUser.Emails[0]))
		Expect(allUsers[1].Emails[0]).To(Equal(user.Emails[0]))
	})
})
