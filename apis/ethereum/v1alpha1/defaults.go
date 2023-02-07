package v1alpha1

import "github.com/kotalco/kotal/apis/shared"

var (
	// DefaultAPIs is the default rpc, ws APIs
	DefaultAPIs []API = []API{Web3API, ETHAPI, NetworkAPI}
	// DefaultOrigins is the default origins
	DefaultOrigins []string = []string{"*"}
)

const (
	// DefaultBesuImage is hyperledger besu image
	DefaultBesuImage = "hyperledger/besu:22.10.0"
	// DefaultGethImage is go-ethereum image
	DefaultGethImage = "kotalco/geth:v1.10.26"
	// DefaultNethermindImage is nethermind image
	DefaultNethermindImage = "kotalco/nethermind:v1.14.5"
)

// Node defaults
const (
	// DefaultLogging is the default logging verbosity level
	DefaultLogging = shared.InfoLogs
	// DefaultP2PPort is the default p2p port
	DefaultP2PPort uint = 30303
	// DefaultPublicNetworkSyncMode is the default sync mode for public networks
	DefaultPublicNetworkSyncMode = FastSynchronization
	// DefaultPrivateNetworkSyncMode is the default sync mode for private networks
	DefaultPrivateNetworkSyncMode = FullSynchronization
	// DefaultEngineRPCPort is the default engine rpc port
	DefaultEngineRPCPort uint = 8551
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
	DefaultCoinbase = shared.EthereumAddress("0x0000000000000000000000000000000000000000")
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
	// DefaultIBFT2DuplicateMessageLimit is the default ibft2 duplicate message limit
	DefaultIBFT2DuplicateMessageLimit uint = 100
	// DefaultIBFT2FutureMessagesLimit is the default ibft2 future message limit
	DefaultIBFT2FutureMessagesLimit uint = 1000
	// DefaultIBFT2FutureMessagesMaxDistance is the default ibft2 future message max distance
	DefaultIBFT2FutureMessagesMaxDistance uint = 10
)

// Resources
const (
	// DefaultPrivateNetworkNodeCPURequest is the cpu requested by private network node
	DefaultPrivateNetworkNodeCPURequest = "2"
	// DefaultPrivateNetworkNodeCPULimit is the cpu limit for private network node
	DefaultPrivateNetworkNodeCPULimit = "3"
	// DefaultPublicNetworkNodeCPURequest is the cpu requested by public network node
	DefaultPublicNetworkNodeCPURequest = "4"
	// DefaultPublicNetworkNodeCPULimit is the cpu limit for public network node
	DefaultPublicNetworkNodeCPULimit = "6"
	// DefaultPrivateNetworkNodeMemoryRequest is the memory requested by private network node
	DefaultPrivateNetworkNodeMemoryRequest = "4Gi"
	// DefaultPrivateNetworkNodeMemoryLimit is the memory limit for private network node
	DefaultPrivateNetworkNodeMemoryLimit = "6Gi"
	// DefaultPublicNetworkNodeMemoryRequest is the Memory requested by public network node
	DefaultPublicNetworkNodeMemoryRequest = "8Gi"
	// DefaultPublicNetworkNodeMemoryLimit is the Memory limit for public network node
	DefaultPublicNetworkNodeMemoryLimit = "16Gi"
	// DefaultPrivateNetworkNodeStorageRequest is the Storage requested by private network node
	DefaultPrivateNetworkNodeStorageRequest = "100Gi"
	// DefaultMainNetworkFullNodeStorageRequest is the Storage requested by main network archive node
	DefaultMainNetworkFullNodeStorageRequest = "6Ti"
	// DefaultMainNetworkFastNodeStorageRequest is the Storage requested by main network full node
	DefaultMainNetworkFastNodeStorageRequest = "750Gi"
	// DefaultTestNetworkStorageRequest is the Storage requested by main network full node
	DefaultTestNetworkStorageRequest = "25Gi"
)
