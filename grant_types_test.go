package warrant_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/warrant"
)

var _ = Describe("GrantTypes", func() {
	Describe("GenerateForm", func() {
		Context("when only client credentials are provided", func() {
			It("returns the correct url form values", func() {
				form := warrant.GrantTypes{
					ClientCredentials: warrant.ClientCredentials{
						ID:     "some-id",
						Secret: "some-secret",
					},
				}.GenerateForm()

				Expect(form.Get("response_type")).To(Equal("token"))
				Expect(form.Get("client_id")).To(Equal("some-id"))
				Expect(form.Get("client_secret")).To(Equal("some-secret"))
				Expect(form.Get("grant_type")).To(Equal("client_credentials"))
			})
		})
	})
})
