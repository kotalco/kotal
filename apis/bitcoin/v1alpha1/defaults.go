package v1alpha1

const (
	// DefaultMainnetRPCPort is the default JSON-RPC port for mainnet
	DefaultMainnetRPCPort uint = 8332
	// DefaultTestnetRPCPort is the default JSON-RPC port for testnet
	DefaultTestnetRPCPort uint = 18332
	// DefaultMainnetP2PPort is the default p2p port for mainnet
	DefaultMainnetP2PPort uint = 8333
	// DefaultTestnetP2PPort is the default p2p port for testnet
	DefaultTestnetP2PPort uint = 18333
)

const (
	// DefaultBitcoinCoreImage is the default Bitcoin core client image
	DefaultBitcoinCoreImage = "ruimarinho/bitcoin-core:23.0"
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
