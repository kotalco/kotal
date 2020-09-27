package v1alpha1

import "sigs.k8s.io/controller-runtime/pkg/webhook"

// +kubebuilder:webhook:path=/mutate-ethereum-kotal-io-v1alpha1-network,mutating=true,failurePolicy=fail,groups=ethereum.kotal.io,resources=networks,verbs=create;update,versions=v1alpha1,name=mnetwork.kb.io

var _ webhook.Defaulter = &Network{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Network) Default() {
	networklog.Info("default", "name", r.Name)

	if r.Spec.HighlyAvailable {
		if r.Spec.TopologyKey == "" {
			r.Spec.TopologyKey = DefaultTopologyKey
		}
	}

	// default genesis block
	if r.Spec.Genesis != nil {
		r.DefaultGenesis()
	}

	// default network nodes
	for i := range r.Spec.Nodes {
		r.DefaultNode(&r.Spec.Nodes[i])
	}

}

// DefaultNodeResources defaults node cpu, memory and storage resources
func (r *Network) DefaultNodeResources(node *Node) {
	var cpu, cpuLimit, memory, memoryLimit, storage string
	privateNetwork := r.Spec.Genesis != nil
	join := r.Spec.Join

	if node.Resources == nil {
		node.Resources = &NodeResources{}
	}

	if node.Resources.CPU == "" {
		if privateNetwork {
			cpu = DefaultPrivateNetworkNodeCPURequest
		} else {
			cpu = DefaultPublicNetworkNodeCPURequest
		}

		node.Resources.CPU = cpu
	}

	if node.Resources.CPULimit == "" {
		if privateNetwork {
			cpuLimit = DefaultPrivateNetworkNodeCPULimit
		} else {
			cpuLimit = DefaultPublicNetworkNodeCPULimit
		}

		node.Resources.CPULimit = cpuLimit
	}

	if node.Resources.Memory == "" {
		if privateNetwork {
			memory = DefaultPrivateNetworkNodeMemoryRequest
		} else {
			memory = DefaultPublicNetworkNodeMemoryRequest
		}

		node.Resources.Memory = memory
	}

	if node.Resources.MemoryLimit == "" {
		if privateNetwork {
			memoryLimit = DefaultPrivateNetworkNodeMemoryLimit
		} else {
			memoryLimit = DefaultPublicNetworkNodeMemoryLimit
		}

		node.Resources.MemoryLimit = memoryLimit
	}

	if node.Resources.Storage == "" {
		if privateNetwork {
			storage = DefaultPrivateNetworkNodeStorageRequest
		} else if join == MainNetwork && node.SyncMode == FastSynchronization {
			storage = DefaultMainNetworkFastNodeStorageRequest
		} else if join == MainNetwork && node.SyncMode == FullSynchronization {
			storage = DefaultMainNetworkFullNodeStorageRequest
		} else {
			storage = DefaultTestNetworkStorageRequest
		}

		node.Resources.Storage = storage
	}

}

// DefaultNode defaults a single node
func (r *Network) DefaultNode(node *Node) {
	defaultAPIs := []API{Web3API, ETHAPI, NetworkAPI}

	if node.Client == "" {
		node.Client = DefaultClient
	}

	if node.P2PPort == 0 {
		node.P2PPort = DefaultP2PPort
	}

	if node.SyncMode == "" {
		// public network
		if r.Spec.Genesis == nil {
			node.SyncMode = FastSynchronization
		} else {
			node.SyncMode = FullSynchronization
		}
	}

	// must be called after defaulting sync mode because it's depending on its value
	r.DefaultNodeResources(node)

	if node.RPC || node.WS || node.GraphQL {
		if len(node.Hosts) == 0 {
			node.Hosts = DefaultOrigins
		}

		if len(node.CORSDomains) == 0 {
			node.CORSDomains = DefaultOrigins
		}
	}

	if node.RPC {
		if node.RPCPort == 0 {
			node.RPCPort = 8545
		}

		if len(node.RPCAPI) == 0 {
			node.RPCAPI = defaultAPIs
		}
	}

	if node.WS {
		if node.WSPort == 0 {
			node.WSPort = DefaultWSPort
		}

		if len(node.WSAPI) == 0 {
			node.WSAPI = defaultAPIs
		}
	}

	if node.GraphQL {
		if node.GraphQLPort == 0 {
			node.GraphQLPort = DefaultGraphQLPort
		}
	}

	if node.Logging == "" {
		node.Logging = DefaultLogging
	}

}

// DefaultGenesis defaults genesis block parameters
func (r *Network) DefaultGenesis() {
	if r.Spec.Genesis.Coinbase == "" {
		r.Spec.Genesis.Coinbase = DefaultCoinbase
	}

	if r.Spec.Genesis.Difficulty == "" {
		r.Spec.Genesis.Difficulty = DefaultDifficulty
	}

	if r.Spec.Genesis.Forks == nil {
		// all milestones will be activated at block 0
		r.Spec.Genesis.Forks = &Forks{}
	}

	if r.Spec.Genesis.Forks.EIP150Hash == "" {
		r.Spec.Genesis.Forks.EIP150Hash = DefaultEIP150Hash
	}

	if r.Spec.Genesis.MixHash == "" {
		r.Spec.Genesis.MixHash = DefaultMixHash
	}

	if r.Spec.Genesis.GasLimit == "" {
		r.Spec.Genesis.GasLimit = DefaultGasLimit
	}

	if r.Spec.Genesis.Nonce == "" {
		r.Spec.Genesis.Nonce = DefaultNonce
	}

	if r.Spec.Genesis.Timestamp == "" {
		r.Spec.Genesis.Timestamp = DefaultTimestamp
	}

	if r.Spec.Consensus == ProofOfWork {
		if r.Spec.Genesis.Ethash == nil {
			r.Spec.Genesis.Ethash = &Ethash{}
		}
	}

	if r.Spec.Consensus == ProofOfAuthority {
		if r.Spec.Genesis.Clique == nil {
			r.Spec.Genesis.Clique = &Clique{}
		}
		if r.Spec.Genesis.Clique.BlockPeriod == 0 {
			r.Spec.Genesis.Clique.BlockPeriod = DefaultCliqueBlockPeriod
		}
		if r.Spec.Genesis.Clique.EpochLength == 0 {
			r.Spec.Genesis.Clique.EpochLength = DefaultCliqueEpochLength
		}
	}

	if r.Spec.Consensus == IstanbulBFT {
		if r.Spec.Genesis.IBFT2 == nil {
			r.Spec.Genesis.IBFT2 = &IBFT2{}
		}
		if r.Spec.Genesis.IBFT2.BlockPeriod == 0 {
			r.Spec.Genesis.IBFT2.BlockPeriod = DefaultIBFT2BlockPeriod
		}
		if r.Spec.Genesis.IBFT2.EpochLength == 0 {
			r.Spec.Genesis.IBFT2.EpochLength = DefaultIBFT2EpochLength
		}
		if r.Spec.Genesis.IBFT2.RequestTimeout == 0 {
			r.Spec.Genesis.IBFT2.RequestTimeout = DefaultIBFT2RequestTimeout
		}
		if r.Spec.Genesis.IBFT2.MessageQueueLimit == 0 {
			r.Spec.Genesis.IBFT2.MessageQueueLimit = DefaultIBFT2MessageQueueLimit
		}
		if r.Spec.Genesis.IBFT2.DuplicateMessageLimit == 0 {
			r.Spec.Genesis.IBFT2.DuplicateMessageLimit = DefaultIBFT2DuplicateMessageLimit
		}
		if r.Spec.Genesis.IBFT2.FutureMessagesLimit == 0 {
			r.Spec.Genesis.IBFT2.FutureMessagesLimit = DefaultIBFT2FutureMessagesLimit
		}
		if r.Spec.Genesis.IBFT2.FutureMessagesMaxDistance == 0 {
			r.Spec.Genesis.IBFT2.FutureMessagesMaxDistance = DefaultIBFT2FutureMessagesMaxDistance
		}

	}
}
