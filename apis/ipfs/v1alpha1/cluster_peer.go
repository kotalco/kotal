package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterPeerSpec defines the desired state of ClusterPeer
type ClusterPeerSpec struct {
	// Image is ipfs cluster peer client image
	Image string `json:"image,omitempty"`
	// ID is the the cluster peer id
	ID string `json:"id,omitempty"`
	// PrivateKeySecretName is k8s secret holding private key
	PrivateKeySecretName string `json:"privateKeySecretName,omitempty"`
	// TrustedPeers is CRDT trusted cluster peers who can manage the pinset
	// +listType=set
	TrustedPeers []string `json:"trustedPeers,omitempty"`
	// BootstrapPeers are ipfs cluster peers to connect to
	// +listType=set
	BootstrapPeers []string `json:"bootstrapPeers,omitempty"`
	// Consensus is ipfs cluster consensus algorithm
	Consensus ConsensusAlgorithm `json:"consensus,omitempty"`
	// ClusterSecretName is k8s secret holding cluster secret
	ClusterSecretName string `json:"clusterSecretName"`
	// PeerEndpoint is ipfs peer http API endpoint
	PeerEndpoint string `json:"peerEndpoint"`
	// Logging is logging verboisty level
	// +kubebuilder:validation:Enum=error;warn;info;debug
	Logging shared.VerbosityLevel `json:"logging,omitempty"`
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
	Client    string `json:"client"`
	Consensus string `json:"consensus"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ClusterPeer is the Schema for the clusterpeers API
// +kubebuilder:printcolumn:name="Client",type=string,JSONPath=".status.client"
// +kubebuilder:printcolumn:name="Consensus",type=string,JSONPath=".spec.consensus"
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
