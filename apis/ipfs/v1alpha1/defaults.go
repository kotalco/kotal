package v1alpha1

const (
	// DefaultRoutingMode is default content routing mechanism
	DefaultRoutingMode = DHTRouting
	// DefaultAPIPort is default API port
	DefaultAPIPort = 5001
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
