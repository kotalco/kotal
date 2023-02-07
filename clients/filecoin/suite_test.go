package filecoin

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFilecoinClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Filecoin Clients Suite")
}
