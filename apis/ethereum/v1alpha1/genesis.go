package v1alpha1

import "github.com/kotalco/kotal/apis/shared"

// Genesis is genesis block sepcficition
type Genesis struct {
	// Accounts is array of accounts to fund or associate with code and storage
	Accounts []Account `json:"accounts,omitempty"`

	// NetworkID is network id
	NetworkID uint `json:"networkId"`

	// ChainID is the the chain ID used in transaction signature to prevent reply attack
	// more details https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md
	ChainID uint `json:"chainId"`

	// Address to pay mining rewards to
	Coinbase shared.EthereumAddress `json:"coinbase,omitempty"`

	// Difficulty is the diffculty of the genesis block
	Difficulty HexString `json:"difficulty,omitempty"`

	// MixHash is hash combined with nonce to prove effort spent to create block
	MixHash Hash `json:"mixHash,omitempty"`

	// Ethash PoW engine configuration
	Ethash *Ethash `json:"ethash,omitempty"`

	// Clique PoA engine cinfiguration
	Clique *Clique `json:"clique,omitempty"`

	// IBFT2 PoA engine configuration
	IBFT2 *IBFT2 `json:"ibft2,omitempty"`

	// Forks is supported forks (network upgrade) and corresponding block number
	Forks *Forks `json:"forks,omitempty"`

	// GastLimit is the total gas limit for all transactions in a block
	GasLimit HexString `json:"gasLimit,omitempty"`

	// Nonce is random number used in block computation
	Nonce HexString `json:"nonce,omitempty"`

	// Timestamp is block creation date
	Timestamp HexString `json:"timestamp,omitempty"`
}

// PoA is Shared PoA engine config
type PoA struct {
	// BlockPeriod is block time in seconds
	BlockPeriod uint `json:"blockPeriod,omitempty"`

	// EpochLength is the Number of blocks after which to reset all votes
	EpochLength uint `json:"epochLength,omitempty"`
}

// IBFT2 configuration
type IBFT2 struct {
	PoA `json:",inline"`

	// Validators are initial ibft2 validators
	// +kubebuilder:validation:MinItems=1
	Validators []shared.EthereumAddress `json:"validators,omitempty"`

	// RequestTimeout is the timeout for each consensus round in seconds
	RequestTimeout uint `json:"requestTimeout,omitempty"`

	// MessageQueueLimit is the message queue limit
	MessageQueueLimit uint `json:"messageQueueLimit,omitempty"`

	// DuplicateMessageLimit is duplicate messages limit
	DuplicateMessageLimit uint `json:"duplicateMessageLimit,omitempty"`

	// futureMessagesLimit is future messages buffer limit
	FutureMessagesLimit uint `json:"futureMessagesLimit,omitempty"`

	// FutureMessagesMaxDistance is maximum height from current chain height for buffering future messages
	FutureMessagesMaxDistance uint `json:"futureMessagesMaxDistance,omitempty"`
}

// Clique configuration
type Clique struct {
	PoA `json:",inline"`

	// Signers are PoA initial signers, at least one signer is required
	// +kubebuilder:validation:MinItems=1
	Signers []shared.EthereumAddress `json:"signers,omitempty"`
}

// Ethash configurations
type Ethash struct {
	// FixedDifficulty is fixed difficulty to be used in private PoW networks
	FixedDifficulty *uint `json:"fixedDifficulty,omitempty"`
}

// Forks is the supported forks by the network
type Forks struct {
	// Homestead fork
	Homestead uint `json:"homestead,omitempty"`

	// DAO fork
	DAO *uint `json:"dao,omitempty"`

	// EIP150 (Tangerine Whistle) fork
	EIP150 uint `json:"eip150,omitempty"`

	// EIP155 (Spurious Dragon) fork
	EIP155 uint `json:"eip155,omitempty"`

	// EIP158 (state trie clearing) fork
	EIP158 uint `json:"eip158,omitempty"`

	// Byzantium fork
	Byzantium uint `json:"byzantium,omitempty"`

	// Constantinople fork
	Constantinople uint `json:"constantinople,omitempty"`

	// Petersburg fork
	Petersburg uint `json:"petersburg,omitempty"`

	// Istanbul fork
	Istanbul uint `json:"istanbul,omitempty"`

	// MuirGlacier fork
	MuirGlacier uint `json:"muirglacier,omitempty"`

	// Berlin fork
	Berlin uint `json:"berlin,omitempty"`

	// London fork
	London uint `json:"london,omitempty"`

	// ArrowGlacier fork
	ArrowGlacier uint `json:"arrowGlacier,omitempty"`
}

// Account is Ethereum account
type Account struct {
	// Address is account address
	Address shared.EthereumAddress `json:"address"`

	// Balance is account balance in wei
	Balance HexString `json:"balance,omitempty"`

	// Code is account contract byte code
	Code HexString `json:"code,omitempty"`

	// Storage is account contract storage as key value pair
	Storage map[HexString]HexString `json:"storage,omitempty"`
}
