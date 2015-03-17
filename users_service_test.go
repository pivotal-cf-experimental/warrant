package warrant_test

import (
	"time"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UsersService", func() {
	var service warrant.UsersService
	var token string

	BeforeEach(func() {
		service = warrant.NewUsersService(warrant.Config{
			Host:          fakeUAAServer.URL(),
			SkipVerifySSL: true,
		})
		token = fakeUAAServer.TokenFor([]string{"scim.write", "scim.read", "password.write"}, []string{"scim", "password"})
	})

	Describe("Create", func() {
		It("creates a new user", func() {
			user, err := service.Create("created-user", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())
			Expect(user.ID).NotTo(BeEmpty())
			Expect(user.UserName).To(Equal("created-user"))
			Expect(user.CreatedAt).To(BeTemporally("~", time.Now()))
			Expect(user.UpdatedAt).To(BeTemporally("~", time.Now()))
			Expect(user.Version).To(Equal(0))
			Expect(user.Emails).To(ConsistOf([]string{"user@example.com"}))
			Expect(user.Groups).To(ConsistOf([]warrant.Group{}))
			Expect(user.Active).To(BeTrue())
			Expect(user.Verified).To(BeFalse())
			Expect(user.Origin).To(Equal("uaa"))

			fetchedUser, err := service.Get(user.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedUser).To(Equal(user))
		})

		It("requires the scim.write scope", func() {
			token = fakeUAAServer.TokenFor([]string{"scim.banana"}, []string{"scim"})
			_, err := service.Create("created-user", "user@example.com", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})

		It("requires the scim audience", func() {
			token = fakeUAAServer.TokenFor([]string{"scim.write"}, []string{"banana"})
			_, err := service.Create("created-user", "user@example.com", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})

		It("requires an email address", func() {
			_, err := service.Create("created-user", "", token)
			Expect(err.Error()).To(Equal(`Warrant UnexpectedStatusError: 400 {"message":"[Assertion failed] - this String argument must have text; it must not be null, empty, or blank","error":"invalid_scim_resource"}`))
		})
	})

	Describe("Get", func() {
		var createdUser warrant.User

		BeforeEach(func() {
			var err error
			createdUser, err = service.Create("created-user", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the found user", func() {
			user, err := service.Get(createdUser.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(user).To(Equal(createdUser))
		})

		It("requires the scim.read scope", func() {
			token = fakeUAAServer.TokenFor([]string{"scim.banana"}, []string{"scim"})
			_, err := service.Get(createdUser.ID, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})

		It("requires the scim audience", func() {
			token = fakeUAAServer.TokenFor([]string{"scim.read"}, []string{"banana"})
			_, err := service.Get(createdUser.ID, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})
	})

	Describe("Delete", func() {
		var user warrant.User

		BeforeEach(func() {
			var err error
			user, err = service.Create("deleted-user", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("deletes the user", func() {
			err := service.Delete(user.ID, token)
			Expect(err).NotTo(HaveOccurred())

			_, err = service.Get(user.ID, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})

		It("requires the scim.write scope", func() {
			token = fakeUAAServer.TokenFor([]string{"scim.banana"}, []string{"scim"})
			err := service.Delete(user.ID, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})

		It("requires the scim audience", func() {
			token = fakeUAAServer.TokenFor([]string{"scim.write"}, []string{"banana"})
			err := service.Delete(user.ID, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})

		It("returns an error when the user does not exist", func() {
			err := service.Delete("non-existant-user-guid", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})
	})

	Describe("Update", func() {
		var user warrant.User

		BeforeEach(func() {
			var err error
			user, err = service.Create("new-user", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("updates an existing user", func() {
			user.UserName = "updated-user"
			updatedUser, err := service.Update(user, token)
			Expect(err).NotTo(HaveOccurred())

			fetchedUser, err := service.Get(user.ID, token)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedUser).To(Equal(updatedUser))
		})

		It("requires the scim.write scope", func() {
			token = fakeUAAServer.TokenFor([]string{"scim.banana"}, []string{"scim"})
			_, err := service.Update(user, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})

		It("requires the scim audience", func() {
			token = fakeUAAServer.TokenFor([]string{"scim.write"}, []string{"banana"})
			_, err := service.Update(user, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})

		It("must match the 'If-Match' header value", func() {
			user.Version = 24
			_, err := service.Update(user, token)
			Expect(err).To(MatchError(`Warrant UnexpectedStatusError: 400 {"message":"Missing If-Match for PUT","error":"scim"}`))
		})

		It("returns an error if the user does not exist", func() {
			user.ID = "non-existant-guid"
			_, err := service.Update(user, token)
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})
	})

	Describe("SetPassword", func() {
		var user warrant.User

		BeforeEach(func() {
			var err error
			user, err = service.Create("password-user", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("sets the password belonging to the given user guid", func() {
			err := service.SetPassword(user.ID, "password", token)
			Expect(err).NotTo(HaveOccurred())
		})

		It("requires the password.write scope", func() {
			token = fakeUAAServer.TokenFor([]string{"password.banana"}, []string{"password"})
			err := service.SetPassword(user.ID, "password", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})

		It("requires the password audience", func() {
			token = fakeUAAServer.TokenFor([]string{"password.write"}, []string{"banana"})
			err := service.SetPassword(user.ID, "password", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})

		It("returns an error if the user does not exist", func() {
			err := service.SetPassword("non-existant-guid", "password", token)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})
	})

	Describe("ChangePassword", func() {
		var (
			user      warrant.User
			userToken string
		)

		BeforeEach(func() {
			var err error
			user, err = service.Create("change-password-user", "user@example.com", token)
			Expect(err).NotTo(HaveOccurred())

			err = service.SetPassword(user.ID, "old-password", token)
			Expect(err).NotTo(HaveOccurred())

			userToken = fakeUAAServer.UserTokenFor(user.ID, []string{}, []string{})
		})

		It("changes the password given the old password", func() {
			err := service.ChangePassword(user.ID, "old-password", "new-password", userToken)
			Expect(err).NotTo(HaveOccurred())
		})

		It("does not change password if the old password does not match", func() {
			err := service.ChangePassword(user.ID, "bad-password", "new-password", userToken)
			Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
		})

		Context("with a client token", func() {
			It("changes the password regardless of the old password", func() {
				err := service.ChangePassword(user.ID, "bad-password", "new-password", token)
				Expect(err).NotTo(HaveOccurred())
			})

			It("requires the password.write scope", func() {
				token = fakeUAAServer.TokenFor([]string{"password.banana"}, []string{"password"})
				err := service.ChangePassword(user.ID, "old-password", "new-password", token)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})

			It("requires the password audience", func() {
				token = fakeUAAServer.TokenFor([]string{"password.write"}, []string{"banana"})
				err := service.ChangePassword(user.ID, "old-password", "new-password", token)
				Expect(err).To(BeAssignableToTypeOf(warrant.UnauthorizedError{}))
			})
		})
	})

	Describe("ScorePassword", func() {
		PIt("returns a score value for the given password", func() {
			//score, requiredScore, err := service.ScorePassword("d9327654lhuf")
			//Expect(err).NotTo(HaveOccurred())
			//Expect(score).To(Equal(8))
		})
	})
})
