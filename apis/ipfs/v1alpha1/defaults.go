package v1alpha1

const (
	// DefaultRoutingMode is the default content routing mechanism
	DefaultRoutingMode = DHTRouting
	// DefaultAPIPort is the default API port
	DefaultAPIPort = 5001
	// DefaultHost is the default API host
	DefaultHost = "0.0.0.0"
)

// Resources
const (
	// DefaultNodeCPURequest is the cpu requested by ipfs node
	DefaultNodeCPURequest = "1"
	// DefaultNodeCPULimit is the cpu limit for ipfs node
	DefaultNodeCPULimit = "2"

	// DefaultNodeMemoryRequest is the memory requested by ipfs node
	DefaultNodeMemoryRequest = "2Gi"
	// DefaultNodeMemoryLimit is the memory limit for ipfs node
	DefaultNodeMemoryLimit = "4Gi"

	// DefaultNodeStorageRequest is the Storage requested by ipfs node
	DefaultNodeStorageRequest = "10Gi"
)
