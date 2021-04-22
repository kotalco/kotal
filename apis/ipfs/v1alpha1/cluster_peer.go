package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterPeerSpec defines the desired state of ClusterPeer
type ClusterPeerSpec struct {
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// ClusterPeerStatus defines the observed state of ClusterPeer
type ClusterPeerStatus struct {
}

// +kubebuilder:object:root=true

// ClusterPeer is the Schema for the clusterpeers API
type ClusterPeer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterPeerSpec   `json:"spec,omitempty"`
	Status ClusterPeerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterPeerList contains a list of ClusterPeer
type ClusterPeerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterPeer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterPeer{}, &ClusterPeerList{})
}
