package v1alpha1

// Resources
const (
	// DefaultNodeCPURequest is the cpu requested by Aptos node
	DefaultNodeCPURequest = "2"
	// DefaultNodeCPULimit is the cpu limit for Aptos node
	DefaultNodeCPULimit = "4"

	// DefaultNodeMemoryRequest is the memory requested by Aptos node
	DefaultNodeMemoryRequest = "4Gi"
	// DefaultNodeMemoryLimit is the memory limit for Aptos node
	DefaultNodeMemoryLimit = "8Gi"

	// DefaultNodeStorageRequest is the Storage requested by Aptos node
	DefaultNodeStorageRequest = "250Gi"
)

const (
	// DefaultAptosCoreDevnetImage is the default Aptos core Devnet client image
	DefaultAptosCoreDevnetImage = "aptoslab/validator:devnet"
	// DefaultAptosCoreTestnetImage is the default Aptos core Testnet client image
	DefaultAptosCoreTestnetImage = "aptoslab/validator:testnet"
)

const (
	// DefaultAPIPort is the default API server port
	DefaultAPIPort uint = 8080
	// DefaultFullnodeP2PPort is the default full node p2p port
	DefaultFullnodeP2PPort uint = 6182
	// DefaultValidatorP2PPort is the default validator node p2p port
	DefaultValidatorP2PPort uint = 6180
)
