package near

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEthereum2Client(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "NEAR Core Client Suite")
}
