package v1alpha1

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
