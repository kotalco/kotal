package ethereum

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEthereumClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ethereum Clients Suite")
}
