package v1alpha1

// Ethereum2Client is Ethereum 2.0 client
// +kubebuilder:validation:Enum=teku;prysm;lighthouse;nimbus
type Ethereum2Client string

const (
	// TekuClient is ConsenSys Pegasys Ethereum 2.0 client
	TekuClient Ethereum2Client = "teku"
	// PrysmClient is Prysmatic Labs Ethereum 2.0 client
	PrysmClient Ethereum2Client = "prysm"
	// LighthouseClient is SigmaPrime Ethereum 2.0 client
	LighthouseClient Ethereum2Client = "lighthouse"
	// NimbusClient is Status Ethereum 2.0 client
	NimbusClient Ethereum2Client = "nimbus"
)
