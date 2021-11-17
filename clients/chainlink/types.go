package chainlink

// environment variables
const (
	// EnvRoot is the environment variable for root directory
	EnvRoot = "ROOT"
	// EnvChainID is the environment variable for ethereum chain id
	EnvChainID = "ETH_CHAIN_ID"
	// EnvEthereumURL is the environment variable for ethereum websocket url
	EnvEthereumURL = "ETH_URL"
	// EnvLinkContractAddress is the environment variable for chainlink contract address
	EnvLinkContractAddress = "LINK_CONTRACT_ADDRESS"
	// EnvDatabaseURL is the environment variable for database connection string
	EnvDatabaseURL = "DATABASE_URL"
	// EnvTLSCertPath is the environment variable for tls cert
	EnvTLSCertPath = "TLS_CERT_PATH"
	// EnvTLSKeyPath is the environment variable for tls key
	EnvTLSKeyPath = "TLS_KEY_PATH"
	// EnvTLSPort is the environment variable for tls port
	EnvTLSPort = "CHAINLINK_TLS_PORT"
	// EnvPort is the environment variable for API and GUI port
	EnvPort = "CHAINLINK_PORT"
	// EnvHTTPURL is the environment variable for http url
	EnvHTTPURL = "ETH_HTTP_URL"
	// EnvSecondaryURLs is the environment variable for extra http urls
	EnvSecondaryURLs = "ETH_SECONDARY_URLS"
	// EnvLogLevel is the environment variable for logging verbosity
	EnvLogLevel = "LOG_LEVEL"
	// EnvAllowOrigins is the environment variable for allowing cross origin requests from domains
	EnvAllowOrigins = "ALLOW_ORIGINS"
	// EnvSecureCookies is the environment variable for allowing cross origin requests from domains
	EnvSecureCookies = "SECURE_COOKIES"
	// EnvP2PListenPort is the environment variable for allowing cross origin requests from domains
	EnvP2PListenPort = "P2P_LISTEN_PORT"
)

// arguments
const (
	// ChainlinkPassword is the argument used to locate keystore password file
	ChainlinkPassword = "--password"
	// ChainlinkAPI is the argument used to locate api credentials file
	ChainlinkAPI = "--api"
)
