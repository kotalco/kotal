package polkadot

import (
	"fmt"
	"strings"

	polkadotv1alpha1 "github.com/kotalco/kotal/apis/polkadot/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// PolkadotClient is parity Polkadot client
// https://github.com/paritytech/polkadot
type PolkadotClient struct {
	node *polkadotv1alpha1.Node
}

// Images
const (
	//  PolkadotHomeDir is go ipfs image home dir
	PolkadotHomeDir = "/polkadot"
)

// Command returns environment variables for the client
func (c *PolkadotClient) Env() []corev1.EnvVar {
	return nil
}

// Command is go-ipfs entrypoint
func (c *PolkadotClient) Command() []string {
	return nil
}

// Args returns go-ipfs args
func (c *PolkadotClient) Args() (args []string) {

	node := c.node

	args = append(args, PolkadotArgBasePath, shared.PathData(c.HomeDir()))
	args = append(args, PolkadotArgChain, node.Spec.Network)
	args = append(args, PolkadotArgName, node.Name)
	args = append(args, PolkadotArgPort, fmt.Sprintf("%d", node.Spec.P2PPort))
	args = append(args, PolkadotArgSync, string(node.Spec.SyncMode))
	args = append(args, PolkadotArgLogging, string(node.Spec.Logging))

	if node.Spec.Pruning != nil {
		var pruning bool = *node.Spec.Pruning
		if pruning {
			args = append(args, PolkadotArgPruning, fmt.Sprintf("%d", node.Spec.RetainedBlocks))
		} else {
			args = append(args, PolkadotArgPruning, "archive")
		}
	}

	if node.Spec.RPC {
		args = append(args, PolkadotArgRPCExternal)
		args = append(args, PolkadotArgRPCPort, fmt.Sprintf("%d", node.Spec.RPCPort))
	}

	if node.Spec.WS {
		args = append(args, PolkadotArgWSExternal)
		args = append(args, PolkadotArgWSPort, fmt.Sprintf("%d", node.Spec.WSPort))
	}

	if node.Spec.WS || node.Spec.RPC {
		args = append(args, PolkadotArgRPCCors, strings.Join(node.Spec.CORSDomains, ","))
	}

	if node.Spec.NodePrivateKeySecretName != "" {
		args = append(args, PolkadotArgNodeKeyType, "Ed25519")
		args = append(args, PolkadotArgNodeKeyFile, fmt.Sprintf("%s/kotal_nodekey", shared.PathData(c.HomeDir())))
	}

	if node.Spec.Telemetry {
		args = append(args, PolkadotArgTelemetryURL, node.Spec.TelemetryURL)
	} else {
		args = append(args, PolkadotArgNoTelemetry)
	}

	if node.Spec.Prometheus {
		args = append(args, PolkadotArgPrometheusExternal)
		args = append(args, PolkadotArgPrometheusPort, fmt.Sprintf("%d", node.Spec.PrometheusPort))
	} else {
		args = append(args, PolkadotArgNoPrometheus)
	}

	if node.Spec.Validator {
		args = append(args, PolkadotArgValidator)
	}

	return
}

func (c *PolkadotClient) HomeDir() string {
	return PolkadotHomeDir
}
