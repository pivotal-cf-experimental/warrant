package acceptance

import (
	"fmt"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Finding a user from UAA", func() {
	var (
		warrantClient warrant.Warrant
		user          warrant.User
	)

	BeforeEach(func() {
		warrantClient = warrant.New(warrant.Config{
			Host:          UAAHost,
			SkipVerifySSL: true,
			TraceWriter:   TraceWriter,
		})

		var err error
		user, err = warrantClient.Users.Create("username", "username@example.com", UAAToken)
		Expect(err).NotTo(HaveOccurred())
	})

	It("finds users given a filter", func() {
		users, err := warrantClient.Users.Find(warrant.UsersQuery{
			Filter: fmt.Sprintf("id eq '%s'", user.ID),
		}, UAAToken)
		Expect(err).NotTo(HaveOccurred())

		Expect(users).To(HaveLen(1))
		Expect(users[0].ID).To(Equal(user.ID))
	})
})
