package v1alpha1

var (
	// DefaultOrigins is the default domains from which to accept cross origin requests
	DefaultOrigins = []string{"*"}
)

const (
	// DefaultP2PPort is the default port used for p2p and discovery
	DefaultP2PPort uint = 9000
	// DefaultRestPort is the default Beacon REST api port
	DefaultRestPort uint = 5051
	// DefaultRPCPort is the default RPC server port
	DefaultRPCPort uint = 4000
	// DefaultGRPCPort is the default GRPC gateway server port
	DefaultGRPCPort uint = 3500
	// DefaultRPCHost is the default host on which RPC server should listen
	DefaultRPCHost = "0.0.0.0"
	// DefaultGRPCHost is the default host on which GRPC gateway server should listen
	DefaultGRPCHost = "0.0.0.0"
	// DefaultRestHost is the default Beacon REST api host
	DefaultRestHost = "0.0.0.0"
	// DefaultGraffiti is the default text to include in proposed blocks
	DefaultGraffiti = "Powered by Kotal"
)

const (
	// DefaultCPURequest is the default CPU cores required by Ethereum 2.0 node
	DefaultCPURequest = "4"
	// DefaultCPULimit is the default CPU cores limit by Ethereum 2.0 node
	DefaultCPULimit = "8"
	// DefaultMemoryRequest is the default memory required by Ethereum 2.0 node
	DefaultMemoryRequest = "8Gi"
	// DefaultMemoryLimit is the default memory limit by Ethereum 2.0 node
	DefaultMemoryLimit = "16Gi"
	// DefaultStorage is the default disk space used by Ethereum 2.0 node
	DefaultStorage = "200Gi"
)
