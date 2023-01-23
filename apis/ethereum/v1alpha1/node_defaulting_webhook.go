package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate-ethereum-kotal-io-v1alpha1-node,mutating=true,failurePolicy=fail,groups=ethereum.kotal.io,resources=nodes,verbs=create;update,versions=v1alpha1,name=mutate-ethereum-v1alpha1-node.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Defaulter = &Node{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (n *Node) Default() {
	defaultAPIs := []API{Web3API, ETHAPI, NetworkAPI}
	client := n.Spec.Client
	nethermindNode := client == NethermindClient

	if n.Spec.Image == "" {
		var image string

		switch client {
		case BesuClient:
			image = DefaultBesuImage
		case GethClient:
			image = DefaultGethImage
		case NethermindClient:
			image = DefaultNethermindImage
		}

		n.Spec.Image = image
	}

	// default genesis block
	if n.Spec.Genesis != nil {
		n.Spec.Genesis.Default()
	}

	if n.Spec.P2PPort == 0 {
		n.Spec.P2PPort = DefaultP2PPort
	}

	if n.Spec.SyncMode == "" {
		// public network
		if n.Spec.Genesis == nil {
			if n.Spec.Client == GethClient {
				n.Spec.SyncMode = SnapSynchronization
			} else {
				n.Spec.SyncMode = FastSynchronization
			}
		} else {
			n.Spec.SyncMode = FullSynchronization
		}
	}

	// must be called after defaulting sync mode because it's depending on its value
	n.DefaultNodeResources()

	// nethermind doesn't support host whitelisting
	if len(n.Spec.Hosts) == 0 && !nethermindNode {
		n.Spec.Hosts = DefaultOrigins
	}

	// nethermind doesn't support cors domains
	if len(n.Spec.CORSDomains) == 0 && !nethermindNode {
		n.Spec.CORSDomains = DefaultOrigins
	}

	if n.Spec.EnginePort == 0 {
		n.Spec.EnginePort = DefaultEngineRPCPort
	}

	if n.Spec.RPCPort == 0 {
		n.Spec.RPCPort = DefaultRPCPort
	}

	if len(n.Spec.RPCAPI) == 0 {
		n.Spec.RPCAPI = defaultAPIs
	}

	if n.Spec.WSPort == 0 {
		n.Spec.WSPort = DefaultWSPort
	}

	if len(n.Spec.WSAPI) == 0 {
		n.Spec.WSAPI = defaultAPIs
	}

	if n.Spec.GraphQLPort == 0 {
		n.Spec.GraphQLPort = DefaultGraphQLPort
	}

	if n.Spec.Logging == "" {
		n.Spec.Logging = DefaultLogging
	}

}

// DefaultNodeResources defaults node cpu, memory and storage resources
func (n *Node) DefaultNodeResources() {
	var cpu, cpuLimit, memory, memoryLimit, storage string
	privateNetwork := n.Spec.Genesis != nil
	network := n.Spec.Network

	if n.Spec.Resources.CPU == "" {
		if privateNetwork {
			cpu = DefaultPrivateNetworkNodeCPURequest
		} else {
			cpu = DefaultPublicNetworkNodeCPURequest
		}

		n.Spec.Resources.CPU = cpu
	}

	if n.Spec.Resources.CPULimit == "" {
		if privateNetwork {
			cpuLimit = DefaultPrivateNetworkNodeCPULimit
		} else {
			cpuLimit = DefaultPublicNetworkNodeCPULimit
		}

		n.Spec.Resources.CPULimit = cpuLimit
	}

	if n.Spec.Resources.Memory == "" {
		if privateNetwork {
			memory = DefaultPrivateNetworkNodeMemoryRequest
		} else {
			memory = DefaultPublicNetworkNodeMemoryRequest
		}

		n.Spec.Resources.Memory = memory
	}

	if n.Spec.Resources.MemoryLimit == "" {
		if privateNetwork {
			memoryLimit = DefaultPrivateNetworkNodeMemoryLimit
		} else {
			memoryLimit = DefaultPublicNetworkNodeMemoryLimit
		}

		n.Spec.Resources.MemoryLimit = memoryLimit
	}

	if n.Spec.Resources.Storage == "" {
		if privateNetwork {
			storage = DefaultPrivateNetworkNodeStorageRequest
		} else if network == MainNetwork && n.Spec.SyncMode == FastSynchronization {
			storage = DefaultMainNetworkFastNodeStorageRequest
		} else if network == MainNetwork && n.Spec.SyncMode == FullSynchronization {
			storage = DefaultMainNetworkFullNodeStorageRequest
		} else {
			storage = DefaultTestNetworkStorageRequest
		}

		n.Spec.Resources.Storage = storage
	}

}
