package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NodeSpec defines the desired state of Node
type NodeSpec struct {
	// Network is NEAR network to join and sync
	// +kubebuilder:validation:Enum=mainnet;testnet;betanet
	Network string `json:"network"`
	// NodePrivateKeySecretName is the secret name holding node Ed25519 private key
	NodePrivateKeySecretName string `json:"nodePrivateKeySecretName,omitempty"`
	// MinPeers is minimum number of peers to start syncing/producing blocks
	MinPeers uint `json:"minPeers,omitempty"`
	// P2PPort is p2p port
	P2PPort uint `json:"p2pPort,omitempty"`
	// P2PHost is p2p host
	P2PHost string `json:"p2pHost,omitempty"`
	// RPC enables JSON-RPC server
	RPC bool `json:"rpc,omitempty"`
	// RPCPort is JSON-RPC server listening port
	RPCPort uint `json:"rpcPort,omitempty"`
	// RPCHost is JSON-RPC server listening host
	RPCHost string `json:"rpcHost,omitempty"`
	// PrometheusPort is prometheus exporter port
	PrometheusPort uint `json:"prometheusPort,omitempty"`
	// PrometheusHost is prometheus exporter host
	PrometheusHost string `json:"prometheusHost,omitempty"`
	// TelemetryURL is telemetry service URL
	TelemetryURL string `json:"telemetryURL,omitempty"`
	// Bootnodes is array of boot nodes to bootstrap network from
	Bootnodes []string `json:"bootnodes,omitempty"`
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// NodeStatus defines the observed state of Node
type NodeStatus struct {
}

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
