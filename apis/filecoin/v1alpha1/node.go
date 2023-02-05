package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeSpec defines the desired state of Node
type NodeSpec struct {
	// Image is Filecoin node client image
	Image string `json:"image,omitempty"`
	// API enables API server
	API bool `json:"api,omitempty"`
	// APIPort is API server listening port
	APIPort uint `json:"apiPort,omitempty"`
	// APIRequestTimeout is API request timeout in seconds
	APIRequestTimeout uint `json:"apiRequestTimeout,omitempty"`
	// DisableMetadataLog disables metadata log
	DisableMetadataLog bool `json:"disableMetadataLog,omitempty"`
	// P2PPort is p2p port
	P2PPort uint `json:"p2pPort,omitempty"`
	// Network is the Filecoin network the node will join and sync
	Network FilecoinNetwork `json:"network"`
	// IPFSPeerEndpoint is ipfs peer endpoint
	IPFSPeerEndpoint string `json:"ipfsPeerEndpoint,omitempty"`
	// IPFSOnlineMode sets ipfs online mode
	IPFSOnlineMode bool `json:"ipfsOnlineMode,omitempty"`
	// IPFSForRetrieval uses ipfs for retrieval
	IPFSForRetrieval bool `json:"ipfsForRetrieval,omitempty"`
	// Logging is logging verboisty level
	// +kubebuilder:validation:Enum=error;warn;info;debug
	Logging shared.VerbosityLevel `json:"logging,omitempty"`
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// FilecoinNetwork is Filecoin network
// +kubebuilder:validation:Enum=mainnet;calibration
type FilecoinNetwork string

const (
	// MainNetwork is the Filecoin main network
	MainNetwork FilecoinNetwork = "mainnet"
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
