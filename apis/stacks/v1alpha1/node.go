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

// NodeSpec defines the desired state of Node
type NodeSpec struct {
	// Network is stacks network
	// +kubebuilder:validation:Enum=mainnet;testnet
	Network StacksNetwork `json:"network"`
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
