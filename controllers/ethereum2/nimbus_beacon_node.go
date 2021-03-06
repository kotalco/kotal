package controllers

import (
	"fmt"
	"os"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// NimbusBeaconNode is Status Ethereum 2.0 client
type NimbusBeaconNode struct{}

// Images
const (
	// EnvNimbusBeaconNodeImage is the environment variable used for Status Ethereum 2.0 beacon node image
	EnvNimbusBeaconNodeImage = "NIMBUS_BEACON_NODE_IMAGE"
	// DefaultNimbusBeaconNodeImage is the default Status Ethereum 2.0 beacon node image
	DefaultNimbusBeaconNodeImage = "kotalco/nimbus:v1.0.8"
)

// HomeDir returns container home directory
func (t *NimbusBeaconNode) HomeDir() string {
	return NimbusHomeDir
}

// Args returns command line arguments required for client
func (t *NimbusBeaconNode) Args(node *ethereum2v1alpha1.BeaconNode) (args []string) {

	args = append(args, NimbusNonInteractive)

	args = append(args, argWithVal(NimbusDataDir, PathBlockchainData(t.HomeDir())))

	args = append(args, argWithVal(NimbusNetwork, node.Spec.Join))

	if len(node.Spec.Eth1Endpoints) != 0 {
		args = append(args, argWithVal(NimbusEth1Endpoint, node.Spec.Eth1Endpoints[0]))
	}

	if node.Spec.RPC {
		args = append(args, NimbusRPC)
		if node.Spec.RPCPort != 0 {
			args = append(args, argWithVal(NimbusRPCPort, fmt.Sprintf("%d", node.Spec.RPCPort)))
		}
		if node.Spec.RPCHost != "" {
			args = append(args, argWithVal(NimbusRPCAddress, node.Spec.RPCHost))
		}
	}

	if node.Spec.P2PPort != 0 {
		args = append(args, argWithVal(NimbusTCPPort, fmt.Sprintf("%d", node.Spec.P2PPort)))
		args = append(args, argWithVal(NimbusUDPPort, fmt.Sprintf("%d", node.Spec.P2PPort)))
	}

	return
}

// Command returns command for running the client
func (t *NimbusBeaconNode) Command() (command []string) {
	command = []string{"nimbus_beacon_node"}
	return
}

// Image returns prysm docker image
func (t *NimbusBeaconNode) Image() string {
	if os.Getenv(EnvNimbusBeaconNodeImage) == "" {
		return DefaultNimbusBeaconNodeImage
	}
	return os.Getenv(EnvNimbusBeaconNodeImage)
}

// nimbus accepts arguments in the form of --arg=val
// --arg val is not recoginized by nimbus
func argWithVal(arg, val string) string {
	return fmt.Sprintf("%s=%s", arg, val)
}
