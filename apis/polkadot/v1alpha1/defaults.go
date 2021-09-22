package v1alpha1

const (
	// DefaultSyncMode is the default blockchain sync mode
	DefaultSyncMode = FullSynchronization
	// DefaultLoggingVerbosity is the default node logging verbosity
	DefaultLoggingVerbosity = InfoLogs
	// DefaultRPCPort is the default JSON-RPC server port
	DefaultRPCPort uint = 9933
	// DefaultWSPort is the default websocket server port
	DefaultWSPort uint = 9944
	// DefaultTelemetryURL is the default telemetry service URL
	DefaultTelemetryURL = "wss://telemetry.polkadot.io/submit/ 0"
	// DefaultPrometheusPort is the default prometheus exporter port
	DefaultPrometheusPort uint = 9615
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
