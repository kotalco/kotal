package ethereum2

import (
	"fmt"
	"os"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// PrysmBeaconNode is Prysmatic Labs Ethereum 2.0 client
type PrysmBeaconNode struct {
	node *ethereum2v1alpha1.BeaconNode
}

// Images
const (
	// EnvPrysmBeaconNodeImage is the environment variable used for Prysmatic Labs beacon node image
	EnvPrysmBeaconNodeImage = "PRYSM_BEACON_NODE_IMAGE"
	// DefaultPrysmBeaconNodeImage is Prysmatic Labs beacon node image
	DefaultPrysmBeaconNodeImage = "gcr.io/prysmaticlabs/prysm/beacon-chain:v2.0.1"
)

// HomeDir returns container home directory
func (t *PrysmBeaconNode) HomeDir() string {
	return PrysmHomeDir
}

// Args returns command line arguments required for client
func (t *PrysmBeaconNode) Args() (args []string) {

	node := t.node

	args = append(args, PrysmAcceptTermsOfUse)

	args = append(args, PrysmDataDir, shared.PathData(t.HomeDir()))

	if len(node.Spec.Eth1Endpoints) != 0 {
		args = append(args, PrysmWeb3Provider, node.Spec.Eth1Endpoints[0])
		for _, provider := range node.Spec.Eth1Endpoints[1:] {
			args = append(args, PrysmFallbackWeb3Provider, provider)
		}
	}

	args = append(args, fmt.Sprintf("--%s", node.Spec.Network))

	if node.Spec.RPCPort != 0 {
		args = append(args, PrysmRPCPort, fmt.Sprintf("%d", node.Spec.RPCPort))
	}

	if node.Spec.RPCHost != "" {
		args = append(args, PrysmRPCHost, node.Spec.RPCHost)
	}

	if node.Spec.GRPC {
		if node.Spec.GRPCPort != 0 {
			args = append(args, PrysmGRPCPort, fmt.Sprintf("%d", node.Spec.GRPCPort))
		}
		if node.Spec.GRPCHost != "" {
			args = append(args, PrysmGRPCHost, node.Spec.GRPCHost)
		}
	} else {
		args = append(args, PrysmDisableGRPC)
	}

	if node.Spec.P2PPort != 0 {
		args = append(args, PrysmP2PTCPPort, fmt.Sprintf("%d", node.Spec.P2PPort))
		args = append(args, PrysmP2PUDPPort, fmt.Sprintf("%d", node.Spec.P2PPort))
	}

	return
}

// Command returns command for running the client
func (t *PrysmBeaconNode) Command() (command []string) {
	return
}

// Image returns prysm docker image
func (t *PrysmBeaconNode) Image() string {
	if os.Getenv(EnvPrysmBeaconNodeImage) == "" {
		return DefaultPrysmBeaconNodeImage
	}
	return os.Getenv(EnvPrysmBeaconNodeImage)
}
