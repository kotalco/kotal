package shared

// EthereumAddress is ethereum address
// +kubebuilder:validation:Pattern="^0[xX][0-9a-fA-F]{40}$"
type EthereumAddress string
