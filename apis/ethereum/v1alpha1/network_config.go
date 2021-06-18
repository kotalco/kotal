package v1alpha1

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

// NetworkConfig is the shared network config between node and network
type NetworkConfig struct {
	// ID is network id
	ID uint `json:"id,omitempty"`

	// Network specifies the network to join
	Network string `json:"network,omitempty"`

	// Consensus is the consensus algorithm to be used by the network nodes to reach consensus
	Consensus ConsensusAlgorithm `json:"consensus,omitempty"`

	// Genesis is genesis block specification
	Genesis *Genesis `json:"genesis,omitempty"`
}
