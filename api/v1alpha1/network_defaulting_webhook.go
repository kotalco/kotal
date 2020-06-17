package v1alpha1

import "sigs.k8s.io/controller-runtime/pkg/webhook"

// +kubebuilder:webhook:path=/mutate-ethereum-kotal-io-v1alpha1-network,mutating=true,failurePolicy=fail,groups=ethereum.kotal.io,resources=networks,verbs=create;update,versions=v1alpha1,name=mnetwork.kb.io

var _ webhook.Defaulter = &Network{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Network) Default() {
	networklog.Info("default", "name", r.Name)

	// default genesis block
	if r.Spec.Genesis != nil {
		r.DefaultGenesis()
	}

	// default network nodes
	for i := range r.Spec.Nodes {
		r.DefaultNode(&r.Spec.Nodes[i])
	}

}

// DefaultNode defaults a single node
func (r *Network) DefaultNode(node *Node) {
	defaultAPIs := []API{Web3API, ETHAPI, NetworkAPI}
	anyAddress := "0.0.0.0"
	allOrigins := []string{"*"}

	if node.P2PPort == 0 {
		node.P2PPort = 30303
	}

	if node.SyncMode == "" {
		node.SyncMode = FullSynchronization
	}

	if node.RPC || node.WS || node.GraphQL {
		if len(node.Hosts) == 0 {
			node.Hosts = allOrigins
		}

		if len(node.CORSDomains) == 0 {
			node.CORSDomains = allOrigins
		}
	}

	if node.RPC {
		if node.RPCHost == "" {
			node.RPCHost = anyAddress
		}

		if node.RPCPort == 0 {
			node.RPCPort = 8545
		}

		if len(node.RPCAPI) == 0 {
			node.RPCAPI = defaultAPIs
		}
	}

	if node.WS {
		if node.WSHost == "" {
			node.WSHost = anyAddress
		}

		if node.WSPort == 0 {
			node.WSPort = 8546
		}

		if len(node.WSAPI) == 0 {
			node.WSAPI = defaultAPIs
		}
	}

	if node.GraphQL {
		if node.GraphQLHost == "" {
			node.GraphQLHost = anyAddress
		}

		if node.GraphQLPort == 0 {
			node.GraphQLPort = 8547
		}
	}

}

// DefaultGenesis defaults genesis block parameters
func (r *Network) DefaultGenesis() {
	if r.Spec.Genesis.Coinbase == "" {
		r.Spec.Genesis.Coinbase = "0x0000000000000000000000000000000000000000"
	}

	if r.Spec.Genesis.Difficulty == "" {
		r.Spec.Genesis.Difficulty = "0x1"
	}

	if r.Spec.Genesis.Forks == nil {
		// all milestones will be activated at block 0
		r.Spec.Genesis.Forks = &Forks{
			EIP150Hash: "0x2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f0",
		}
	}

	if r.Spec.Genesis.MixHash == "" {
		r.Spec.Genesis.MixHash = "0x0000000000000000000000000000000000000000000000000000000000000000"
	}

	if r.Spec.Genesis.GasLimit == "" {
		r.Spec.Genesis.GasLimit = "0x47b760"
	}

	if r.Spec.Genesis.Nonce == "" {
		r.Spec.Genesis.Nonce = "0x0"
	}

	if r.Spec.Genesis.Timestamp == "" {
		r.Spec.Genesis.Timestamp = "0x0"
	}

	if r.Spec.Consensus == ProofOfAuthority {
		if r.Spec.Genesis.Clique.BlockPeriod == 0 {
			r.Spec.Genesis.Clique.BlockPeriod = 15
		}
		if r.Spec.Genesis.Clique.EpochLength == 0 {
			r.Spec.Genesis.Clique.EpochLength = 3000
		}
	}

	if r.Spec.Consensus == IstanbulBFT {
		if r.Spec.Genesis.IBFT2.BlockPeriod == 0 {
			r.Spec.Genesis.IBFT2.BlockPeriod = 15
		}
		if r.Spec.Genesis.IBFT2.EpochLength == 0 {
			r.Spec.Genesis.IBFT2.EpochLength = 3000
		}
		if r.Spec.Genesis.IBFT2.RequestTimeout == 0 {
			r.Spec.Genesis.IBFT2.RequestTimeout = 10
		}
		if r.Spec.Genesis.IBFT2.MessageQueueLimit == 0 {
			r.Spec.Genesis.IBFT2.MessageQueueLimit = 1000
		}
		if r.Spec.Genesis.IBFT2.DuplicateMesageLimit == 0 {
			r.Spec.Genesis.IBFT2.DuplicateMesageLimit = 100
		}
		if r.Spec.Genesis.IBFT2.FutureMessagesLimit == 0 {
			r.Spec.Genesis.IBFT2.FutureMessagesLimit = 1000
		}
		if r.Spec.Genesis.IBFT2.FutureMessagesMaxDistance == 0 {
			r.Spec.Genesis.IBFT2.FutureMessagesMaxDistance = 10
		}

	}
}
