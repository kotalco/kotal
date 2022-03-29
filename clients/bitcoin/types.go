package bitcoin

const (
	// EnvBitcoinData is environment variable used to set data directory
	EnvBitcoinData = "BITCOIN_DATA"
)

const (
	// BitcoinArgChain is argument used to set chain
	BitcoinArgChain = "-chain"
	// BitcoinArgBind is argument used to bind and listen to the given address
	BitcoinArgBind = "-bind"
	// BitcoinArgServer is argument used to enable CLI and JSON-RPC server
	BitcoinArgServer = "-server"
	// BitcoinArgRPCPort is argument used to set JSON-RPC port
	BitcoinArgRPCPort = "-rpcport"
	// BitcoinArgDataDir is argument used to set data directory
	BitcoinArgDataDir = "-datadir"
	// BitcoinArgRPCBind is argument used to set JSON-RPC server host
	BitcoinArgRPCBind = "-rpcbind"
	// BitcoinArgRPCAllowIp is argument used to allow JSON-RPC connections from specific sources
	BitcoinArgRPCAllowIp = "-rpcallowip"
	// BitcoinArgRPCAuth is argument used to set JSON-RPC user and password in the format of user:salt$hash
	BitcoinArgRPCAuth = "-rpcauth"
	// BitcoinArgDisableWallet is argument used to disable wallet and RPC calls
	BitcoinArgDisableWallet = "-disablewallet"
)
