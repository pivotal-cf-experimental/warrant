package acceptance

import (
	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Groups", func() {
	var (
		client warrant.Warrant
		group  warrant.Group
		err    error
	)

	BeforeEach(func() {
		client = warrant.New(warrant.Config{
			Host:          UAAHost,
			SkipVerifySSL: true,
			TraceWriter:   TraceWriter,
		})
	})

	It("creates, retrieves, and deletes a group", func() {
		By("creating a new group", func() {
			group, err = client.Groups.Create("banana.read", UAAToken)
			Expect(err).NotTo(HaveOccurred())

			Expect(group.ID).NotTo(BeEmpty())
			Expect(group.DisplayName).To(Equal("banana.read"))
		})

		By("getting the created group", func() {
			group, err = client.Groups.Get(group.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())

			Expect(group.ID).NotTo(BeEmpty())
			Expect(group.DisplayName).To(Equal("banana.read"))
		})

		By("deleting the created group", func() {
			err = client.Groups.Delete(group.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())

			_, err = client.Groups.Get(group.ID, UAAToken)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})
	})
})
