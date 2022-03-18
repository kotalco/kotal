package v1alpha1

const (
	// DefaultMainnetRPCPort is the default JSON-RPC port for mainnet
	DefaultMainnetRPCPort uint = 8332
	// DefaultTestnetRPCPort is the default JSON-RPC port for testnet
	DefaultTestnetRPCPort uint = 18332
	// DefaultRPCHost is the default JSON-RPC server host
	DefaultRPCHost = "0.0.0.0"
)

// Resources
const (
	// DefaultNodeCPURequest is the cpu requested by Bitcoin node
	DefaultNodeCPURequest = "2"
	// DefaultNodeCPULimit is the cpu limit for Bitcoin node
	DefaultNodeCPULimit = "4"

	// DefaultNodeMemoryRequest is the memory requested by Bitcoin node
	DefaultNodeMemoryRequest = "4Gi"
	// DefaultNodeMemoryLimit is the memory limit for Bitcoin node
	DefaultNodeMemoryLimit = "8Gi"

	// DefaultNodeStorageRequest is the Storage requested by Bitcoin node
	DefaultNodeStorageRequest = "100Gi"
)
