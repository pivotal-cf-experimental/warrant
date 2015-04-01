package acceptance

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/warrant"
)

var (
	UAAHost        string
	UAAAdminClient string
	UAAAdminSecret string
	UAAToken       string
)

func TestAcceptanceSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Suite")
}

var _ = BeforeSuite(func() {
	UAAHost = os.Getenv("UAA_HOST")
	UAAAdminClient = os.Getenv("UAA_ADMIN_CLIENT")
	UAAAdminSecret = os.Getenv("UAA_ADMIN_SECRET")

	warrantClient := warrant.New(warrant.Config{
		Host:          UAAHost,
		SkipVerifySSL: true,
	})

	var err error
	UAAToken, err = warrantClient.Clients.GetToken(UAAAdminClient, UAAAdminSecret)
	Expect(err).NotTo(HaveOccurred())
})
