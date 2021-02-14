package controllers

const (
	// PathBlockchainData is the blockchain data path
	PathBlockchainData = "/mnt/data"
)

// Teku client arguments
const (
	// TekuNetwork is the argument used for selecting network
	TekuNetwork = "--network"
	// TekuEth1Endpoint is the argument used for Ethereum 1 JSON RPC endpoint
	TekuEth1Endpoint = "--eth1-endpoint"
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

	// TekuBeaconNodeEndpoint is the argument used for beacon node api endpoint
	TekuBeaconNodeEndpoint = "--beacon-node-api-endpoint"
)

// Prysm client arguments
const (
	// PrysmDataDir is the argument used for data directory
	PrysmDataDir = "--datadir"
	// PrysmWeb3Provider is the argument used for Ethereum 1 JSON RPC endpoint
	PrysmWeb3Provider = "--http-web3provider"
	// PrysmFallbackWeb3Provider is the argument used for fallback Ethereum 1 JSON RPC endpoints
	PrysmFallbackWeb3Provider = "--fallback-web3provider"
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

	//	PrysmBeaconRPCProvider is the argument used for beacon node rpc endpoint
	PrysmBeaconRPCProvider = "--beacon-rpc-provider"
)

// Lighthouse client arguments
const (
	// LighthouseDataDir is the argument used for data directory
	LighthouseDataDir = "--datadir"
	// LighthouseNetwork is the argument used for selecting network
	LighthouseNetwork = "--network"
	// LighthouseEth1 is the argument used for connecting to Ethereum 1 node
	LighthouseEth1 = "--eth1"
	// LighthouseHTTP is the argument used to enable Beacon REST API
	LighthouseHTTP = "--http"
	// LighthouseHTTPPort is the argument used for Beacon REST API server port
	LighthouseHTTPPort = "--http-port"
	// LighthouseHTTPAddress is the argument used for Beacon REST API server host
	LighthouseHTTPAddress = "--http-address"
	// LighthouseEth1Endpoints is the argument used for Ethereum 1 JSON RPC endpoints
	LighthouseEth1Endpoints = "--eth1-endpoints"
	// LighthousePort is the argument used for p2p tcp port
	LighthousePort = "--port"
	// LighthouseDiscoveryPort is the argument used for discovery udp port
	LighthouseDiscoveryPort = "--discovery-port"

	// LighthouseBeaconNodeEndpoint is the argument used for beacon node endpoint
	LighthouseBeaconNodeEndpoint = "--beacon-node"
)

// Nimbus client arguments
const (
	// NimbusDataDir is the argument used for data directory
	NimbusDataDir = "--data-dir"
	// NimbusNonInteractive is the argument used for non interactive mode
	NimbusNonInteractive = "--non-interactive"
	// NimbusNetwork is the argument used for selecting network
	NimbusNetwork = "--network"
	// NimbusEth1Endpoint is the argument used for Ethereum 1 JSON RPC endpoint
	NimbusEth1Endpoint = "--web3-url"
	// NimbusRPC is the argument used to enable RPC server
	NimbusRPC = "--rpc"
	// NimbusRPCPort is the argument used for RPC server port
	NimbusRPCPort = "--rpc-port"
	// NimbusRPCAddress is the argument used for host on which RPC server should listen
	NimbusRPCAddress = "--rpc-address"
	// NimbusTCPPort is the argument used for p2p tcp port
	NimbusTCPPort = "--tcp-port"
	// NimbusUDPPort is the argument used for discovery udp port
	NimbusUDPPort = "--udp-port"
)
