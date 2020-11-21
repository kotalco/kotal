package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum defaulting", func() {
	It("Should default network joining mainnet", func() {
		network := &Network{
			Spec: NetworkSpec{
				Join:            MainNetwork,
				HighlyAvailable: true,
				Nodes: []XNode{
					{
						Name: "node-1",
					},
					{
						Name:     "node-2",
						SyncMode: FullSynchronization,
					},
				},
			},
		}
		network.Default()
		Expect(network.Spec.TopologyKey).To(Equal(DefaultTopologyKey))
		node1 := network.Spec.Nodes[0]
		node2 := network.Spec.Nodes[1]
		// node1 defaulting
		Expect(node1.P2PPort).To(Equal(DefaultP2PPort))
		Expect(node1.SyncMode).To(Equal(DefaultPublicNetworkSyncMode))
		Expect(node1.Client).To(Equal(DefaultClient))
		Expect(node1.Resources.CPU).To(Equal(DefaultPublicNetworkNodeCPURequest))
		Expect(node1.Resources.CPULimit).To(Equal(DefaultPublicNetworkNodeCPULimit))
		Expect(node1.Resources.Memory).To(Equal(DefaultPublicNetworkNodeMemoryRequest))
		Expect(node1.Resources.MemoryLimit).To(Equal(DefaultPublicNetworkNodeMemoryLimit))
		Expect(node1.Resources.Storage).To(Equal(DefaultMainNetworkFastNodeStorageRequest))
		Expect(node1.Logging).To(Equal(DefaultLogging))
		// node2 defaulting
		Expect(node2.P2PPort).To(Equal(DefaultP2PPort))
		Expect(node2.SyncMode).To(Equal(FullSynchronization))
		Expect(node2.Client).To(Equal(DefaultClient))
		Expect(node2.Resources.CPU).To(Equal(DefaultPublicNetworkNodeCPURequest))
		Expect(node2.Resources.CPULimit).To(Equal(DefaultPublicNetworkNodeCPULimit))
		Expect(node2.Resources.Memory).To(Equal(DefaultPublicNetworkNodeMemoryRequest))
		Expect(node2.Resources.MemoryLimit).To(Equal(DefaultPublicNetworkNodeMemoryLimit))
		Expect(node2.Resources.Storage).To(Equal(DefaultMainNetworkFullNodeStorageRequest))
		Expect(node2.Logging).To(Equal(DefaultLogging))

	})

	It("Should default network joining rinkeby", func() {
		network := &Network{
			Spec: NetworkSpec{
				Join:            RinkebyNetwork,
				HighlyAvailable: true,
				Nodes: []XNode{
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
		Expect(node.Resources.CPU).To(Equal(DefaultPublicNetworkNodeCPURequest))
		Expect(node.Resources.CPULimit).To(Equal(DefaultPublicNetworkNodeCPULimit))
		Expect(node.Resources.Memory).To(Equal(DefaultPublicNetworkNodeMemoryRequest))
		Expect(node.Resources.MemoryLimit).To(Equal(DefaultPublicNetworkNodeMemoryLimit))
		Expect(node.Resources.Storage).To(Equal(DefaultTestNetworkStorageRequest))
		Expect(node.Logging).To(Equal(DefaultLogging))
	})

	It("Should default network with pow consensus", func() {
		network := &Network{
			Spec: NetworkSpec{
				Consensus: ProofOfWork,
				Genesis: &Genesis{
					ChainID: 55555,
				},
				Nodes: []XNode{
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
		Expect(node.Resources.CPU).To(Equal(DefaultPrivateNetworkNodeCPURequest))
		Expect(node.Resources.CPULimit).To(Equal(DefaultPrivateNetworkNodeCPULimit))
		Expect(node.Resources.Memory).To(Equal(DefaultPrivateNetworkNodeMemoryRequest))
		Expect(node.Resources.MemoryLimit).To(Equal(DefaultPrivateNetworkNodeMemoryLimit))
		Expect(node.Resources.Storage).To(Equal(DefaultPrivateNetworkNodeStorageRequest))
		Expect(node.Logging).To(Equal(DefaultLogging))
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
				Nodes: []XNode{
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
		Expect(node.RPCPort).To(Equal(DefaultRPCPort))
		Expect(node.RPCAPI).To(Equal(DefaultAPIs))
		Expect(node.Resources.CPU).To(Equal(DefaultPrivateNetworkNodeCPURequest))
		Expect(node.Resources.CPULimit).To(Equal(DefaultPrivateNetworkNodeCPULimit))
		Expect(node.Resources.Memory).To(Equal(DefaultPrivateNetworkNodeMemoryRequest))
		Expect(node.Resources.MemoryLimit).To(Equal(DefaultPrivateNetworkNodeMemoryLimit))
		Expect(node.Resources.Storage).To(Equal(DefaultPrivateNetworkNodeStorageRequest))
		Expect(node.Logging).To(Equal(DefaultLogging))
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
				Nodes: []XNode{
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
		Expect(node.WSPort).To(Equal(DefaultWSPort))
		Expect(node.WSAPI).To(Equal(DefaultAPIs))
		Expect(node.GraphQLPort).To(Equal(DefaultGraphQLPort))
		Expect(node.Resources.CPU).To(Equal(DefaultPrivateNetworkNodeCPURequest))
		Expect(node.Resources.CPULimit).To(Equal(DefaultPrivateNetworkNodeCPULimit))
		Expect(node.Resources.Memory).To(Equal(DefaultPrivateNetworkNodeMemoryRequest))
		Expect(node.Resources.MemoryLimit).To(Equal(DefaultPrivateNetworkNodeMemoryLimit))
		Expect(node.Resources.Storage).To(Equal(DefaultPrivateNetworkNodeStorageRequest))
		Expect(node.Logging).To(Equal(DefaultLogging))
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
