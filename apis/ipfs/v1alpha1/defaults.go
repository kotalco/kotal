package v1alpha1

import "github.com/kotalco/kotal/apis/shared"

const (
	// DefaultRoutingMode is the default content routing mechanism
	DefaultRoutingMode = DHTRouting
	// DefaultAPIPort is the default API port
	DefaultAPIPort uint = 5001
	// DefaultGatewayPort is the default local gateway port
	DefaultGatewayPort uint = 8080
	// LocalHost is the local host
	LocalHost = "127.0.0.1"
	// DefaultHost is the default API host
	DefaultHost = "0.0.0.0"
	// DefaultLogging is the default logging verbosity level
	DefaultLogging = shared.InfoLogs
)

const (
	// DefaultGoIPFSImage is the default go ipfs client image
	DefaultGoIPFSImage = "kotalco/kubo:v0.17.0"
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

// Cluster peer
const (
	// DefaultIPFSClusterConsensus is the default ipfs cluster consensus algorithm
	DefaultIPFSClusterConsensus = CRDT
)
