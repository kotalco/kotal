package v1alpha1

import "fmt"

//Node is the specification of the node
type Node struct {
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

	// P2PPort is port used for peer to peer communication
	P2PPort uint `json:"p2pPort,omitempty"`

	// SyncMode is the node synchronization mode
	SyncMode SynchronizationMode `json:"syncMode,omitempty"`

	// Miner is whether node is mining/validating blocks or no
	Miner bool `json:"miner,omitempty"`

	// Coinbase is the account to which mining rewards are paid
	Coinbase EthereumAddress `json:"coinbase,omitempty"`

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

	// Resources is node compute and storage resources
	Resources *NodeResources `json:"resources,omitempty"`
}

// IsBootnode is whether node is bootnode or no
func (n *Node) IsBootnode() bool {
	return n.Bootnode
}

// WithNodekey is whether node is configured with private key
func (n *Node) WithNodekey() bool {
	return n.Nodekey != ""
}

// DeploymentName returns name to be used by node deployment
func (n *Node) DeploymentName(network string) string {
	return fmt.Sprintf("%s-%s", network, n.Name)
}

// PVCName returns name to be used by node pvc
func (n *Node) PVCName(network string) string {
	return n.DeploymentName(network) // same as deployment name
}

// SecretName returns name to be used by node secret
func (n *Node) SecretName(network string) string {
	return n.DeploymentName(network) // same as deployment name
}

// ServiceName returns name to be used by node service
func (n *Node) ServiceName(network string) string {
	return n.DeploymentName(network) // same as deployment name
}

// ImportedAccountName returns the imported account secret name
func (n *Node) ImportedAccountName(network string) string {
	return fmt.Sprintf("%s-%s-imported-account", network, n.Name)
}

// Labels to be used by node resources
func (n *Node) Labels(network string) map[string]string {
	return map[string]string{
		"name":     "node",
		"instance": n.Name,
		"network":  network,
	}
}

// NodeResources is node compute and storage resources
type NodeResources struct {
	// CPU is cpu cores the node requires
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*m?$"
	CPU string `json:"cpu,omitempty"`
	// CPULimit is cpu cores the node is limited to
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*m?$"
	CPULimit string `json:"cpuLimit,omitempty"`
	// Memory is memmory requirements
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*[KMGTPE]i$"
	Memory string `json:"memory,omitempty"`
	// MemoryLimit is cpu cores the node is limited to
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*[KMGTPE]i$"
	MemoryLimit string `json:"memoryLimit,omitempty"`
	// Storage is disk space storage requirements
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*[KMGTPE]i$"
	Storage string `json:"storage,omitempty"`
}

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
// +kubebuilder:validation:Enum=besu;geth
type EthereumClient string

const (
	// BesuClient is hyperledger besu ethereum client
	BesuClient EthereumClient = "besu"
	// GethClient is go ethereum client
	GethClient EthereumClient = "geth"
)

// ImportedAccount is account derived from private key
type ImportedAccount struct {
	// Privatekey is the account private key
	PrivateKey `json:"privatekey"`
	// Password is the password used to encrypt account private key
	Password string `json:"password"`
}
