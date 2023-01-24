package v1alpha1

const (
	// DefaultHost is the default JSON-RPC server host
	DefaultHost = "0.0.0.0"
	// DefaultRPCPort is the default JSON-RPC port
	DefaultRPCPort uint = 20443
	// DefaultP2PPort is the default p2p bind port
	DefaultP2PPort uint = 20444
)

const (
	// DefaultStacksNodeImage is the default Stacks node client image
	DefaultStacksNodeImage = "kotalco/stacks:2.05.0.1.0"
)

// Resources
const (
	// DefaultNodeCPURequest is the cpu requested by Stacks node
	DefaultNodeCPURequest = "2"
	// DefaultNodeCPULimit is the cpu limit for Stacks node
	DefaultNodeCPULimit = "4"

	// DefaultNodeMemoryRequest is the memory requested by Stacks node
	DefaultNodeMemoryRequest = "4Gi"
	// DefaultNodeMemoryLimit is the memory limit for Stacks node
	DefaultNodeMemoryLimit = "8Gi"

	// DefaultNodeStorageRequest is the Storage requested by Stacks node
	DefaultNodeStorageRequest = "100Gi"
)
