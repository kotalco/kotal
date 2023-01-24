package v1alpha1

import "github.com/kotalco/kotal/apis/shared"

const (
	// DefaultSyncMode is the default blockchain sync mode
	DefaultSyncMode = FullSynchronization
	// DefaultLoggingVerbosity is the default node logging verbosity
	DefaultLoggingVerbosity = shared.InfoLogs
	// DefaultRetainedBlocks is the default node of blocks to retain if node isn't archive
	DefaultRetainedBlocks uint = 256
	// DefaultRPCPort is the default JSON-RPC server port
	DefaultRPCPort uint = 9933
	// DefaultP2PPort is the p2p protocol tcp port
	DefaultP2PPort uint = 30333
	// DefaultWSPort is the default websocket server port
	DefaultWSPort uint = 9944
	// DefaultTelemetryURL is the default telemetry service URL
	DefaultTelemetryURL = "wss://telemetry.polkadot.io/submit/ 0"
	// DefaultPrometheusPort is the default prometheus exporter port
	DefaultPrometheusPort uint = 9615
	// DefaultCORSDomain is the default browser origin allowed to access the JSON-RPC HTTP and WS servers
	DefaultCORSDomain = "all"
)

const (
	// DefaultPolkadotImage is the default polkadot client image
	DefaultPolkadotImage = "parity/polkadot:v0.9.32"
)

// Resources
const (
	// DefaultNodeCPURequest is the cpu requested by polkadot node
	DefaultNodeCPURequest = "4"
	// DefaultNodeCPULimit is the cpu limit for polkadot node
	DefaultNodeCPULimit = "8"

	// DefaultNodeMemoryRequest is the memory requested by polkadot node
	DefaultNodeMemoryRequest = "4Gi"
	// DefaultNodeMemoryLimit is the memory limit for polkadot node
	DefaultNodeMemoryLimit = "8Gi"

	// DefaultNodeStorageRequest is the Storage requested by polkadot node
	DefaultNodeStorageRequest = "80Gi"
)
