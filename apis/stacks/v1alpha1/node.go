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
	// Image is Stacks node client image
	Image string `json:"image,omitempty"`
	// Network is stacks network
	// +kubebuilder:validation:Enum=mainnet;testnet
	Network StacksNetwork `json:"network"`
	// RPC enables JSON-RPC server
	RPC bool `json:"rpc,omitempty"`
	// RPCPort is JSON-RPC server port
	RPCPort uint `json:"rpcPort,omitempty"`
	// P2PPort is p2p bind port
	P2PPort uint `json:"p2pPort,omitempty"`
	// BitcoinNode is Bitcoin node
	BitcoinNode BitcoinNode `json:"bitcoinNode"`
	// Miner enables mining
	Miner bool `json:"miner,omitempty"`
	// SeedPrivateKeySecretName is k8s secret holding seed private key used for mining
	SeedPrivateKeySecretName string `json:"seedPrivateKeySecretName,omitempty"`
	// MineMicroblocks mines Stacks micro blocks
	MineMicroblocks bool `json:"mineMicroblocks,omitempty"`
	// NodePrivateKeySecretName is k8s secret holding node private key
	NodePrivateKeySecretName string `json:"nodePrivateKeySecretName,omitempty"`
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
// +kubebuilder:printcolumn:name="Miner",type=boolean,JSONPath=".spec.miner"
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
