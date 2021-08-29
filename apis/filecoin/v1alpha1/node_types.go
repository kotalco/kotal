package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeSpec defines the desired state of Node
type NodeSpec struct {
	// Network is the Filecoin network the node will join and sync
	Network FilecoinNetwork `json:"network"`
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// FilecoinNetwork is Filecoin network
// +kubebuilder:validation:Enum=mainnet;nerpa;calibration
type FilecoinNetwork string

const (
	// MainNetwork is the Filecoin main network
	MainNetwork FilecoinNetwork = "mainnet"
	// NerpaNetwork is the Filecoin main network
	NerpaNetwork FilecoinNetwork = "nerpa"
	// CalibrationNetwork is the Filecoin main network
	CalibrationNetwork FilecoinNetwork = "calibration"
)

// NodeStatus defines the observed state of Node
type NodeStatus struct {
	Client string `json:"client"`
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
