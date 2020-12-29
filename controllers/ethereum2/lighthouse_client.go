package controllers

import (
	"os"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// LighthouseClient is SigmaPrime Ethereum 2.0 client
type LighthouseClient struct{}

// Images
const (
	// EnvLighthouseImage is the environment variable used for SigmaPrime Ethereum 2.0 client
	EnvLighthouseImage = "LIGHTHOUSE_IMAGE"
	// DefaultLighthouseImage is the default SigmaPrime Ethereum 2.0 client image
	DefaultLighthouseImage = "sigp/lighthouse:v1.0.5"
)

// Args returns command line arguments required for client
func (t *LighthouseClient) Args(node *ethereum2v1alpha1.Node) (args []string) {

	args = append(args, LighthouseNetwork, node.Spec.Join)

	if node.Spec.Eth1Endpoint != "" {
		args = append(args, LighthouseEth1)
		args = append(args, LighthouseEth1Endpoints, node.Spec.Eth1Endpoint)
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
