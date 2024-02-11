package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BitcoinNetwork is Bitcoin network
type BitcoinNetwork string

// Bitcoin networks
const (
	Mainnet BitcoinNetwork = "mainnet"
	Testnet BitcoinNetwork = "testnet"
)

// RPCUsers is JSON-RPC users credentials
type RPCUser struct {
	// Username is JSON-RPC username
	Username string `json:"username"`
	// PasswordSecretName is k8s secret name holding JSON-RPC user password
	PasswordSecretName string `json:"passwordSecretName"`
}

// NodeSpec defines the desired state of Node
type NodeSpec struct {
	// Image is Bitcoin node client image
	Image string `json:"image,omitempty"`
	// ExtraArgs is extra arguments to pass down to the cli
	ExtraArgs shared.ExtraArgs `json:"extraArgs,omitempty"`
	// Replicas is number of replicas
	// +kubebuilder:validation:Enum=0;1
	Replicas *uint `json:"replicas,omitempty"`
	// Network is Bitcoin network to join and sync
	// +kubebuilder:validation:Enum=mainnet;testnet
	Network BitcoinNetwork `json:"network"`
	// Listen accepts connections from outside
	Listen *bool `json:"listen,omitempty"`
	// P2PPort is p2p communications port
	P2PPort uint `json:"p2pPort,omitempty"`
	// RPC enables JSON-RPC server
	RPC bool `json:"rpc,omitempty"`
	// RPCPort is JSON-RPC server port
	RPCPort uint `json:"rpcPort,omitempty"`
	// RPCUsers is JSON-RPC users credentials
	RPCUsers []RPCUser `json:"rpcUsers,omitempty"`
	// RPCWhitelist is a list of whitelisted rpc method
	// +listType=set
	RPCWhitelist []string `json:"rpcWhitelist,omitempty"`
	// Wallet load wallet and enables wallet RPC calls
	Wallet bool `json:"wallet,omitempty"`
	// TransactionIndex maintains a full tx index
	TransactionIndex bool `json:"txIndex,omitempty"`
	// CoinStatsIndex maintains coinstats index used by the gettxoutsetinfo RPC
	CoinStatsIndex bool `json:"coinStatsIndex,omitempty"`
	// Pruning allows pruneblockchain RPC to delete specific blocks
	Pruning bool `json:"pruning,omitempty"`
	// BlocksOnly rejects transactions from network peers
	// https://bitcointalk.org/index.php?topic=1377345.0
	BlocksOnly bool `json:"blocksOnly,omitempty"`
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// NodeStatus defines the observed state of Node
type NodeStatus struct {
	Client string `json:"client,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Node is the Schema for the nodes API
// +kubebuilder:printcolumn:name="Network",type=string,JSONPath=".spec.network"
// +kubebuilder:printcolumn:name="Client",type=string,JSONPath=".status.client"
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

func init() {
	SchemeBuilder.Register(&Node{}, &NodeList{})
}
