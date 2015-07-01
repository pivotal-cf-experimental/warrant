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
	)

	BeforeEach(func() {
		client = warrant.New(warrant.Config{
			Host:          UAAHost,
			SkipVerifySSL: true,
			TraceWriter:   TraceWriter,
		})
	})

	AfterEach(func() {
		//client.Groups.Delete(group.ID, UAAToken)

		//_, err := client.Group.Get(group.ID, UAAToken)
		//Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
	})

	It("creates, retrieves, and deletes a group", func() {
		By("creating a new group", func() {
			var err error
			group, err = client.Groups.Create("banana.read", UAAToken)
			Expect(err).NotTo(HaveOccurred())

			Expect(group.ID).NotTo(BeEmpty())
			Expect(group.DisplayName).To(Equal("banana.read"))
		})
	})
})
