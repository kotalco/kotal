package v1alpha1

import "github.com/kotalco/kotal/apis/shared"

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

func (client Ethereum2Client) SupportsVerbosityLevel(level shared.VerbosityLevel, validator bool) bool {
	switch client {

	case TekuClient:
		switch level {
		case shared.NoLogs,
			shared.FatalLogs,
			shared.ErrorLogs,
			shared.WarnLogs,
			shared.InfoLogs,
			shared.DebugLogs,
			shared.TraceLogs,
			shared.AllLogs:
			return true
		}

	case PrysmClient:
		switch level {
		case shared.TraceLogs,
			shared.DebugLogs,
			shared.InfoLogs,
			shared.WarnLogs,
			shared.ErrorLogs,
			shared.FatalLogs,
			shared.PanicLogs:
			return true
		}

	case LighthouseClient:
		switch level {
		case shared.InfoLogs,
			shared.DebugLogs,
			shared.TraceLogs,
			shared.WarnLogs,
			shared.ErrorLogs,
			shared.CriticalLogs:
			return true
		}

	case NimbusClient:
		switch level {
		case shared.TraceLogs,
			shared.DebugLogs,
			shared.InfoLogs,
			shared.NoticeLogs,
			shared.WarnLogs,
			shared.ErrorLogs,
			shared.FatalLogs,
			shared.NoneLogs:
			return true
		}
	}

	return false
}
