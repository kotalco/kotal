package filecoin

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFilecoinClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Filecoin Clients Suite")
}
