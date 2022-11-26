package ethereum2

import (
	"fmt"
	"os"
	"strings"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// LighthouseBeaconNode is SigmaPrime Ethereum 2.0 client
// https://github.com/sigp/lighthouse
type LighthouseBeaconNode struct {
	node *ethereum2v1alpha1.BeaconNode
}

// Images
const (
	// EnvLighthouseBeaconNodeImage is the environment variable used for SigmaPrime Ethereum 2.0 beacon node image
	EnvLighthouseBeaconNodeImage = "LIGHTHOUSE_BEACON_NODE_IMAGE"
	// DefaultLighthouseBeaconNodeImage is the default SigmaPrime Ethereum 2.0 beacon node image
	DefaultLighthouseBeaconNodeImage = "kotalco/lighthouse:v3.3.0"
)

// HomeDir returns container home directory
func (t *LighthouseBeaconNode) HomeDir() string {
	return LighthouseHomeDir
}

// Command returns environment variables for running the client
func (t *LighthouseBeaconNode) Env() []corev1.EnvVar {
	return nil
}

// Args returns command line arguments required for client
func (t *LighthouseBeaconNode) Args() (args []string) {

	node := t.node

	args = append(args, LighthouseDataDir, shared.PathData(t.HomeDir()))

	args = append(args, LighthouseDebugLevel, string(t.node.Spec.Logging))

	args = append(args, LighthouseNetwork, node.Spec.Network)

	if len(node.Spec.Eth1Endpoints) != 0 {
		args = append(args, LighthouseEth1)
		args = append(args, LighthouseEth1Endpoints, strings.Join(node.Spec.Eth1Endpoints, ","))
	}

	if node.Spec.REST {
		args = append(args, LighthouseHTTP)
		args = append(args, LighthouseAllowOrigins, strings.Join(node.Spec.CORSDomains, ","))

		if node.Spec.RESTPort != 0 {
			args = append(args, LighthouseHTTPPort, fmt.Sprintf("%d", node.Spec.RESTPort))
		}
		if node.Spec.RESTHost != "" {
			args = append(args, LighthouseHTTPAddress, node.Spec.RESTHost)
		}
	}

	if node.Spec.P2PPort != 0 {
		args = append(args, LighthousePort, fmt.Sprintf("%d", node.Spec.P2PPort))
		args = append(args, LighthouseDiscoveryPort, fmt.Sprintf("%d", node.Spec.P2PPort))
	}

	return
}

// Command returns command for running the client
func (t *LighthouseBeaconNode) Command() (command []string) {
	command = []string{"lighthouse", "bn"}
	return
}

// Image returns prysm docker image
func (t *LighthouseBeaconNode) Image() string {
	if img := t.node.Spec.Image; img != nil {
		return *img
	} else if os.Getenv(EnvLighthouseBeaconNodeImage) == "" {
		return DefaultLighthouseBeaconNodeImage
	}
	return os.Getenv(EnvLighthouseBeaconNodeImage)
}
