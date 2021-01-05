package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum 2.0 node defaulting", func() {

	It("Should default node with missing client", func() {
		node := Node{
			Spec: NodeSpec{
				Join: "mainnet",
			},
		}
		node.Default()
		Expect(node.Spec.Client).To(Equal(DefaultClient))
	})

	It("Should default node with missing client and rest port", func() {
		node := Node{
			Spec: NodeSpec{
				Join: "mainnet",
				REST: true,
			},
		}
		node.Default()
		Expect(node.Spec.Client).To(Equal(DefaultClient))
		Expect(node.Spec.RESTPort).To(Equal(DefaultRestPort))
	})

	It("Should default node with missing rpc port", func() {
		node := Node{
			Spec: NodeSpec{
				Client: PrysmClient,
				Join:   "mainnet",
				RPC:    true,
			},
		}
		node.Default()
		Expect(node.Spec.RPCPort).To(Equal(DefaultRPCPort))
	})

})
