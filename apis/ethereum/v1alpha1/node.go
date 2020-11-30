package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeStatus defines the observed state of Node
type NodeStatus struct {
	// EnodeURL is the node URL
	EnodeURL string `json:"enodeURL,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Node is the Schema for the nodes API
// +kubebuilder:printcolumn:name="Consensus",type=string,JSONPath=".spec.consensus"
// +kubebuilder:printcolumn:name="Join",type=string,JSONPath=".spec.join"
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

//NodeSpec is the specification of the node
type NodeSpec struct {
	NetworkConfig      `json:",inline"`
	AvailabilityConfig `json:",inline"`

	// Client is ethereum client running on the node
	Client EthereumClient `json:"client,omitempty"`

	// Name is the node name
	Name string `json:"name"`

	// import is account to import
	Import *ImportedAccount `json:"import,omitempty"`

	// Bootnode is whether node is bootnode or no
	Bootnode bool `json:"bootnode,omitempty"`

	// Nodekey is the node private key
	Nodekey PrivateKey `json:"nodekey,omitempty"`

	// StaticNodes is a set of ethereum nodes to maintain connection to
	StaticNodes []Enode `json:"staticNodes,omitempty"`

	// P2PPort is port used for peer to peer communication
	P2PPort uint `json:"p2pPort,omitempty"`

	// SyncMode is the node synchronization mode
	SyncMode SynchronizationMode `json:"syncMode,omitempty"`

	// Miner is whether node is mining/validating blocks or no
	Miner bool `json:"miner,omitempty"`

	// Logging is logging verboisty level
	Logging VerbosityLevel `json:"logging,omitempty"`

	// Coinbase is the account to which mining rewards are paid
	Coinbase EthereumAddress `json:"coinbase,omitempty"`

	// Hosts is a list of hostnames to to whitelist for RPC access
	Hosts []string `json:"hosts,omitempty"`

	// CORSDomains is the domains from which to accept cross origin requests
	CORSDomains []string `json:"corsDomains,omitempty"`

	// RPC is whether HTTP-RPC server is enabled or not
	RPC bool `json:"rpc,omitempty"`

	// RPCPort is HTTP-RPC server listening port
	RPCPort uint `json:"rpcPort,omitempty"`

	// RPCAPI is a list of rpc services to enable
	RPCAPI []API `json:"rpcAPI,omitempty"`

	// WS is whether web socket server is enabled or not
	WS bool `json:"ws,omitempty"`

	// WSPort is the web socket server listening port
	WSPort uint `json:"wsPort,omitempty"`

	// WSAPI is a list of WS services to enable
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
// +kubebuilder:validation:Enum=fast;full;light
type SynchronizationMode string

const (
	//FastSynchronization is the fast synchronization mode
	FastSynchronization SynchronizationMode = "fast"

	//LightSynchronization is the light synchronization mode
	LightSynchronization SynchronizationMode = "light"

	//FullSynchronization is the fast (non-archival) synchronization mode
	FullSynchronization SynchronizationMode = "full"
)

// VerbosityLevel is logging verbosity levels
// +kubebuilder:validation:Enum=off;fatal;error;warn;info;debug;trace;all
type VerbosityLevel string

const (
	// NoLogs outputs no logs
	NoLogs VerbosityLevel = "off"
	// FatalLogs outputs only fatal logs
	FatalLogs VerbosityLevel = "fatal"
	// ErrorLogs outputs only error logs
	ErrorLogs VerbosityLevel = "error"
	// WarnLogs outputs only warning logs
	WarnLogs VerbosityLevel = "warn"
	// InfoLogs outputs only informational logs
	InfoLogs VerbosityLevel = "info"
	// DebugLogs outputs only debugging logs
	DebugLogs VerbosityLevel = "debug"
	// TraceLogs outputs only tracing logs
	TraceLogs VerbosityLevel = "trace"
	// AllLogs outputs only all logs
	AllLogs VerbosityLevel = "all"
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
// +kubebuilder:validation:Enum=besu;geth;parity
type EthereumClient string

const (
	// BesuClient is hyperledger besu ethereum client
	BesuClient EthereumClient = "besu"
	// GethClient is go ethereum client
	GethClient EthereumClient = "geth"
	// ParityClient is Parity (OpenEthereum) client
	ParityClient EthereumClient = "parity"
)

// ImportedAccount is account derived from private key
type ImportedAccount struct {
	// Privatekey is the account private key
	PrivateKey `json:"privatekey"`
	// Password is the password used to encrypt account private key
	Password string `json:"password"`
}

func init() {
	SchemeBuilder.Register(&Node{}, &NodeList{})
}
