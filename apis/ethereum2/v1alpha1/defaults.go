package v1alpha1

const (
	// DefaultClient is the default ethereum 2.0 client
	DefaultClient = TekuClient
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
)
