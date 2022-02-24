package bitcoin

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBitcoinCoreClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Bitcoin Core Client Suite")
}
