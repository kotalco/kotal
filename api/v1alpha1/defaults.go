package v1alpha1

var (
	// DefaultAPIs is the default rpc, ws APIs
	DefaultAPIs []API = []API{Web3API, ETHAPI, NetworkAPI}
	// DefaultOrigins is the default origins
	DefaultOrigins []string = []string{"*"}
)

// Network defaults
const (
	DefaultTopologyKey = "topology.kubernetes.io/zone"
)

// Node defaults
const (
	// DefaultClient is the default ethereum client
	DefaultClient = BesuClient
	// DefaultHost is the default host
	DefaultHost = "0.0.0.0"
	// DefaultP2PPort is the default p2p port
	DefaultP2PPort uint = 30303
	// DefaultSyncMode is the default sync mode
	DefaultSyncMode = FullSynchronization
	// DefaultRPCPort is the default rpc port
	DefaultRPCPort uint = 8545
	// DefaultWSPort is the default ws port
	DefaultWSPort uint = 8546
	// DefaultGraphQLPort is the default graphQL port
	DefaultGraphQLPort uint = 8547
)

// Genesis block defaults
const (
	// DefaultCoinbase is the default coinbase
	DefaultCoinbase = EthereumAddress("0x0000000000000000000000000000000000000000")
	// DefaultDifficulty is the default difficulty
	DefaultDifficulty = HexString("0x1")
	// DefaultMixHash is the default mix hash
	DefaultMixHash = Hash("0x0000000000000000000000000000000000000000000000000000000000000000")
	// DefaultGasLimit is the default gas limit
	DefaultGasLimit = HexString("0x47b760")
	// DefaultNonce is the default nonce
	DefaultNonce = HexString("0x0")
	// DefaultTimestamp is the default timestamp
	DefaultTimestamp = HexString("0x0")
	// DefaultEIP150Hash is the default eip150 hash
	DefaultEIP150Hash = Hash("0x2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f0")
)

// Ethash engine defaults
const (
	// DefaultEthashFixedDifficulty is the default ethash fixed difficulty
	DefaultEthashFixedDifficulty uint = 1000
)

// Clique engine defaults
const (
	// DefaultCliqueBlockPeriod is the default clique block period
	DefaultCliqueBlockPeriod uint = 15
	// DefaultCliqueEpochLength is th default clique epoch length
	DefaultCliqueEpochLength uint = 3000
)

// IBFT2 engine defaults
const (
	// DefaultIBFT2BlockPeriod is the default ibft2 block period
	DefaultIBFT2BlockPeriod uint = 15
	// DefaultIBFT2EpochLength is the default ibft2 epoch length
	DefaultIBFT2EpochLength uint = 3000
	// DefaultIBFT2RequestTimeout is the default ibft2 request timeout
	DefaultIBFT2RequestTimeout uint = 10
	// DefaultIBFT2MessageQueueLimit is the default ibft2 message queue limit
	DefaultIBFT2MessageQueueLimit uint = 1000
	// DefaultIBFT2DuplicateMesageLimit is the default ibft2 duplicate message limit
	DefaultIBFT2DuplicateMesageLimit uint = 100
	// DefaultIBFT2FutureMessagesLimit is the default ibft2 future message limit
	DefaultIBFT2FutureMessagesLimit uint = 1000
	// DefaultIBFT2FutureMessagesMaxDistance is the default ibft2 future message max distance
	DefaultIBFT2FutureMessagesMaxDistance uint = 10
)
