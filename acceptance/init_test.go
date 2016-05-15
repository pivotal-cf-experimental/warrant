package acceptance

import (
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/warrant"
)

const (
	UAADefaultUsername = "warrant-user"
	UAADefaultClientID = "warrant-client"
)

var (
	UAAHost        string
	UAAAdminClient string
	UAAAdminSecret string
	UAAToken       string
	TraceWriter    io.Writer
)

func TestAcceptanceSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "acceptance")
}

var _ = BeforeSuite(func() {
	if os.Getenv("TRACE") == "true" {
		TraceWriter = os.Stdout
	}

	UAAHost = fmt.Sprintf("https://%s", os.Getenv("UAA_HOST"))
	UAAAdminClient = os.Getenv("UAA_ADMIN_CLIENT")
	UAAAdminSecret = os.Getenv("UAA_ADMIN_SECRET")

	switch {
	case UAAHost == "":
		Fail("UAA_HOST is a required environment variable")
	case UAAAdminClient == "":
		Fail("UAA_ADMIN_CLIENT is a required enviroment variable")
	case UAAAdminSecret == "":
		Fail("UAA_ADMIN_SECRET is a required envrionment variable")
	}

	warrantClient := warrant.New(warrant.Config{
		Host:          UAAHost,
		SkipVerifySSL: true,
		TraceWriter:   TraceWriter,
	})

	var err error
	UAAToken, err = warrantClient.Clients.GetToken(UAAAdminClient, UAAAdminSecret)
	Expect(err).NotTo(HaveOccurred())
	//Add more power to the client
	adminClient, err := warrantClient.Clients.Get(UAAAdminClient, UAAToken)
	Expect(err).NotTo(HaveOccurred())

	adminClient.Authorities = append(adminClient.Authorities, "password.write")
	adminClient.Authorities = append(adminClient.Authorities, "uaa.resource")
	adminClient.AuthorizedGrantTypes = append(adminClient.AuthorizedGrantTypes, "password")
	adminClient.AccessTokenValidity = time.Duration(2) * time.Hour
	err = warrantClient.Clients.Update(adminClient, UAAToken)
	Expect(err).NotTo(HaveOccurred())

})
