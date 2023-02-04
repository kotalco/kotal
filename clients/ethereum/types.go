package ethereum

// Hyperledger Besu client arguments
const (
	// BesuLogging is the argument used for logging verbosity level
	BesuLogging = "--logging"
	// BesuNetworkID is the argument used for network id
	BesuNetworkID = "--network-id"
	// BesuNatMethod is the argument used for nat method
	BesuNatMethod = "--nat-method"
	// BesuNodePrivateKey is the argument used for node private key
	BesuNodePrivateKey = "--node-private-key-file"
	// BesuGenesisFile is the argument used for genesis file
	BesuGenesisFile = "--genesis-file"
	// BesuDataPath is the argument used for data path
	BesuDataPath = "--data-path"
	// BesuNetwork is the argument used for selecting network
	BesuNetwork = "--network"
	// BesuDiscoveryEnabled is the argument used to enabled discovery
	BesuDiscoveryEnabled = "--discovery-enabled"
	// BesuP2PPort is the argument used for p2p port
	BesuP2PPort = "--p2p-port"
	// BesuBootnodes is the argument used for bootnodes
	BesuBootnodes = "--bootnodes"
	// BesuSyncMode is the argument used for sync mode
	BesuSyncMode = "--sync-mode"
	// BesuMinerEnabled is the argument used for turning on mining
	BesuMinerEnabled = "--miner-enabled"
	// BesuMinerCoinbase is the argument used for setting coinbase account
	BesuMinerCoinbase = "--miner-coinbase"
	// BesuRPCHTTPCorsOrigins is the argument used for setting rpc HTTP cors origins
	BesuRPCHTTPCorsOrigins = "--rpc-http-cors-origins"
	// BesuRPCHTTPEnabled is the argument used to enable RPC over HTTP
	BesuRPCHTTPEnabled = "--rpc-http-enabled"
	// BesuRPCHTTPPort is the argument used for RPC HTTP port
	BesuRPCHTTPPort = "--rpc-http-port"
	// BesuRPCHTTPHost is the argument used for RPC HTTP Host
	BesuRPCHTTPHost = "--rpc-http-host"
	// BesuRPCHTTPAPI is the argument used for RPC HTTP APIs
	BesuRPCHTTPAPI = "--rpc-http-api"

	// BesuEngineRpcEnabled is the argument used to enable Engine RPC APIs
	BesuEngineRpcEnabled = "--engine-rpc-enabled"
	// BesuEngineJwtSecret is the argument used to locate JWT secret
	BesuEngineJwtSecret = "--engine-jwt-secret"
	// BesuEngineRpcPort is the argument used to set Engine RPC listening port
	BesuEngineRpcPort = "--engine-rpc-port"
	// BesuEngineHostAllowList is the argument used to set hosts from which to accept requests
	BesuEngineHostAllowList = "--engine-host-allowlist"

	// BesuRPCWSEnabled is the argument used to enable RPC WS
	BesuRPCWSEnabled = "--rpc-ws-enabled"
	// BesuRPCWSPort is the argument used for RPC WS port
	BesuRPCWSPort = "--rpc-ws-port"
	// BesuRPCWSHost is the argument used for RPC WS host
	BesuRPCWSHost = "--rpc-ws-host"
	// BesuRPCWSAPI is the argument used for RPC WS APIs
	BesuRPCWSAPI = "--rpc-ws-api"
	// BesuGraphQLHTTPEnabled is the argument used to enable QraphQL HTTP server
	BesuGraphQLHTTPEnabled = "--graphql-http-enabled"
	// BesuGraphQLHTTPPort is the argument used for GraphQL HTTP port
	BesuGraphQLHTTPPort = "--graphql-http-port"
	// BesuGraphQLHTTPHost is the argument used for GraphQL HTTP host
	BesuGraphQLHTTPHost = "--graphql-http-host"
	// BesuGraphQLHTTPCorsOrigins is the argument used for GraphQL HTTP Cors origins
	BesuGraphQLHTTPCorsOrigins = "--graphql-http-cors-origins"
	// BesuHostAllowlist is the argument used for whitelisting hosts
	BesuHostAllowlist = "--host-allowlist"
	// BesuStaticNodesFile is the argument used to locate static nodes file
	BesuStaticNodesFile = "--static-nodes-file"
)

// Go ethereum client arguments
const (
	// GethLogging is the argument used for logging verbosity level
	GethLogging = "--verbosity"
	// GethConfig is the argument used for config file
	GethConfig = "--config"
	// GethNetworkID is the argument used for network id
	GethNetworkID = "--networkid"
	// GethNodeKey is the argument used for node private key
	GethNodeKey = "--nodekey"
	// GethNoDiscovery is the argument used to disable discovery
	GethNoDiscovery = "--nodiscover"
	// GethDataDir is the argument used for data path
	GethDataDir = "--datadir"
	// GethDisableIPC is the argument used to disable ipc servr
	GethDisableIPC = "--ipcdisable"
	// GethP2PPort is the argument used for p2p port
	GethP2PPort = "--port"
	// GethBootnodes is the argument used for bootnodes
	GethBootnodes = "--bootnodes"
	// GethSyncMode is the argument used for sync mode
	GethSyncMode = "--syncmode"
	// GethGcMode is the argument used for garbage collection mode
	GethGcMode = "--gcmode"
	// GethTxLookupLimit is the argument used to set recent number of blocks to maintain transactions index for
	GethTxLookupLimit = "--txlookuplimit"
	// GethCachePreImages is the argument used to enable recording the sha3 preimages of trie keys
	GethCachePreImages = "--cache.preimages"

	// GethMinerEnabled is the argument used for turning on mining
	GethMinerEnabled = "--mine"
	// GethMinerCoinbase is the argument used for setting coinbase account
	GethMinerCoinbase = "--miner.etherbase"

	// GethRPCHTTPCorsOrigins is the argument used for setting rpc HTTP cors origins
	GethRPCHTTPCorsOrigins = "--http.corsdomain"
	// GethRPCHTTPEnabled is the argument used to enable RPC over HTTP
	GethRPCHTTPEnabled = "--http"
	// GethRPCHTTPPort is the argument used for RPC HTTP port
	GethRPCHTTPPort = "--http.port"
	// GethRPCHTTPHost is the argument used for RPC HTTP Host
	GethRPCHTTPHost = "--http.addr"
	// GethRPCHTTPAPI is the argument used for RPC HTTP APIs
	GethRPCHTTPAPI = "--http.api"
	// GethRPCHostWhitelist is the argument used for whitelisting hosts
	GethRPCHostWhitelist = "--http.vhosts"

	// GethAuthRPCAddress is the argument used for listening address for authenticated APIs
	GethAuthRPCAddress = "--authrpc.addr"
	// GethAuthRPCPort is the argument used for listening port for authenticated APIs
	GethAuthRPCPort = "--authrpc.port"
	// GethAuthRPCHosts is the argument used for hostnames from which to accept requests
	GethAuthRPCHosts = "--authrpc.vhosts"
	// GethAuthRPCJwtSecret is the argument used for JWT secret to use for authenticated RPC endpoints
	GethAuthRPCJwtSecret = "--authrpc.jwtsecret"

	// --authrpc.jwtsecret

	// GethRPCWSEnabled is the argument used to enable RPC WS
	GethRPCWSEnabled = "--ws"
	// GethRPCWSPort is the argument used for RPC WS port
	GethRPCWSPort = "--ws.port"
	// GethRPCWSHost is the argument used for RPC WS host
	GethRPCWSHost = "--ws.addr"
	// GethRPCWSAPI is the argument used for RPC WS APIs
	GethRPCWSAPI = "--ws.api"
	// GethWSOrigins is the argument used for RPC WS origins
	GethWSOrigins = "--ws.origins"

	// GethGraphQLHTTPEnabled is the argument used to enable QraphQL HTTP server
	GethGraphQLHTTPEnabled = "--graphql"
	// GethGraphQLHTTPCorsOrigins is the argument used for GraphQL HTTP Cors origins
	GethGraphQLHTTPCorsOrigins = "--graphql.corsdomain"
	// GethGraphQLHostWhitelist is the argument used for whitelisting hosts
	GethGraphQLHostWhitelist = "--graphql.vhosts"
	// GethUnlock is the argument used for unlocking imported ethereum account
	GethUnlock = "--unlock"
	// GethPassword is the argument used for locking imported ethereum address
	GethPassword = "--password"
)

// Parity client arguments
const (
	// ParityLogging is the argument used for logging verbosity level
	ParityLogging = "--logging"
	// ParityNetworkID is the argument used for network id
	ParityNetworkID = "--network-id"
	// ParityNodeKey is the argument used for node key
	ParityNodeKey = "--node-key"
	// ParityDataDir is the argument used for data path
	ParityDataDir = "--base-path"
	// ParityReservedPeers is the argument used for static nodes (reserved peers)
	ParityReservedPeers = "--reserved-peers"
	// ParityNetwork is the argument used for selecting network
	ParityNetwork = "--chain"
	// ParityNoDiscovery is the argument used to disable discovery
	ParityNoDiscovery = "--no-discovery"
	// ParityP2PPort is the argument used for p2p port
	ParityP2PPort = "--port"
	// ParityBootnodes is the argument used for bootnodes
	ParityBootnodes = "--bootnodes"
	// ParitySyncMode is the argument used for sync mode
	ParitySyncMode = "--pruning"
	// ParityMinerCoinbase is the argument used for setting coinbase account
	ParityMinerCoinbase = "--author"
	// ParityEngineSigner is the argument used for engine singer
	ParityEngineSigner = "--engine-signer"

	// ParityDisableRPC is the argument used to disable JSON RPC HTTP server
	ParityDisableRPC = "--no-jsonrpc"
	// ParityRPCHTTPCorsOrigins is the argument used for setting rpc HTTP cors origins
	ParityRPCHTTPCorsOrigins = "--jsonrpc-cors"
	// ParityRPCHTTPPort is the argument used for RPC HTTP port
	ParityRPCHTTPPort = "--jsonrpc-port"
	// ParityRPCHTTPHost is the argument used for RPC HTTP Host
	ParityRPCHTTPHost = "--jsonrpc-interface"
	// ParityRPCHTTPAPI is the argument used for RPC HTTP APIs
	ParityRPCHTTPAPI = "--jsonrpc-apis"
	// ParityRPCHostWhitelist is the argument used for whitelisting hosts
	ParityRPCHostWhitelist = "--jsonrpc-hosts"

	// ParityDisableWS is the argument used for RPC WS port
	ParityDisableWS = "--no-ws"
	// ParityRPCWSCorsOrigins is the argument used for setting RPC WS cors origins
	ParityRPCWSCorsOrigins = "--ws-origins"
	// ParityRPCWSPort is the argument used for RPC WS port
	ParityRPCWSPort = "--ws-port"
	// ParityRPCWSHost is the argument used for RPC WS host
	ParityRPCWSHost = "--ws-interface"
	// ParityRPCWSAPI is the argument used for RPC WS APIs
	ParityRPCWSAPI = "--ws-apis"
	// ParityRPCWSWhitelist is the argument used for whitelisting hosts for WS server
	ParityRPCWSWhitelist = "--ws-hosts"

	// ParityUnlock is the argument used for unlocking imported ethereum account
	ParityUnlock = "--unlock"
	// ParityPassword is the argument used for locking imported ethereum address
	ParityPassword = "--password"
)

// Nethermind client arguments
const (
	// NethermindLogging is the argument used for logging verbosity level
	NethermindLogging = "--log"
	// NethermindNodePrivateKey is the argument used for node private key
	NethermindNodePrivateKey = "--KeyStore.EnodeKeyFile"
	// NethermindStaticNodesFile is the argument used to locate static nodes file
	NethermindStaticNodesFile = "--Init.StaticNodesPath"
	// NethermindBootnodes is the argument used to set bootnodes
	NethermindBootnodes = "--Discovery.Bootnodes"
	// NethermindGenesisFile is the argument used for genesis file
	NethermindGenesisFile = "--Init.ChainSpecPath"
	// NethermindDataPath is the argument used for data path
	NethermindDataPath = "--datadir"
	// NethermindNetwork is the argument used for selecting network
	NethermindNetwork = "--config"
	// NethermindDiscoveryEnabled is the argument used to enabled discovery
	NethermindDiscoveryEnabled = "--Init.DiscoveryEnabled"
	// NethermindP2PPort is the argument used for p2p port
	NethermindP2PPort = "--Network.P2PPort"
	// NethermindFastSync is the argument used to enable beam sync
	NethermindFastSync = "--Sync.FastSync"
	// NethermindFastBlocks is the argument used to enable fast blocks sync
	NethermindFastBlocks = "--Sync.FastBlocks"
	// NethermindDownloadBodiesInFastSync is the argument used to enable downloading block bodies in fast sync
	NethermindDownloadBodiesInFastSync = "--Sync.DownloadBodiesInFastSync"
	// NethermindDownloadReceiptsInFastSync is the argument used to enable downloading block receipts in fast sync
	NethermindDownloadReceiptsInFastSync = "--Sync.DownloadReceiptsInFastSync"
	// NethermindDownloadHeadersInFastSync is the argument used to enable downloading block headers in fast sync
	NethermindDownloadHeadersInFastSync = "--Sync.DownloadHeadersInFastSync"
	// NethermindMinerCoinbase is the argument used for setting coinbase account
	NethermindMinerCoinbase = "--KeyStore.BlockAuthorAccount"
	// NethermindRPCHTTPEnabled is the argument used to enable RPC over HTTP
	NethermindRPCHTTPEnabled = "--JsonRpc.Enabled"
	// NethermindRPCHTTPHost is the argument used for RPC HTTP Host
	NethermindRPCHTTPHost = "--JsonRpc.Host"
	// NethermindRPCHTTPPort is the argument used for RPC HTTP port
	NethermindRPCHTTPPort = "--JsonRpc.Port"
	// NethermindRPCHTTPAPI is the argument used for RPC HTTP APIs
	NethermindRPCHTTPAPI = "--JsonRpc.EnabledModules"

	// NethermindRPCEnginePort is the argument used to set engine API listening port
	NethermindRPCEnginePort = "--JsonRpc.EnginePort"
	// NethermindRPCEngineHost is the argument used to set engine API listening address
	NethermindRPCEngineHost = "--JsonRpc.EngineHost"
	// NethermindRPCJwtSecretFile is the argument used to locate jwt secret file
	NethermindRPCJwtSecretFile = "--JsonRpc.JwtSecretFile"

	// NethermindRPCWSEnabled is the argument used to enable RPC WS
	NethermindRPCWSEnabled = "--Init.WebSocketsEnabled"
	// NethermindRPCWSPort is the argument used for RPC WS port
	NethermindRPCWSPort = "--JsonRpc.WebSocketsPort"
	// NethermindUnlockAccounts is the argument used to unlock accounts
	NethermindUnlockAccounts = "--KeyStore.UnlockAccounts"
	// NethermindPasswordFiles is the argument used locate password files for unlocked accounts
	NethermindPasswordFiles = "--KeyStore.PasswordFiles"
	// NethermindMiningEnabled is the argument used for turning on mining
	NethermindMiningEnabled = "--Mining.Enabled"
)
