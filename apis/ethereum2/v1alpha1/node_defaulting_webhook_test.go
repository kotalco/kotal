package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum 2.0 node defaulting", func() {

	It("Should default node with missing client and p2p port", func() {
		node := Node{
			Spec: NodeSpec{
				Join: "mainnet",
			},
		}
		node.Default()
		Expect(node.Spec.Client).To(Equal(DefaultClient))
		Expect(node.Spec.P2PPort).To(Equal(DefaultP2PPort))
	})

	It("Should default node with missing client and rest port/host", func() {
		node := Node{
			Spec: NodeSpec{
				Join: "mainnet",
				REST: true,
			},
		}
		node.Default()
		Expect(node.Spec.Client).To(Equal(DefaultClient))
		Expect(node.Spec.RESTPort).To(Equal(DefaultRestPort))
		Expect(node.Spec.RESTHost).To(Equal(DefaultRestHost))
	})

	It("Should default node with missing rpc port and host", func() {
		node := Node{
			Spec: NodeSpec{
				Client: PrysmClient,
				Join:   "mainnet",
				RPC:    true,
			},
		}
		node.Default()
		Expect(node.Spec.RPCPort).To(Equal(DefaultRPCPort))
		Expect(node.Spec.RPCHost).To(Equal(DefaultRPCHost))
	})

	It("Should default node with missing grpc port", func() {
		node := Node{
			Spec: NodeSpec{
				Client: PrysmClient,
				Join:   "mainnet",
				GRPC:   true,
			},
		}
		node.Default()
		Expect(node.Spec.GRPCPort).To(Equal(DefaultGRPCPort))
	})

	It("Should default node with missing grpc host", func() {
		node := Node{
			Spec: NodeSpec{
				Client: PrysmClient,
				Join:   "mainnet",
				GRPC:   true,
			},
		}
		node.Default()
		Expect(node.Spec.GRPCHost).To(Equal(DefaultGRPCHost))
	})

})
