package ipfs

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIPFSClients(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "IPFS Clients Suite")
}
