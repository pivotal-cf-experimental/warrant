package warrant_test

import (
	"io"
	"os"
	"testing"

	"github.com/pivotal-cf-experimental/warrant/testserver"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	fakeUAA     *testserver.UAA
	TraceWriter io.Writer
)

func TestWarrantSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "warrant")
}

var _ = BeforeSuite(func() {
	if os.Getenv("TRACE") == "true" {
		TraceWriter = os.Stdout
	}

	fakeUAA = testserver.NewUAA()
	fakeUAA.Start()
})

var _ = AfterSuite(func() {
	fakeUAA.Close()
})

var _ = BeforeEach(func() {
	fakeUAA.Reset()
})
