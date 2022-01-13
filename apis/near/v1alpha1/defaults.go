package v1alpha1

const (
	// DefaultRPCPort is the default JSON-RPC port
	DefaultRPCPort uint = 3030
	// DefaultP2PPort is the default p2p port
	DefaultP2PPort uint = 24567
	// DefaultRPCHost is the default JSON-RPC host
	DefaultRPCHost = "0.0.0.0"
)

// Resources
const (
	// DefaultNodeCPURequest is the cpu requested by NEAR node
	DefaultNodeCPURequest = "4"
	// DefaultNodeCPULimit is the cpu limit for NEAR node
	DefaultNodeCPULimit = "8"

	// DefaultNodeMemoryRequest is the memory requested by NEAR node
	DefaultNodeMemoryRequest = "4Gi"
	// DefaultNodeMemoryLimit is the memory limit for NEAR node
	DefaultNodeMemoryLimit = "8Gi"

	// DefaultNodeStorageRequest is the Storage requested by NEAR node
	DefaultNodeStorageRequest = "250Gi"
	// DefaultArchivalNodeStorageRequest is the Storage requested by NEAR archival node
	DefaultArchivalNodeStorageRequest = "4Ti"
)
