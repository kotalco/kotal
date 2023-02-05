package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeSpec defines the desired state of Node
type NodeSpec struct {
	// Image is NEAR node client image
	Image string `json:"image,omitempty"`
	// Network is NEAR network to join and sync
	// +kubebuilder:validation:Enum=mainnet;testnet;betanet
	Network string `json:"network"`
	// NodePrivateKeySecretName is the secret name holding node Ed25519 private key
	NodePrivateKeySecretName string `json:"nodePrivateKeySecretName,omitempty"`
	// ValidatorSecretName is the secret name holding node Ed25519 validator key
	ValidatorSecretName string `json:"validatorSecretName,omitempty"`
	// MinPeers is minimum number of peers to start syncing/producing blocks
	MinPeers uint `json:"minPeers,omitempty"`
	// Archive keeps old blocks in the storage
	Archive bool `json:"archive,omitempty"`
	// P2PPort is p2p port
	P2PPort uint `json:"p2pPort,omitempty"`
	// RPC enables JSON-RPC server
	RPC bool `json:"rpc,omitempty"`
	// RPCPort is JSON-RPC server listening port
	RPCPort uint `json:"rpcPort,omitempty"`
	// PrometheusPort is prometheus exporter port
	PrometheusPort uint `json:"prometheusPort,omitempty"`
	// TelemetryURL is telemetry service URL
	TelemetryURL string `json:"telemetryURL,omitempty"`
	// Bootnodes is array of boot nodes to bootstrap network from
	// +listType=set
	Bootnodes []string `json:"bootnodes,omitempty"`
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
// +kubebuilder:printcolumn:name="Validator",type=boolean,JSONPath=".spec.validator",priority=10
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
