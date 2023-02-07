package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidatorSpec defines the desired state of Validator
type ValidatorSpec struct {
	// Image is Ethereum 2.0 validator client image
	Image string `json:"image,omitempty"`

	// Network is the network this validator is validating blocks for
	Network string `json:"network"`
	// Client is the Ethereum 2.0 client to use
	Client Ethereum2Client `json:"client"`
	// FeeRecipient is ethereum address collecting transaction fees
	FeeRecipient shared.EthereumAddress `json:"feeRecipient,omitempty"`
	// BeaconEndpoints is beacon node endpoints
	// +kubebuilder:validation:MinItems=1
	// +listType=set
	BeaconEndpoints []string `json:"beaconEndpoints"`
	// Graffiti is the text to include in proposed blocks
	Graffiti string `json:"graffiti,omitempty"`
	// Logging is logging verboisty level
	// +kubebuilder:validation:Enum=off;fatal;error;warn;info;debug;trace;all;notice;crit;panic;none
	Logging shared.VerbosityLevel `json:"logging,omitempty"`
	// CertSecretName is k8s secret name that holds tls.crt
	CertSecretName string `json:"certSecretName,omitempty"`
	// Keystores is a list of Validator keystores
	// +kubebuilder:validation:MinItems=1
	Keystores []Keystore `json:"keystores"`
	// WalletPasswordSecret is wallet password secret
	WalletPasswordSecret string `json:"walletPasswordSecret,omitempty"`
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// Keystore is Ethereum 2.0 validator EIP-2335 BLS12-381 keystore https://eips.ethereum.org/EIPS/eip-2335
type Keystore struct {
	// PublicKey is the validator public key in hexadecimal
	// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]{96}$"
	PublicKey string `json:"publicKey,omitempty"`
	// SecretName is the kubernetes secret holding [keystore] and [password]
	SecretName string `json:"secretName"`
}

// ValidatorStatus defines the observed state of Validator
type ValidatorStatus struct{}

// +kubebuilder:object:root=true

// Validator is the Schema for the validators API
// +kubebuilder:printcolumn:name="Client",type=string,JSONPath=".spec.client"
// +kubebuilder:printcolumn:name="Network",type=string,JSONPath=".spec.network"
type Validator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ValidatorSpec   `json:"spec,omitempty"`
	Status ValidatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ValidatorList contains a list of Validator
type ValidatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Validator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Validator{}, &ValidatorList{})
}
