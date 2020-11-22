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

// NetworkSpec defines the desired state of Network
type NetworkSpec struct {
	// ID is network id
	ID uint `json:"id,omitempty"`

	// Join specifies the network to join
	Join string `json:"join,omitempty"`

	// Consensus is the consensus algorithm to be used by the network nodes to reach consensus
	Consensus ConsensusAlgorithm `json:"consensus,omitempty"`

	// Genesis is genesis block specification
	Genesis *Genesis `json:"genesis,omitempty"`

	// Nodes is array of node specifications
	// +kubebuilder:validation:MinItems=1
	Nodes []NodeSpec `json:"nodes"`

	// HighlyAvailable is whether blockchain nodes can land on the same k8s node or no
	HighlyAvailable bool `json:"highlyAvailable,omitempty"`

	// TopologyKey is the k8s node label used to distribute blockchain nodes
	TopologyKey string `json:"TopologyKey,omitempty"`
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

// PrivateKey is a private key
// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]{64}$"
type PrivateKey string

// ConsensusAlgorithm is the algorithm nodes use to reach consensus
// +kubebuilder:validation:Enum=poa;pow;ibft2;quorum
type ConsensusAlgorithm string

const (
	// ProofOfAuthority is proof of authority consensus algorithm
	ProofOfAuthority ConsensusAlgorithm = "poa"

	// ProofOfWork is proof of work (nakamoto consensus) consensus algorithm
	ProofOfWork ConsensusAlgorithm = "pow"

	// IstanbulBFT is Istanbul Byzantine Fault Tolerant consensus algorithm
	IstanbulBFT ConsensusAlgorithm = "ibft2"

	//Quorum is Quorum IBFT consensus algorithm
	Quorum ConsensusAlgorithm = "quorum"
)

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
