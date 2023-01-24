package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SynchronizationMode is the blockchain synchronization mode
// +kubebuilder:validation:Enum=fast;full
type SynchronizationMode string

const (
	//FastSynchronization is the fast synchronization mode
	FastSynchronization SynchronizationMode = "fast"

	//FullSynchronization is the full archival synchronization mode
	FullSynchronization SynchronizationMode = "full"
)

// NodeSpec defines the desired state of Node
type NodeSpec struct {
	// Image is Polkadot node client image
	Image string `json:"image,omitempty"`
	// Network is the polkadot network/chain to join
	Network string `json:"network"`
	// P2PPort is p2p protocol tcp port
	P2PPort uint `json:"p2pPort,omitempty"`
	// NodePrivateKeySecretName is the secret name holding node Ed25519 private key
	NodePrivateKeySecretName string `json:"nodePrivateKeySecretName,omitempty"`
	// Validator enables validator mode
	Validator bool `json:"validator,omitempty"`
	// SyncMode is the blockchain synchronization mode
	SyncMode SynchronizationMode `json:"syncMode,omitempty"`
	// Pruning keeps recent or all blocks
	Pruning *bool `json:"pruning,omitempty"`
	// RetainedBlocks is the number of blocks to keep state for
	RetainedBlocks uint `json:"retainedBlocks,omitempty"`
	// Logging is logging verboisty level
	// +kubebuilder:validation:Enum=error;warn;info;debug;trace
	Logging shared.VerbosityLevel `json:"logging,omitempty"`
	// Telemetry enables connecting to telemetry server
	Telemetry bool `json:"telemetry,omitempty"`
	// TelemetryURL is telemetry service URL
	TelemetryURL string `json:"telemetryURL,omitempty"`
	// Prometheus exposes a prometheus exporter endpoint.
	Prometheus bool `json:"prometheus,omitempty"`
	// PrometheusPort is prometheus exporter port
	PrometheusPort uint `json:"prometheusPort,omitempty"`
	// RPC enables JSON-RPC server
	RPC bool `json:"rpc,omitempty"`
	// RPCPort is JSON-RPC server port
	RPCPort uint `json:"rpcPort,omitempty"`
	// WS enables Websocket server
	WS bool `json:"ws,omitempty"`
	// WSPort is Websocket server port
	WSPort uint `json:"wsPort,omitempty"`
	// CORSDomains is browser origins allowed to access the JSON-RPC HTTP and WS servers
	// +listType=set
	CORSDomains []string `json:"corsDomains,omitempty"`
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// NodeStatus defines the observed state of Node
type NodeStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Node is the Schema for the nodes API
// +kubebuilder:printcolumn:name="Network",type=string,JSONPath=".spec.network"
// +kubebuilder:printcolumn:name="Validator",type=boolean,JSONPath=".spec.validator"
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
