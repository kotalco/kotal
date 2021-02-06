package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidatorSpec defines the desired state of Validator
type ValidatorSpec struct {
	// Network is the network this validator is validating blocks for
	Network string `json:"network"`
}

// ValidatorStatus defines the observed state of Validator
type ValidatorStatus struct{}

// +kubebuilder:object:root=true

// Validator is the Schema for the validators API
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
