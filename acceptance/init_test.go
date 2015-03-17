package acceptance

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAcceptanceSuite(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Acceptance Suite")
}
