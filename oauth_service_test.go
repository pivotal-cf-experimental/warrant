package warrant_test

import (
	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("OAuthService", func() {
	var (
		service warrant.OAuthService
		token   string
		user    warrant.User
		config  warrant.Config
	)

	BeforeEach(func() {
		config = warrant.Config{
			Host:          fakeUAAServer.URL(),
			SkipVerifySSL: true,
		}
		service = warrant.NewOAuthService(config)
		token = fakeUAAServer.TokenFor([]string{"scim.write", "scim.read", "password.write"}, []string{"scim", "password"})

		usersService := warrant.NewUsersService(config)
		var err error
		user, err = usersService.Create("username", "user@example.com", token)
		Expect(err).NotTo(HaveOccurred())

		err = usersService.SetPassword(user.ID, "password", token)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("GetUserToken", func() {
		It("returns a valid token given a username and password", func() {
			token, err := service.GetUserToken("username", "password", "cf", "https://cf.example.com/redirect")
			Expect(err).NotTo(HaveOccurred())
			Expect(token).NotTo(BeEmpty())

			decodedToken := fakeUAAServer.Tokenizer.Decrypt(token)
			Expect(decodedToken.UserID).To(Equal(user.ID))
		})

		It("returns an error when the request does not succeed", func() {
			_, err := service.GetUserToken("unknown-user", "password", "cf", "https://cf.example.com/redirect")
			Expect(err).To(BeAssignableToTypeOf(warrant.NotFoundError{}))
		})

		It("returns an error when the response is not parsable", func() {
			_, err := service.GetUserToken("username", "password", "cf", "%%%")
			Expect(err).To(MatchError(`parse %%%: invalid URL escape "%%%"`))
		})
	})
})
