package v1alpha1

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

// HexString is String in hexadecial format
// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]+$"
type HexString string

// EthereumAddress is ethereum address
// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]{40}$"
type EthereumAddress string

// Hash is KECCAK-256 hash
// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]{64}$"
type Hash string
