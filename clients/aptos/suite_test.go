package aptos

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAptosCoreClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aptos Core Client Suite")
}
