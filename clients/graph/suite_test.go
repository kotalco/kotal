package graph

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGraphNodeClient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Graph Node Client Suite")
}
