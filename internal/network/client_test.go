package network_test

import (
	"net/http"
	"reflect"
	"time"

	"github.com/pivotal-cf-experimental/warrant/internal/network"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("using a global HTTP client", func() {
	It("retrieves the exact same client reference for a given value of config.SkipVerifySSL", func() {
		client1 := network.GetClient(network.Config{SkipVerifySSL: true})
		client2 := network.GetClient(network.Config{SkipVerifySSL: true})

		transportPointer1 := reflect.ValueOf(client1.Transport).Pointer()
		transportPointer2 := reflect.ValueOf(client2.Transport).Pointer()
		Expect(transportPointer1).To(Equal(transportPointer2))
	})

	It("retrieves difference client references for different values of config.SkipVerifySSL", func() {
		client1 := network.GetClient(network.Config{SkipVerifySSL: true})
		client2 := network.GetClient(network.Config{SkipVerifySSL: false})

		transportPointer1 := reflect.ValueOf(client1.Transport).Pointer()
		transportPointer2 := reflect.ValueOf(client2.Transport).Pointer()
		Expect(transportPointer1).NotTo(Equal(transportPointer2))
	})

	It("uses the configuration to configure the HTTP client", func() {
		config := network.Config{
			SkipVerifySSL: true,
		}
		transport := network.GetClient(config).Transport.(*http.Transport)

		Expect(transport.TLSClientConfig.InsecureSkipVerify).To(BeTrue())
		Expect(reflect.ValueOf(transport.Proxy).Pointer()).To(Equal(reflect.ValueOf(http.ProxyFromEnvironment).Pointer()))
		Expect(transport.TLSHandshakeTimeout).To(Equal(10 * time.Second))
	})
})
