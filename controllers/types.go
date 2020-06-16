package controllers

const (
	// PathNodekey is the node key path
	PathNodekey = "/mnt/bootnode"
	// PathGenesisFile is the genesis file path
	PathGenesisFile = "/mnt/config"
	// PathBlockchainData is the blockchain data path
	PathBlockchainData = "/mnt/data"
)

const (
	// ArgNatMethod is the argument used for nat method
	ArgNatMethod = "--nat-method"
	// ArgNodePrivateKey is the argument used for node private key
	ArgNodePrivateKey = "--node-private-key-file"
	// ArgGenesisFile is the argument used for genesis file
	ArgGenesisFile = "--genesis-file"
	// ArgDataPath is the argument used for data path
	ArgDataPath = "--data-path"
	// ArgNetwork is the argument used for selecting network
	ArgNetwork = "--network"
	// ArgP2PPort is the argument used for p2p port
	ArgP2PPort = "--p2p-port"
	// ArgBootnodes is the argument used for bootnodes
	ArgBootnodes = "--bootnodes"
	// ArgSyncMode is the argument used for sync mode
	ArgSyncMode = "--sync-mode"
	// ArgMinerEnabled is the argument used for turning on mining
	ArgMinerEnabled = "--miner-enabled"
	// ArgMinerCoinbase is the argument used for setting coinbase account
	ArgMinerCoinbase = "--miner-coinbase"
	// ArgRPCHTTPCorsOrigins is the argument used for setting rpc HTTP cors origins
	ArgRPCHTTPCorsOrigins = "--rpc-http-cors-origins"
	// ArgRPCHTTPEnabled is the argument used to enable RPC over HTTP
	ArgRPCHTTPEnabled = "--rpc-http-enabled"
	// ArgRPCHTTPPort is the argument used for RPC HTTP port
	ArgRPCHTTPPort = "--rpc-http-port"
	// ArgRPCHTTPHost is the argument used for RPC HTTP Host
	ArgRPCHTTPHost = "--rpc-http-host"
	// ArgRPCHTTPAPI is the argument used for RPC HTTP APIs
	ArgRPCHTTPAPI = "--rpc-http-api"
	// ArgRPCWSEnabled is the argument used to enable RPC WS
	ArgRPCWSEnabled = "--rpc-ws-enabled"
	// ArgRPCWSPort is the argument used for RPC WS port
	ArgRPCWSPort = "--rpc-ws-port"
	// ArgRPCWSHost is the argument used for RPC WS host
	ArgRPCWSHost = "--rpc-ws-host"
	// ArgRPCWSAPI is the argument used for RPC WS APIs
	ArgRPCWSAPI = "--rpc-ws-api"
	// ArgGraphQLHTTPEnabled is the argument used to enable QraphQL HTTP server
	ArgGraphQLHTTPEnabled = "--graphql-http-enabled"
	// ArgGraphQLHTTPPort is the argument used for GraphQL HTTP port
	ArgGraphQLHTTPPort = "--graphql-http-port"
	// ArgGraphQLHTTPHost is the argument used for GraphQL HTTP host
	ArgGraphQLHTTPHost = "--graphql-http-host"
	// ArgGraphQLHTTPCorsOrigins is the argument used for GraphQL HTTP Cors origins
	ArgGraphQLHTTPCorsOrigins = "--graphql-http-cors-origins"
	// ArgHostWhitelist is the argument used for whitelisting hosts
	ArgHostWhitelist = "--host-whitelist"
)
