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

// VerbosityLevel is logging verbosity levels
// +kubebuilder:validation:Enum=error;warn;info;debug;trace
type VerbosityLevel string

const (
	// ErrorLogs outputs only error logs
	ErrorLogs VerbosityLevel = "error"
	// WarnLogs outputs only warning logs
	WarnLogs VerbosityLevel = "warn"
	// InfoLogs outputs only informational logs
	InfoLogs VerbosityLevel = "info"
	// DebugLogs outputs only debugging logs
	DebugLogs VerbosityLevel = "debug"
	// TraceLogs outputs only tracing logs
	TraceLogs VerbosityLevel = "trace"
)

// NodeSpec defines the desired state of Node
type NodeSpec struct {
	// Network is the polkadot network/chain to join
	Network string `json:"network"`
	// NodePrivatekeySecretName is the secret name holding node Ed25519 private key
	NodePrivatekeySecretName string `json:"nodePrivatekeySecretName,omitempty"`
	// Validator enables validator mode
	Validator bool `json:"validator,omitempty"`
	// SyncMode is the blockchain synchronization mode
	SyncMode SynchronizationMode `json:"syncMode,omitempty"`
	// Logging is logging verboisty level
	Logging VerbosityLevel `json:"logging,omitempty"`
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
