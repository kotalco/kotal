package stacks

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestStacksNodeClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Stacks Node Client Suite")
}
