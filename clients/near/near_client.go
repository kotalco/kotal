package near

import (
	"fmt"
	"strings"

	nearv1alpha1 "github.com/kotalco/kotal/apis/near/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// NearClient is NEAR core client
// https://github.com/near/nearcore/
type NearClient struct {
	node *nearv1alpha1.Node
}

// Images
const (
	// NearHomeDir is go ipfs image home dir
	// TODO: update home dir after building docker image with non-root user and home dir
	NearHomeDir = "/home/near"
)

// Command returns environment variables for the client
func (c *NearClient) Env() []corev1.EnvVar {
	return nil
}

// Command is NEAR core client entrypoint
func (c *NearClient) Command() []string {
	return nil
}

// Args returns NEAR core client args
func (c *NearClient) Args() (args []string) {

	node := c.node

	args = append(args, "neard")
	args = append(args, NearArgHome, shared.PathData(c.HomeDir()))
	args = append(args, "run")

	args = append(args, NearArgNetworkAddress, fmt.Sprintf("%s:%d", shared.Host(true), node.Spec.P2PPort))

	if node.Spec.RPC {
		args = append(args, NearArgRPCAddress, fmt.Sprintf("%s:%d", shared.Host(node.Spec.RPC), node.Spec.RPCPort))
		args = append(args, NearArgPrometheusAddress, fmt.Sprintf("%s:%d", shared.Host(true), node.Spec.PrometheusPort))
	} else {
		args = append(args, NearArgDisableRPC)
	}

	if node.Spec.TelemetryURL != "" {
		args = append(args, NearArgTelemetryURL, node.Spec.TelemetryURL)
	}

	if node.Spec.Archive {
		args = append(args, NearArgArchive)
	}

	args = append(args, NearArgMinimumPeers, fmt.Sprintf("%d", node.Spec.MinPeers))

	if len(node.Spec.Bootnodes) != 0 {
		args = append(args, NearArgBootnodes, strings.Join(node.Spec.Bootnodes, ","))
	}

	return
}

// HomeDir is the home directory of NEAR core client image
func (c *NearClient) HomeDir() string {
	return NearHomeDir
}
