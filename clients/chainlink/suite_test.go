package chainlink

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestChainlinkClients(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Chainlink Client Suite")
}
