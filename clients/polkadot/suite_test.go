package polkadot

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestEthereum2Client(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Polkadot Client Suite")
}
