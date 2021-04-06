package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PeerSpec defines the desired state of Peer
type PeerSpec struct {
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// PeerStatus defines the observed state of Peer
type PeerStatus struct {
}

// +kubebuilder:object:root=true

// Peer is the Schema for the peers API
type Peer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PeerSpec   `json:"spec,omitempty"`
	Status PeerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PeerList contains a list of Peer
type PeerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Peer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Peer{}, &PeerList{})
}
