package acceptance

import (
	"fmt"

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

	It("creates, lists, retrieves, and deletes a group", func() {
		By("creating a new group", func() {
			group, err = ensureGroupExists(client, "banana.read", UAAToken)
			Expect(err).NotTo(HaveOccurred())

			Expect(group.ID).NotTo(BeEmpty())
			Expect(group.DisplayName).To(Equal("banana.read"))
		})

		By("listing the groups", func() {
			groups, err := client.Groups.List(warrant.Query{}, UAAToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(groups).To(ContainElement(group))
		})

		By("getting the created group", func() {
			group, err = client.Groups.Get(group.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())

			Expect(group.ID).NotTo(BeEmpty())
			Expect(group.DisplayName).To(Equal("banana.read"))
		})
	})

	Describe("Membership", func() {
		It("adds, lists, and deletes members from a group", func() {
			By("creating a group")
			group, err = ensureGroupExists(client, "banana.read", UAAToken)
			Expect(err).NotTo(HaveOccurred())

			By("creating a user")
			user, err := ensureUserExists(client, "a-user", "email", UAAToken)
			Expect(err).NotTo(HaveOccurred())

			By("adding a member")
			client.Groups.AddMember(group.ID, user.ID, UAAToken)

			By("listing the members")
			members, err := client.Groups.ListMembers(group.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(1))
			Expect(members[0].ID).To(Equal(user.ID))

			By("deleting the members")
			err = client.Groups.RemoveMember(group.ID, user.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())

			By("checking the group does not have the deleted member")
			members, err = client.Groups.ListMembers(group.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(members).To(HaveLen(0))
		})
	})
})

func ensureUserExists(client warrant.Warrant, username, email, token string) (warrant.User, error) {
	allUsers, err := client.Users.List(warrant.Query{
		Filter: fmt.Sprintf(`userName Eq %q`, username),
	}, token)
	Expect(err).NotTo(HaveOccurred())

	var user warrant.User
	if len(allUsers) == 1 {
		user = allUsers[0]
	} else {
		user, err = client.Users.Create(username, email, token)
		Expect(err).NotTo(HaveOccurred())
	}

	return user, nil
}

func ensureGroupExists(client warrant.Warrant, displayName, token string) (warrant.Group, error) {
	allGroups, err := client.Groups.List(warrant.Query{
		Filter: fmt.Sprintf(`displayName Eq %q`, displayName),
	}, token)
	Expect(err).NotTo(HaveOccurred())

	var group warrant.Group
	if len(allGroups) == 1 {
		group = allGroups[0]
	} else {
		group, err = client.Groups.Create(displayName, token)
		Expect(err).NotTo(HaveOccurred())
	}

	return group, nil
}
