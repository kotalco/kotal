package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StacksNetwork is Stacks network
type StacksNetwork string

const (
	Mainnet StacksNetwork = "mainnet"
	Testnet StacksNetwork = "testnet"
)

// BitcoinNode is Bitcoin node
type BitcoinNode struct {
	// Endpoint is bitcoin node JSON-RPC endpoint
	Endpoint string `json:"endpoint"`
	// P2pPort is bitcoin node p2p port
	P2pPort uint `json:"p2pPort"`
	// RpcPort is bitcoin node JSON-RPC port
	RpcPort uint `json:"rpcPort"`
	// RpcUsername is bitcoin node JSON-RPC username
	RpcUsername string `json:"rpcUsername"`
	// RpcPasswordSecretName is k8s secret name holding bitcoin node JSON-RPC password
	RpcPasswordSecretName string `json:"rpcPasswordSecretName"`
}

// NodeSpec defines the desired state of Node
type NodeSpec struct {
	// Network is stacks network
	// +kubebuilder:validation:Enum=mainnet;testnet
	Network StacksNetwork `json:"network"`
	// RPCPort is JSON-RPC server port
	RPCPort uint `json:"rpcPort,omitempty"`
	// RPCHost is JSON-RPC server host
	RPCHost string `json:"rpcHost,omitempty"`
	// P2PPort is p2p bind port
	P2PPort uint `json:"p2pPort,omitempty"`
	// P2PHost is p2p bind host
	P2PHost string `json:"p2pHost,omitempty"`
	// BitcoinNode is Bitcoin node
	BitcoinNode BitcoinNode `json:"bitcoinNode"`
	// Miner enables mining
	Miner bool `json:"miner,omitempty"`
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// NodeStatus defines the observed state of Node
type NodeStatus struct{}

// +kubebuilder:object:root=true

// Node is the Schema for the nodes API
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
