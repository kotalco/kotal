package controllers

import (
	"fmt"
	"os"
	"strings"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// LighthouseClient is SigmaPrime Ethereum 2.0 client
type LighthouseClient struct{}

// Images
const (
	// EnvLighthouseImage is the environment variable used for SigmaPrime Ethereum 2.0 client
	EnvLighthouseImage = "LIGHTHOUSE_IMAGE"
	// DefaultLighthouseImage is the default SigmaPrime Ethereum 2.0 client image
	DefaultLighthouseImage = "sigp/lighthouse:v1.0.6"
)

// Args returns command line arguments required for client
func (t *LighthouseClient) Args(node *ethereum2v1alpha1.Node) (args []string) {

	args = append(args, LighthouseDataDir, PathBlockchainData)

	args = append(args, LighthouseNetwork, node.Spec.Join)

	if len(node.Spec.Eth1Endpoints) != 0 {
		args = append(args, LighthouseEth1)
		args = append(args, LighthouseEth1Endpoints, strings.Join(node.Spec.Eth1Endpoints, ","))
	}

	if node.Spec.REST {
		args = append(args, LighthouseHTTP)

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
func (t *LighthouseClient) Command() (command []string) {
	command = []string{"lighthouse", "beacon_node"}
	return
}

// Image returns prysm docker image
func (t *LighthouseClient) Image() string {
	if os.Getenv(EnvLighthouseImage) == "" {
		return DefaultLighthouseImage
	}
	return os.Getenv(EnvLighthouseImage)
}
