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
	// DefaultAptosCoreMainnetImage is the default Aptos core Mainnet client image
	DefaultAptosCoreMainnetImage = "aptoslabs/validator@sha256:06ca1753786724805e7efb525bd2dbfbc5a114e8792a8d05ef522dba9830b613"
	// DefaultAptosCoreDevnetImage is the default Aptos core Devnet client image
	DefaultAptosCoreDevnetImage = "aptoslabs/validator@sha256:d017e7f56781ff26c4755c2a379810b9d9c2f263c5ade15a26ccb719c743f7de"
	// DefaultAptosCoreTestnetImage is the default Aptos core Testnet client image
	DefaultAptosCoreTestnetImage = "aptoslabs/validator@sha256:c109ab86066fc35cbff5d7f57340ea6da9ed480896d08cd1bbd30c3dec683033"
)

const (
	// DefaultMetricsPort is the default metrics server port
	DefaultMetricsPort uint = 9101
	// DefaultAPIPort is the default API server port
	DefaultAPIPort uint = 8080
	// DefaultFullnodeP2PPort is the default full node p2p port
	DefaultFullnodeP2PPort uint = 6182
	// DefaultValidatorP2PPort is the default validator node p2p port
	DefaultValidatorP2PPort uint = 6180
)
