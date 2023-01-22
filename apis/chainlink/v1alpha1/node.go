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

// NodeSpec defines the desired state of Node
type NodeSpec struct {
	// Image is Chainlink node client image
	Image string `json:"image,omitempty"`
	// EthereumChainId is ethereum chain id
	EthereumChainId uint `json:"ethereumChainId"`
	// EthereumWSEndpoint is ethereum websocket endpoint
	EthereumWSEndpoint string `json:"ethereumWsEndpoint"`
	// EthereumHTTPEndpoints is ethereum http endpoints
	// +listType=set
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
	// +listType=set
	CORSDomains []string `json:"corsDomains,omitempty"`
	// CertSecretName is k8s secret name that holds tls.key and tls.cert
	CertSecretName string `json:"certSecretName,omitempty"`
	// TLSPort is port used for HTTPS connections
	TLSPort uint `json:"tlsPort,omitempty"`
	// P2PPort is port used for p2p communcations
	P2PPort uint `json:"p2pPort,omitempty"`
	// API enables node API server
	API bool `json:"api,omitempty"`
	// APIPort is port used for node API and GUI
	APIPort uint `json:"apiPort,omitempty"`
	// SecureCookies enables secure cookies for authentication
	SecureCookies bool `json:"secureCookies,omitempty"`
	// Logging is logging verboisty level
	// +kubebuilder:validation:Enum=debug;info;warn;error;panic
	Logging shared.VerbosityLevel `json:"logging,omitempty"`
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
// +kubebuilder:printcolumn:name="Client",type=string,JSONPath=".status.client"
// +kubebuilder:printcolumn:name="EthereumChainId",type=number,JSONPath=".spec.ethereumChainId"
// +kubebuilder:printcolumn:name="LinkContractAddress",type=string,JSONPath=".spec.linkContractAddress",priority=10
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
