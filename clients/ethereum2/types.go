package ethereum2

// Client home directories
const (
	// TekuHomeDir is teku home directory
	TekuHomeDir = "/opt/teku"
	// PrysmHomeDir is prysm home directory
	PrysmHomeDir = "/home/prysm"
	// NimbusHomeDir is nimbus home directory
	NimbusHomeDir = "/home/nimbus"
	// LighthouseHomeDir is lighthouse home directory
	LighthouseHomeDir = "/home/lighthouse"
)

// Teku client arguments
const (
	// TekuNetwork is the argument used for selecting network
	TekuNetwork = "--network"
	// TekuExecutionEngineEndpoint is the argument used to set execution engine endpoint
	TekuExecutionEngineEndpoint = "--ee-endpoint"
	// TekuFeeRecipient is the argument used to set default fee recipient
	TekuFeeRecipient = "--validators-proposer-default-fee-recipient"
	// TekuJwtSecretFile is the argument used to locate jwt secret file
	TekuJwtSecretFile = "--ee-jwt-secret-file"
	// TekuDataPath is the argument used for data directory
	TekuDataPath = "--data-path"
	// TekuRestEnabled is the argument used to enable Beacon REST API
	TekuRestEnabled = "--rest-api-enabled"
	// TekuRestPort is the argument used for Beacon REST API server port
	TekuRestPort = "--rest-api-port"
	// TekuRestHost is the argument used for Beacon REST API server host
	TekuRestHost = "--rest-api-interface"
	// TekuP2PPort is the argument used p2p and discovery port
	TekuP2PPort = "--p2p-port"
	// TekuRESTAPICorsOrigins is the argument used to whitelist domains for cross domain requests
	TekuRESTAPICorsOrigins = "--rest-api-cors-origins"
	// TekuRESTAPIHostAllowlist is the argument used to whitelist hosts for API access
	TekuRESTAPIHostAllowlist = "--rest-api-host-allowlist"
	// TekuLogging is the argument used to set logging verbosity level
	TekuLogging = "--logging"
	// TekuInitialState is the argument used for checkpoint sync
	TekuInitialState = "--initial-state"
	// TekuVC is the argument used to run validator client
	TekuVC = "vc"
	// TekuBeaconNodeEndpoint is the argument used for beacon node api endpoint
	TekuBeaconNodeEndpoint = "--beacon-node-api-endpoint"
	// TekuGraffiti is the argument used for text include in proposed blocks
	TekuGraffiti = "--validators-graffiti"
	// TekuValidatorKeys is the argument used for Validator keys and secrets
	TekuValidatorKeys = "--validator-keys"
	// TekuValidatorsKeystoreLockingEnabled is the argument used to enable keystore locking files
	TekuValidatorsKeystoreLockingEnabled = "--validators-keystore-locking-enabled"
)

// Prysm client arguments
const (
	// PrysmDataDir is the argument used for data directory
	PrysmDataDir = "--datadir"
	// PrysmExecutionEngineEndpoint is the argument used to set execution engine endpoint
	PrysmExecutionEngineEndpoint = "--execution-endpoint"
	// PrysmFeeRecipient is the argument used to set fee recipient
	PrysmFeeRecipient = "--suggested-fee-recipient"
	// PrysmJwtSecretFile is the argument used to locate jwt secret file
	PrysmJwtSecretFile = "--jwt-secret"
	// PrysmCheckpointSyncUrl is the argument used to set checkpoint sync beacon node rest api endpoint
	PrysmCheckpointSyncUrl = "--checkpoint-sync-url"
	// PrysmGenesisBeaconApiUrl is the argument used to set genesis beacon node api url
	PrysmGenesisBeaconApiUrl = "--genesis-beacon-api-url"
	// PrysmAcceptTermsOfUse is the argument used for accepting terms of use
	PrysmAcceptTermsOfUse = "--accept-terms-of-use"
	// PrysmRPCPort is the argument used for RPC server port
	PrysmRPCPort = "--rpc-port"
	// PrysmRPCHost is the argument used for host on which RPC server should listen
	PrysmRPCHost = "--rpc-host"
	// PrysmDisableGRPC is the argument used to disable GRPC gateway server
	PrysmDisableGRPC = "--disable-grpc-gateway"
	// PrysmGRPCPort is the argument used for GRPC gateway server port
	PrysmGRPCPort = "--grpc-gateway-port"
	// PrysmGRPCHost is the argument used for GRPC gateway server host
	PrysmGRPCHost = "--grpc-gateway-host"
	// PrysmP2PTCPPort is the argument used p2p tcp port
	PrysmP2PTCPPort = "--p2p-tcp-port"
	// PrysmP2PUDPPort is the argument used p2p discovery udp port
	PrysmP2PUDPPort = "--p2p-udp-port"
	// PrysmGRPCGatewayCorsDomains is the argument used to whitelist domains for cross domain requests
	PrysmGRPCGatewayCorsDomains = "--grpc-gateway-corsdomain"
	// PrysmLogging is the argument used to set logging verbosity level
	PrysmLogging = "--verbosity"

	// PrysmTLSCert is the argument used to locate tls certificate
	PrysmTLSCert = "--tls-cert"
	// PrysmTLSKey is the argument used to locate tls key
	PrysmTLSKey = "--tls-key"

	// PrysmBeaconRPCProvider is the argument used for beacon node rpc endpoint
	PrysmBeaconRPCProvider = "--beacon-rpc-provider"
	// PrysmGraffiti is the argument used to include in proposed blocks
	PrysmGraffiti = "--graffiti"
	// PrysmKeysDir is the argument used to locate keystores to be imported are stored
	PrysmKeysDir = "--keys-dir"
	// PrysmWalletDir is the argument used to locate wallet directory
	PrysmWalletDir = "--wallet-dir"
	// PrysmAccountPasswordFile is the argument used to locate account password file
	PrysmAccountPasswordFile = "--account-password-file"
	// PrysmWalletPasswordFile is the argument used to locate wallet password file
	PrysmWalletPasswordFile = "--wallet-password-file"
)

// Lighthouse client arguments
const (
	// LighthouseDataDir is the argument used for data directory
	LighthouseDataDir = "--datadir"
	// LighthouseNetwork is the argument used for selecting network
	LighthouseNetwork = "--network"
	// LighthouseCheckpointSyncUrl is the argument used for checkpoint sync
	LighthouseCheckpointSyncUrl = "--checkpoint-sync-url"
	// LighthouseHTTP is the argument used to enable Beacon REST API
	LighthouseHTTP = "--http"
	// LighthouseAllowOrigins is the argument used to whitelist domains for cross domain requests
	LighthouseAllowOrigins = "--http-allow-origin"
	// LighthouseHTTPPort is the argument used for Beacon REST API server port
	LighthouseHTTPPort = "--http-port"
	// LighthouseHTTPAddress is the argument used for Beacon REST API server host
	LighthouseHTTPAddress = "--http-address"
	// LighthouseExecutionEngineEndpoint is the argument used for Execution engine endpoint
	LighthouseExecutionEngineEndpoint = "--execution-endpoint"
	// LighthouseJwtSecretFile is the argument used to locate jwt secret file
	LighthouseJwtSecretFile = "--execution-jwt"
	// LighthouseFeeRecipient is the argument used to set fee recipient
	LighthouseFeeRecipient = "--suggested-fee-recipient"
	// LighthousePort is the argument used for p2p tcp port
	LighthousePort = "--port"
	// LighthouseDiscoveryPort is the argument used for discovery udp port
	LighthouseDiscoveryPort = "--discovery-port"
	// LighthouseDebugLevel is the argument used to set logging verbosity level
	LighthouseDebugLevel = "--debug-level"

	// LighthouseBeaconNodeEndpoints is the argument used for beacon node endpoint
	LighthouseBeaconNodeEndpoints = "--beacon-nodes"
	// LighthouseGraffiti is the argument used to include in proposed blocks
	LighthouseGraffiti = "--graffiti"
	// LighthouseDisableAutoDiscover is the argument used to disable auto validator keystores discovery
	LighthouseDisableAutoDiscover = "--disable-auto-discover"
	// LighthouseInitSlashingProtection is the argument used to init slashing protection
	LighthouseInitSlashingProtection = "--init-slashing-protection"
	// LighthouseReusePassword is the argument used to reuse password during keystore import
	LighthouseReusePassword = "--reuse-password"
	// LighthouseKeystore is the argument used to locate keystore file
	LighthouseKeystore = "--keystore"
	// LighthousePasswordFile is the argument used to locate password file
	LighthousePasswordFile = "--password-file"
)

// Nimbus client arguments
const (
	// NimbusDataDir is the argument used for data directory
	NimbusDataDir = "--data-dir"
	// NimbusNonInteractive is the argument used for non interactive mode
	NimbusNonInteractive = "--non-interactive"
	// NimbusNetwork is the argument used for selecting network
	NimbusNetwork = "--network"
	// NimbusExecutionEngineEndpoint is the argument used for Execution engine endpoint
	NimbusExecutionEngineEndpoint = "--web3-url"
	// NimbusJwtSecretFile is the argument used to locate jwt secret file
	NimbusJwtSecretFile = "--jwt-secret"
	// NimbusFeeRecipient is the argument used to set fee recipient
	NimbusFeeRecipient = "--suggested-fee-recipient"
	// NimbusTCPPort is the argument used for p2p tcp port
	NimbusTCPPort = "--tcp-port"
	// NimbusUDPPort is the argument used for discovery udp port
	NimbusUDPPort = "--udp-port"
	// NimbusLogging is the argument used to set logging verbosity level
	NimbusLogging = "--log-level"
	// NimbusREST is the argument used to enable REST server
	NimbusREST = "--rest"
	// NimbusRESTPort is the argument used to set REST server port
	NimbusRESTPort = "--rest-port"
	// NimbusRESTAddress is the argument used to set REST server listening address
	NimbusRESTAddress = "--rest-address"
	// NimbusRESTAllowOrigin is the argument used to limit access to the REST API
	NimbusRESTAllowOrigin = "--rest-allow-origin"
	// NimbusGraffiti is the argument used to include in proposed blocks
	NimbusGraffiti = "--graffiti"
	// NimbusValidatorsDir is the argument used to locate validator keystores directory
	NimbusValidatorsDir = "--validators-dir"
	// NimbusSecretsDir is the argument used to locate validator keystores secrets directory
	NimbusSecretsDir = "--secrets-dir"
	// NimbusBeaconNodes is the argument used to set one or more beacon node HTTP REST APIs
	NimbusBeaconNodes = "--beacon-node"
)
