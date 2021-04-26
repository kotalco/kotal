package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterPeerSpec defines the desired state of ClusterPeer
type ClusterPeerSpec struct {
	// ID is the the cluster peer id
	ID string `json:"id,omitempty"`
	// PrivatekeySecretName is k8s secret holding private key
	PrivatekeySecretName string `json:"privatekeySecretName,omitempty"`
	// TrustedPeers is CRDT trusted cluster peers who can manage the pinset
	TrustedPeers []string `json:"trustedPeers,omitempty"`
	// BootstrapPeers are ipfs cluster peers to connect to
	BootstrapPeers []string `json:"bootstrapPeers,omitempty"`
	// Consensus is ipfs cluster consensus algorithm
	Consensus ConsensusAlgorithm `json:"consensus,omitempty"`
	// ClusterSecretName is k8s secret holding cluster secret
	ClusterSecretName string `json:"clusterSecretName"`
	// PeerEndpoint is ipfs peer http API endpoint
	PeerEndpoint string `json:"peerEndpoint"`
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// ConsensusAlgorithm is IPFS cluster consensus algorithm
// +kubebuilder:validation:Enum=crdt;raft
type ConsensusAlgorithm string

const (
	// CRDT consensus algorithm
	CRDT ConsensusAlgorithm = "crdt"
	// Raft consensus algorithm
	Raft ConsensusAlgorithm = "raft"
)

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
