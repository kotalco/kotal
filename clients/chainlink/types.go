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
	// EnvHTTPURL is the environment variable for http url
	EnvHTTPURL = "ETH_HTTP_URL"
	// EnvSecondaryURLs is the environment variable for extra http urls
	EnvSecondaryURLs = "ETH_SECONDARY_URLS"
)

// arguments
const (
	// ChainlinkPassword is the argument used to locate keystore password file
	ChainlinkPassword = "--password"
	// ChainlinkAPI is the argument used to locate api credentials file
	ChainlinkAPI = "--api"
)
