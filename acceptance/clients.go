package acceptance

import (
	"os/exec"
	"time"

	"github.com/pivotal-cf-experimental/warrant"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client Lifecycle", func() {
	var (
		warrantClient warrant.Warrant
		client        warrant.Client
	)

	BeforeEach(func() {
		client = warrant.Client{
			ID:                   "warrant-client",
			Scope:                []string{"openid"},
			ResourceIDs:          []string{"none"},
			Authorities:          []string{"scim.read", "scim.write"},
			AuthorizedGrantTypes: []string{"client_credentials"},
			AccessTokenValidity:  5000 * time.Second,
		}

		warrantClient = warrant.New(warrant.Config{
			Host:          UAAHost,
			SkipVerifySSL: true,
			TraceWriter:   TraceWriter,
		})
	})

	AfterEach(func() {
		// TODO: replace with implementation that does not call out to UAAC
		cmd := exec.Command("uaac", "client", "delete", client.ID)
		output, err := cmd.Output()
		Expect(err).NotTo(HaveOccurred())
		Expect(output).To(ContainSubstring("client registration deleted"))
	})

	It("creates, and retrieves a client", func() {
		By("creating a client", func() {
			err := warrantClient.Clients.Create(client, "secret", UAAToken)
			Expect(err).NotTo(HaveOccurred())
		})

		By("find the client", func() {
			fetchedClient, err := warrantClient.Clients.Get(client.ID, UAAToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(fetchedClient).To(Equal(client))
		})
	})
})
