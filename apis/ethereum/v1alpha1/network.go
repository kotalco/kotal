package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	// MainNetwork is ethereum main network
	MainNetwork = "mainnet"
	// RopstenNetwork is ropsten pow network
	RopstenNetwork = "ropsten"
	// RinkebyNetwork is rinkeby poa network
	RinkebyNetwork = "rinkeby"
	// GoerliNetwork is goerli poa cross-client network
	GoerliNetwork = "goerli"
	// KottiNetwork is kotti poa ethereum classic test network
	KottiNetwork = "kotti"
	// ClassicNetwork is ethereum classic network
	ClassicNetwork = "classic"
	// MordorNetwork is mordon poe ethereum classic test network
	MordorNetwork = "mordor"
	// DevNetwork is local development network
	DevNetwork = "dev"
)

// AvailabilityConfig is the shared high availability config between node and network
type AvailabilityConfig struct {
	// HighlyAvailable is whether blockchain nodes can land on the same k8s node or no
	HighlyAvailable bool `json:"highlyAvailable,omitempty"`

	// TopologyKey is the k8s node label used to distribute blockchain nodes
	TopologyKey string `json:"TopologyKey,omitempty"`
}

// NetworkSpec defines the desired state of Network
type NetworkSpec struct {
	NetworkConfig      `json:",inline"`
	AvailabilityConfig `json:",inline"`

	// Nodes is array of network node specifications
	// +kubebuilder:validation:MinItems=1
	Nodes []NetworkNodeSpec `json:"nodes"`
}

// NetworkNodeSpec is a network node spec
type NetworkNodeSpec struct {
	NodeSpec `json:",inline"`
	Name     string `json:"name"`
}

// HexString is String in hexadecial format
// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]+$"
type HexString string

// EthereumAddress is ethereum address
// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]{40}$"
type EthereumAddress string

// Hash is KECCAK-256 hash
// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]{64}$"
type Hash string

// NetworkStatus defines the observed state of Network
type NetworkStatus struct {

	// NodesCount is number of nodes in this network
	NodesCount int `json:"nodesCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Network is the Schema for the networks API
// +kubebuilder:printcolumn:name="Consensus",type=string,JSONPath=".spec.consensus"
// +kubebuilder:printcolumn:name="Join",type=string,JSONPath=".spec.join"
// +kubebuilder:printcolumn:name="Nodes",type=integer,JSONPath=".status.nodesCount"
type Network struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkSpec   `json:"spec,omitempty"`
	Status NetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkList contains a list of Network
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Network `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Network{}, &NetworkList{})
}
