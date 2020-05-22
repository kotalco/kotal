/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NetworkSpec defines the desired state of Network
type NetworkSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Join specifies the network to join
	Join string `json:"join,omitempty"`

	// Consensus is the consensus algorithm to be used by the network nodes to reach consensus
	Consensus ConsensusAlgorithm `json:"consensus,omitempty"`

	// Genesis is genesis block specification
	Genesis *Genesis `json:"genesis,omitempty"`

	// Nodes is array of node specifications
	// +kubebuilder:validation:MinItems=1
	Nodes []Node `json:"nodes"`
}

// HexString is String in hexadecial format
// +kubebuilder:validation:Pattern=0[xX][0-9a-fA-F]+
type HexString string

// EthereumAddress is ethereum address
// +kubebuilder:validation:Pattern="0[xX][0-9a-fA-F]{40}"
type EthereumAddress string

// Genesis is genesis block sepcficition
type Genesis struct {
	// Accounts is array of accounts to fund or associate with code and storage
	Accounts []Account `json:"accounts,omitempty"`

	// ChainID is the the chain ID used in transaction signature to prevent reply attack
	// more details https://github.com/ethereum/EIPs/blob/master/EIPS/eip-155.md
	ChainID uint `json:"chainId"`

	// Address to pay mining rewards to
	Coinbase EthereumAddress `json:"coinbase,omitempty"`

	// Difficulty is the diffculty of the genesis block
	Difficulty HexString `json:"difficulty,omitempty"`

	// MixHash is hash combined with nonce to prove effort spent to create block
	MixHash HexString `json:"mixHash,omitempty"`

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
	Validators []EthereumAddress `json:"validators,omitempty"`

	// RequestTimeout is the timeout for each consensus round in seconds
	RequestTimeout uint `json:"requestTimeout,omitempty"`

	// MessageQueueLimit is the message queue limit
	MessageQueueLimit uint `json:"messageQueueLimit,omitempty"`

	// DuplicateMesageLimit is duplicate messages limit
	DuplicateMesageLimit uint `json:"duplicateMesageLimit,omitempty"`

	// futureMessagesLimit is future messages buffer limit
	FutureMessagesLimit uint `json:"futureMessagesLimit,omitempty"`

	// FutureMessagesMaxDistance is maximum height from current chain height for buffering future messages
	FutureMessagesMaxDistance uint `json:"futureMessagesMaxDistance,omitempty"`
}

// Clique configuration
type Clique struct {
	PoA `json:",inline"`

	// InitialSigners are PoA initial signers, at least one signer is required
	// +kubebuilder:validation:MinItems=1
	InitialSigners []EthereumAddress `json:"initialSigners,omitempty"`
}

// Signer is ethereum node address
// +kubebuilder:validation:Pattern=0[xX][0-9a-fA-F]+
type Signer string

// Validator is ethereum node address
// +kubebuilder:validation:Pattern=0[xX][0-9a-fA-F]+
type Validator string

// Ethash configurations
type Ethash struct {
	// FixedDifficulty is fixed difficulty to be used in private PoW networks
	FixedDifficulty uint `json:"fixedDifficulty,omitempty"`
}

// Forks is the supported forks by the network
type Forks struct {
	// Homestead fork
	Homestead uint `json:"homestead,omitempty"`

	// DAO fork
	DAO uint `json:"dao,omitempty"`

	// EIP150 (Tangerine Whistle) fork
	EIP150 uint `json:"eip150,omitempty"`

	// EIP155 (Spurious Dragon) fork
	EIP155 uint `json:"eip155,omitempty"`

	// EIP158 (Tangerine Whistle) fork
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
}

// Account is Ethereum account
type Account struct {
	// Address is account address
	Address EthereumAddress `json:"address"`

	// Balance is account balance in wei
	Balance HexString `json:"balance,omitempty"`

	// Code is account contract byte code
	Code HexString `json:"code,omitempty"`

	// Storage is account contract storage as key value pair
	Storage map[HexString]HexString `json:"storage,omitempty"`
}

// ConsensusAlgorithm is the algorithm nodes use to reach consensus
// +kubebuilder:validation:Enum=poa;pow;ibft2;quorum
type ConsensusAlgorithm string

const (
	// ProofOfAuthority is proof of authority consensus algorithm
	ProofOfAuthority ConsensusAlgorithm = "poa"

	// ProofOfWork is proof of work (nakamoto consensus) consensus algorithm
	ProofOfWork ConsensusAlgorithm = "pow"

	// IstanbulBFT is Istanbul Byzantine Fault Tolerant consensus algorithm
	IstanbulBFT ConsensusAlgorithm = "ibft2"

	//Quorum is Quorum IBFT consensus algorithm
	Quorum ConsensusAlgorithm = "quorum"
)

// SynchronizationMode is the node synchronization mode
// +kubebuilder:validation:Enum=fast;full;archive
type SynchronizationMode string

// String returns the string value of the synchronization mode
func (sm SynchronizationMode) String() string {
	return string(sm)
}

const (
	//FastSynchronization is the full (archive) synchronization mode, alias for archive
	FastSynchronization SynchronizationMode = "fast"

	//ArchiveSynchronization is the archive synchronization mode, alias for full
	ArchiveSynchronization SynchronizationMode = "archive"

	//FullSynchronization is the fast (non-archival) synchronization mode
	FullSynchronization SynchronizationMode = "full"
)

// API is RPC API to be exposed by RPC or web socket server
// +kubebuilder:validation:Enum=admin;clique;debug;eea;eth;ibft;miner;net;perm;plugins;priv;txpool;web3
type API string

// String returns string value of the RPC service
func (a API) String() string {
	return string(a)
}

const (
	// AdminAPI is administration API
	AdminAPI API = "admin"

	// CliqueAPI is clique (Proof of Authority consensus) API
	CliqueAPI API = "clique"

	// DebugAPI is debugging API
	DebugAPI API = "debug"

	// EEAAPI is EEA (Enterprise Ethereum Alliance) API
	EEAAPI API = "eea"

	// ETHAPI is ethereum API
	ETHAPI API = "eth"

	// IBFTAPI is IBFT consensus API
	IBFTAPI API = "ibft"

	// MinerAPI is miner API
	MinerAPI API = "miner"

	// NetworkAPI is network API
	NetworkAPI API = "net"

	// PermissionAPI is permission API
	PermissionAPI API = "perm"

	// PluginsAPI is plugins API
	PluginsAPI API = "plugins"

	// PrivacyAPI is privacy API
	PrivacyAPI API = "privacy"

	// TransactionPoolAPI is transaction pool API
	TransactionPoolAPI API = "txpool"

	// Web3API is web3 API
	Web3API API = "web3"
)

//Node is the specification of the node
type Node struct {
	// Name is the node name
	Name string `json:"name"`

	// Nodekey is the node private key
	Nodekey string `json:"nodekey,omitempty"`

	// P2PPort is port used for peer to peer communication
	P2PPort uint `json:"p2pPort,omitempty"`

	// SyncMode is the node synchronization mode
	SyncMode SynchronizationMode `json:"syncMode,omitempty"`

	// Miner is whether node is mining/validating blocks or no
	Miner bool `json:"miner,omitempty"`

	// MinerAccount is the account to which mining rewards are paid
	MinerAccount EthereumAddress `json:"minerAccount,omitempty"`

	// Hosts is a list of hostnames to to whitelist for RPC access
	Hosts []string `json:"hosts,omitempty"`

	// CORSDomains is the domains from which to accept cross origin requests
	CORSDomains []string `json:"corsDomains,omitempty"`

	// RPC is whether HTTP-RPC server is enabled or not
	RPC bool `json:"rpc,omitempty"`

	// RPCHost is HTTP-RPC server host address
	RPCHost string `json:"rpcHost,omitempty"`

	// RPCPort is HTTP-RPC server listening port
	RPCPort uint `json:"rpcPort,omitempty"`

	// RPCAPI is a list of rpc services to enable
	RPCAPI []API `json:"rpcAPI,omitempty"`

	// WS is whether web socket server is enabled or not
	WS bool `json:"ws,omitempty"`

	// WSHost is HTTP-WS server host address
	WSHost string `json:"wsHost,omitempty"`

	// WSPort is the web socket server listening port
	WSPort uint `json:"wsPort,omitempty"`

	// WSAPI is a list of WS services to enable
	WSAPI []API `json:"wsAPI,omitempty"`

	// GraphQL is whether GraphQL server is enabled or not
	GraphQL bool `json:"graphql,omitempty"`

	// GraphQLHost is GraphQL server host address
	GraphQLHost string `json:"graphqlHost,omitempty"`

	// GraphQLPort is the GraphQL server listening port
	GraphQLPort uint `json:"graphqlPort,omitempty"`
}

// NetworkStatus defines the observed state of Network
type NetworkStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Network is the Schema for the networks API
type Network struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkSpec   `json:"spec,omitempty"`
	Status NetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkList contains a list of Network
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Network `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Network{}, &NetworkList{})
}
