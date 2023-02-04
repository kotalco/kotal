package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BitcoinNetwork is Bitcoin network
type BitcoinNetwork string

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
	// Network is Bitcoin network to join and sync
	// +kubebuilder:validation:Enum=mainnet;testnet
	Network BitcoinNetwork `json:"network"`
	// P2PPort is p2p communications port
	P2PPort uint `json:"p2pPort,omitempty"`
	// RPC enables JSON-RPC server
	RPC bool `json:"rpc,omitempty"`
	// RPCPort is JSON-RPC server port
	RPCPort uint `json:"rpcPort,omitempty"`
	// RPCUsers is JSON-RPC users credentials
	RPCUsers []RPCUser `json:"rpcUsers,omitempty"`
	// Wallet load wallet and enables wallet RPC calls
	Wallet bool `json:"wallet,omitempty"`
	// TransactionIndex maintains a full tx index
	TransactionIndex bool `json:"txIndex,omitempty"`
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
