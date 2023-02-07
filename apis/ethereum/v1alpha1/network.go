package v1alpha1

const (
	// MainNetwork is ethereum main network
	MainNetwork = "mainnet"
	// RopstenNetwork is ropsten pos network
	RopstenNetwork = "ropsten"
	// RinkebyNetwork is rinkeby poa network
	RinkebyNetwork = "rinkeby"
	// GoerliNetwork is goerli pos cross-client network
	GoerliNetwork = "goerli"
	// SepoliaNetwork is sepolia pos network
	SepoliaNetwork = "sepolia"
	// XDaiNetwork is xdai pos network
	XDaiNetwork = "xdai"
	// KottiNetwork is kotti poa ethereum classic test network
	KottiNetwork = "kotti"
	// ClassicNetwork is ethereum classic network
	ClassicNetwork = "classic"
	// MordorNetwork is mordon poe ethereum classic test network
	MordorNetwork = "mordor"
	// DevNetwork is local development network
	DevNetwork = "dev"
)

// HexString is String in hexadecial format
// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]+$"
type HexString string

// Hash is KECCAK-256 hash
// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]{64}$"
type Hash string
