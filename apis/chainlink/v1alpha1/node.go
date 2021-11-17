package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// APICredentials is api credentials
type APICredentials struct {
	// Email is user email
	Email string `json:"email"`
	// PasswordSecretName is the k8s secret name that holds password
	PasswordSecretName string `json:"passwordSecretName"`
}

// VerbosityLevel is logging verbosity levels
// +kubebuilder:validation:Enum=debug;info;warn;error;panic
type VerbosityLevel string

const (
	// DebugLogs outputs only debugging logs
	DebugLogs VerbosityLevel = "debug"
	// InfoLogs outputs only informational logs
	InfoLogs VerbosityLevel = "info"
	// WarnLogs outputs only warning logs
	WarnLogs VerbosityLevel = "warn"
	// ErrorLogs outputs only error logs
	ErrorLogs VerbosityLevel = "error"
	// PanicLogs outputs only panic logs
	PanicLogs VerbosityLevel = "panic"
)

// NodeSpec defines the desired state of Node
type NodeSpec struct {
	// EthereumChainId is ethereum chain id
	EthereumChainId uint `json:"ethereumChainId"`
	// EthereumWSEndpoint is ethereum websocket endpoint
	EthereumWSEndpoint string `json:"ethereumWsEndpoint"`
	// EthereumHTTPEndpoints is ethereum http endpoints
	EthereumHTTPEndpoints []string `json:"ethereumHttpEndpoints,omitempty"`
	// LinkContractAddress is link contract address
	LinkContractAddress string `json:"linkContractAddress"`
	// DatabaseURL is postgres database connection URL
	DatabaseURL string `json:"databaseURL"`
	// KeystorePasswordSecretName is k8s secret name that holds keystore password
	KeystorePasswordSecretName string `json:"keystorePasswordSecretName"`
	// APICredentials is api credentials
	APICredentials APICredentials `json:"apiCredentials"`
	// CORSDomains is the domains from which to accept cross origin requests
	CORSDomains []string `json:"corsDomains,omitempty"`
	// CertSecretName is k8s secret name that holds tls.key and tls.cert
	CertSecretName string `json:"certSecretName,omitempty"`
	// TLSPort is port used for HTTPS connections
	TLSPort uint `json:"tlsPort,omitempty"`
	// P2PPort is port used for p2p communcations
	P2PPort uint `json:"p2pPort,omitempty"`
	// APIPort is port used for node API and GUI
	APIPort uint `json:"apiPort,omitempty"`
	// SecureCookies enables secure cookies for authentication
	SecureCookies bool `json:"secureCookies,omitempty"`
	// Logging is logging verboisty level
	Logging VerbosityLevel `json:"logging,omitempty"`
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// NodeStatus defines the observed state of Node
type NodeStatus struct{}

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
