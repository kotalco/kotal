package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeStatus defines the observed state of Node
type NodeStatus struct {
	// Consensus is network consensus algorithm
	Consensus string `json:"consensus,omitempty"`
	// Network is the network this node is joining
	Network string `json:"network,omitempty"`
	// EnodeURL is the node URL
	EnodeURL string `json:"enodeURL,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Node is the Schema for the nodes API
// +kubebuilder:printcolumn:name="Client",type=string,JSONPath=".spec.client"
// +kubebuilder:printcolumn:name="Consensus",type=string,JSONPath=".status.consensus"
// +kubebuilder:printcolumn:name="Network",type=string,JSONPath=".status.network"
// +kubebuilder:printcolumn:name="enodeURL",type=string,JSONPath=".status.enodeURL",priority=10
type Node struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeSpec   `json:"spec,omitempty"`
	Status NodeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NodeList contains a list of Node
type NodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Node `json:"items"`
}

// NodeSpec is the specification of the node
type NodeSpec struct {

	// Image is Ethereum node client image
	Image string `json:"image,omitempty"`

	// Genesis is genesis block configuration
	Genesis *Genesis `json:"genesis,omitempty"`

	// Network specifies the network to join
	Network string `json:"network,omitempty"`

	// Client is ethereum client running on the node
	Client EthereumClient `json:"client"`

	// import is account to import
	Import *ImportedAccount `json:"import,omitempty"`

	// Bootnodes is set of ethereum node URLS for p2p discovery bootstrap
	// +listType=set
	Bootnodes []Enode `json:"bootnodes,omitempty"`

	// NodePrivateKeySecretName is the secret name holding node private key
	NodePrivateKeySecretName string `json:"nodePrivateKeySecretName,omitempty"`

	// StaticNodes is a set of ethereum nodes to maintain connection to
	// +listType=set
	StaticNodes []Enode `json:"staticNodes,omitempty"`

	// P2PPort is port used for peer to peer communication
	P2PPort uint `json:"p2pPort,omitempty"`

	// SyncMode is the node synchronization mode
	SyncMode SynchronizationMode `json:"syncMode,omitempty"`

	// Miner is whether node is mining/validating blocks or no
	Miner bool `json:"miner,omitempty"`

	// Logging is logging verboisty level
	// +kubebuilder:validation:Enum=off;fatal;error;warn;info;debug;trace;all
	Logging shared.VerbosityLevel `json:"logging,omitempty"`

	// Coinbase is the account to which mining rewards are paid
	Coinbase shared.EthereumAddress `json:"coinbase,omitempty"`

	// Hosts is a list of hostnames to to whitelist for RPC access
	// +listType=set
	Hosts []string `json:"hosts,omitempty"`

	// CORSDomains is the domains from which to accept cross origin requests
	// +listType=set
	CORSDomains []string `json:"corsDomains,omitempty"`

	// Engine enables authenticated Engine RPC APIs
	Engine bool `json:"engine,omitempty"`

	// EnginePort is engine authenticated RPC APIs port
	EnginePort uint `json:"enginePort,omitempty"`

	// JWTSecretName is kubernetes secret name holding JWT secret
	JWTSecretName string `json:"jwtSecretName,omitempty"`

	// RPC is whether HTTP-RPC server is enabled or not
	RPC bool `json:"rpc,omitempty"`

	// RPCPort is HTTP-RPC server listening port
	RPCPort uint `json:"rpcPort,omitempty"`

	// RPCAPI is a list of rpc services to enable
	// +listType=set
	RPCAPI []API `json:"rpcAPI,omitempty"`

	// WS is whether web socket server is enabled or not
	WS bool `json:"ws,omitempty"`

	// WSPort is the web socket server listening port
	WSPort uint `json:"wsPort,omitempty"`

	// WSAPI is a list of WS services to enable
	// +listType=set
	WSAPI []API `json:"wsAPI,omitempty"`

	// GraphQL is whether GraphQL server is enabled or not
	GraphQL bool `json:"graphql,omitempty"`

	// GraphQLPort is the GraphQL server listening port
	GraphQLPort uint `json:"graphqlPort,omitempty"`

	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// Enode is ethereum node url
type Enode string

// SynchronizationMode is the node synchronization mode
// +kubebuilder:validation:Enum=fast;full;light;snap
type SynchronizationMode string

const (
	//SnapSynchronization is the snap synchronization mode
	SnapSynchronization SynchronizationMode = "snap"

	//FastSynchronization is the fast synchronization mode
	FastSynchronization SynchronizationMode = "fast"

	//LightSynchronization is the light synchronization mode
	LightSynchronization SynchronizationMode = "light"

	//FullSynchronization is full archival synchronization mode
	FullSynchronization SynchronizationMode = "full"
)

// API is RPC API to be exposed by RPC or web socket server
// +kubebuilder:validation:Enum=admin;clique;debug;eea;eth;ibft;miner;net;perm;plugins;priv;txpool;web3
type API string

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

// EthereumClient is the ethereum client running on a given node
// +kubebuilder:validation:Enum=besu;geth;nethermind
type EthereumClient string

func (e EthereumClient) SupportsVerbosityLevel(level shared.VerbosityLevel) bool {
	switch e {
	case BesuClient:
		switch level {
		case shared.NoLogs,
			shared.FatalLogs,
			shared.ErrorLogs,
			shared.WarnLogs,
			shared.InfoLogs,
			shared.DebugLogs,
			shared.TraceLogs,
			shared.AllLogs:
			return true
		}
	case GethClient:
		switch level {
		case shared.NoLogs,
			shared.ErrorLogs,
			shared.WarnLogs,
			shared.InfoLogs,
			shared.DebugLogs,
			shared.AllLogs:
			return true
		}
	case NethermindClient:
		switch level {
		case shared.ErrorLogs,
			shared.WarnLogs,
			shared.InfoLogs,
			shared.DebugLogs,
			shared.TraceLogs:
			return true
		}

	}
	return false
}

const (
	// BesuClient is hyperledger besu ethereum client
	BesuClient EthereumClient = "besu"
	// GethClient is go ethereum client
	GethClient EthereumClient = "geth"
	// NethermindClient is Nethermind .NET client
	NethermindClient EthereumClient = "nethermind"
)

// ImportedAccount is account derived from private key
type ImportedAccount struct {
	// PrivateKeySecretName is the secret name holding account private key
	PrivateKeySecretName string `json:"privateKeySecretName"`
	// PasswordSecretName is the secret holding password used to encrypt account private key
	PasswordSecretName string `json:"passwordSecretName"`
}

func init() {
	SchemeBuilder.Register(&Node{}, &NodeList{})
}
