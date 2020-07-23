package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum defaulting", func() {
	It("Should default network joining rinkeby", func() {
		network := &Network{
			Spec: NetworkSpec{
				Join:            "rinkeby",
				HighlyAvailable: true,
				Nodes: []Node{
					{
						Name: "node-1",
					},
				},
			},
		}
		network.Default()
		Expect(network.Spec.TopologyKey).To(Equal(DefaultTopologyKey))
		node := network.Spec.Nodes[0]
		Expect(node.P2PPort).To(Equal(DefaultP2PPort))
		Expect(node.SyncMode).To(Equal(DefaultPublicNetworkSyncMode))
		Expect(node.Client).To(Equal(DefaultClient))
	})

	It("Should default network with pow consensus", func() {
		network := &Network{
			Spec: NetworkSpec{
				Consensus: ProofOfWork,
				Genesis: &Genesis{
					ChainID: 55555,
				},
				Nodes: []Node{
					{
						Name: "node-1",
					},
				},
			},
		}
		network.Default()
		// node defaulting
		node := network.Spec.Nodes[0]
		var block0 uint = 0
		Expect(node.P2PPort).To(Equal(DefaultP2PPort))
		Expect(node.SyncMode).To(Equal(DefaultPrivateNetworkSyncMode))
		Expect(node.Client).To(Equal(DefaultClient))
		// genesis defaulting
		Expect(network.Spec.Genesis.Coinbase).To(Equal(DefaultCoinbase))
		Expect(network.Spec.Genesis.MixHash).To(Equal(DefaultMixHash))
		Expect(network.Spec.Genesis.Difficulty).To(Equal(DefaultDifficulty))
		Expect(network.Spec.Genesis.GasLimit).To(Equal(DefaultGasLimit))
		Expect(network.Spec.Genesis.Nonce).To(Equal(DefaultNonce))
		Expect(network.Spec.Genesis.Timestamp).To(Equal(DefaultTimestamp))
		// forks defaulting
		Expect(network.Spec.Genesis.Forks.Homestead).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.DAO).To(BeNil())
		Expect(network.Spec.Genesis.Forks.EIP150).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.EIP150Hash).To(Equal(DefaultEIP150Hash))
		Expect(network.Spec.Genesis.Forks.EIP155).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.EIP158).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.Byzantium).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.Constantinople).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.Petersburg).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.Istanbul).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.MuirGlacier).To(Equal(block0))
	})

	It("Should default network with poa consensus", func() {
		network := &Network{
			Spec: NetworkSpec{
				Consensus: ProofOfAuthority,
				Genesis: &Genesis{
					ChainID: 55555,
				},
				Nodes: []Node{
					{
						Name: "node-1",
						RPC:  true,
					},
				},
			},
		}
		network.Default()
		// node defaulting
		node := network.Spec.Nodes[0]
		var block0 uint = 0
		Expect(node.Client).To(Equal(DefaultClient))
		Expect(node.P2PPort).To(Equal(DefaultP2PPort))
		Expect(node.SyncMode).To(Equal(DefaultPrivateNetworkSyncMode))
		Expect(node.Hosts).To(Equal(DefaultOrigins))
		Expect(node.CORSDomains).To(Equal(DefaultOrigins))
		Expect(node.RPCHost).To(Equal(DefaultHost))
		Expect(node.RPCPort).To(Equal(DefaultRPCPort))
		Expect(node.RPCAPI).To(Equal(DefaultAPIs))
		// genesis defaulting
		Expect(network.Spec.Genesis.Coinbase).To(Equal(DefaultCoinbase))
		Expect(network.Spec.Genesis.MixHash).To(Equal(DefaultMixHash))
		Expect(network.Spec.Genesis.Difficulty).To(Equal(DefaultDifficulty))
		Expect(network.Spec.Genesis.GasLimit).To(Equal(DefaultGasLimit))
		Expect(network.Spec.Genesis.Nonce).To(Equal(DefaultNonce))
		Expect(network.Spec.Genesis.Timestamp).To(Equal(DefaultTimestamp))
		// forks defaulting
		Expect(network.Spec.Genesis.Forks.Homestead).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.DAO).To(BeNil())
		Expect(network.Spec.Genesis.Forks.EIP150).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.EIP150Hash).To(Equal(DefaultEIP150Hash))
		Expect(network.Spec.Genesis.Forks.EIP155).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.EIP158).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.Byzantium).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.Constantinople).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.Petersburg).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.Istanbul).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.MuirGlacier).To(Equal(block0))
		// clique defaulting
		Expect(network.Spec.Genesis.Clique.BlockPeriod).To(Equal(DefaultCliqueBlockPeriod))
		Expect(network.Spec.Genesis.Clique.EpochLength).To(Equal(DefaultCliqueEpochLength))
	})

	It("Should default network with ibft2 consensus", func() {
		network := &Network{
			Spec: NetworkSpec{
				Consensus: IstanbulBFT,
				Genesis: &Genesis{
					ChainID: 55555,
				},
				Nodes: []Node{
					{
						Name:    "node-1",
						WS:      true,
						GraphQL: true,
					},
				},
			},
		}
		network.Default()
		// node defaulting
		node := network.Spec.Nodes[0]
		var block0 uint = 0
		Expect(node.Client).To(Equal(DefaultClient))
		Expect(node.P2PPort).To(Equal(DefaultP2PPort))
		Expect(node.SyncMode).To(Equal(DefaultPrivateNetworkSyncMode))
		Expect(node.Hosts).To(Equal(DefaultOrigins))
		Expect(node.CORSDomains).To(Equal(DefaultOrigins))
		Expect(node.WSHost).To(Equal(DefaultHost))
		Expect(node.WSPort).To(Equal(DefaultWSPort))
		Expect(node.WSAPI).To(Equal(DefaultAPIs))
		Expect(node.GraphQLHost).To(Equal(DefaultHost))
		Expect(node.GraphQLPort).To(Equal(DefaultGraphQLPort))
		// genesis defaulting
		Expect(network.Spec.Genesis.Coinbase).To(Equal(DefaultCoinbase))
		Expect(network.Spec.Genesis.MixHash).To(Equal(DefaultMixHash))
		Expect(network.Spec.Genesis.Difficulty).To(Equal(DefaultDifficulty))
		Expect(network.Spec.Genesis.GasLimit).To(Equal(DefaultGasLimit))
		Expect(network.Spec.Genesis.Nonce).To(Equal(DefaultNonce))
		Expect(network.Spec.Genesis.Timestamp).To(Equal(DefaultTimestamp))
		// forks defaulting
		Expect(network.Spec.Genesis.Forks.Homestead).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.DAO).To(BeNil())
		Expect(network.Spec.Genesis.Forks.EIP150).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.EIP150Hash).To(Equal(DefaultEIP150Hash))
		Expect(network.Spec.Genesis.Forks.EIP155).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.EIP158).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.Byzantium).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.Constantinople).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.Petersburg).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.Istanbul).To(Equal(block0))
		Expect(network.Spec.Genesis.Forks.MuirGlacier).To(Equal(block0))
		// IBFT2 defaulting
		Expect(network.Spec.Genesis.IBFT2.BlockPeriod).To(Equal(DefaultIBFT2BlockPeriod))
		Expect(network.Spec.Genesis.IBFT2.EpochLength).To(Equal(DefaultIBFT2EpochLength))
		Expect(network.Spec.Genesis.IBFT2.RequestTimeout).To(Equal(DefaultIBFT2RequestTimeout))
		Expect(network.Spec.Genesis.IBFT2.MessageQueueLimit).To(Equal(DefaultIBFT2MessageQueueLimit))
		Expect(network.Spec.Genesis.IBFT2.DuplicateMessageLimit).To(Equal(DefaultIBFT2DuplicateMessageLimit))
		Expect(network.Spec.Genesis.IBFT2.FutureMessagesLimit).To(Equal(DefaultIBFT2FutureMessagesLimit))
		Expect(network.Spec.Genesis.IBFT2.FutureMessagesMaxDistance).To(Equal(DefaultIBFT2FutureMessagesMaxDistance))
	})
})
