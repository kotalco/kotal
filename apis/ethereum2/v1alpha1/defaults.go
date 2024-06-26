package v1alpha1

import "github.com/kotalco/kotal/apis/shared"

var (
	// DefaultOrigins is the default domains from which to accept cross origin requests
	DefaultOrigins = []string{"*"}
)

const (
	// DefaltReplicas is the default replicas
	DefaltReplicas uint = 1
	// ZeroAddress is Ethereum zero address
	ZeroAddress = "0x0000000000000000000000000000000000000000"
	// DefaultP2PPort is the default port used for p2p and discovery
	DefaultP2PPort uint = 9000
	// DefaultRestPort is the default Beacon REST api port
	DefaultRestPort uint = 5051
	// DefaultRPCPort is the default RPC server port
	DefaultRPCPort uint = 4000
	// DefaultGRPCPort is the default GRPC gateway server port
	DefaultGRPCPort uint = 3500
	// DefaultGraffiti is the default text to include in proposed blocks
	DefaultGraffiti = "Powered by Kotal"
	// DefaultLogging is the default logging verbosity
	DefaultLogging = shared.InfoLogs
)

const (
	// DefaultLighthouseBeaconNodeImage is the default SigmaPrime Ethereum 2.0 beacon node image
	DefaultLighthouseBeaconNodeImage = "kotalco/lighthouse:v5.1.3"
	// DefaultTekuBeaconNodeImage is PegaSys Teku beacon node image
	DefaultTekuBeaconNodeImage = "consensys/teku:24.3.1"
	// DefaultPrysmBeaconNodeImage is Prysmatic Labs beacon node image
	DefaultPrysmBeaconNodeImage = "kotalco/prysm:v5.0.3"
	// DefaultNimbusBeaconNodeImage is the default Status Ethereum 2.0 beacon node image
	DefaultNimbusBeaconNodeImage = "kotalco/nimbus:v24.3.0"
)

const (
	// DefaultTekuValidatorImage is PegaSys Teku validator client image
	DefaultTekuValidatorImage = "consensys/teku:24.3.1"
	// DefaultPrysmValidatorImage is Prysmatic Labs validator client image
	DefaultPrysmValidatorImage = "kotalco/prysm:v5.0.3"
	// DefaultNimbusValidatorImage is the default Status Ethereum 2.0 validator client image
	DefaultNimbusValidatorImage = "kotalco/nimbus:v24.3.0"
	// DefaultLighthouseValidatorImage is the default SigmaPrime Ethereum 2.0 validator client image
	DefaultLighthouseValidatorImage = "kotalco/lighthouse:v5.1.3"
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
