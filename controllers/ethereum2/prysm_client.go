package controllers

import (
	"fmt"
	"os"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// PrysmClient is Prysmatic Labs Ethereum 2.0 client
type PrysmClient struct{}

// Images
const (
	// EnvPrysmImage is the environment variable used for Prysmatic Labs Prysm client image
	EnvPrysmImage = "PRYSM_IMAGE"
	// DefaultPrysmImage is Prysmatic Labs Prysm client image
	// TODO: update with validator image
	DefaultPrysmImage = "gcr.io/prysmaticlabs/prysm/beacon-chain:v1.0.4"
)

// Args returns command line arguments required for client
func (t *PrysmClient) Args(node *ethereum2v1alpha1.Node) (args []string) {

	args = append(args, PrysmAcceptTermsOfUse)

	if node.Spec.Eth1Endpoint != "" {
		args = append(args, PrysmWeb3Provider, node.Spec.Eth1Endpoint)
	}

	args = append(args, fmt.Sprintf("--%s", node.Spec.Join))

	return
}

// Command returns command for running the client
func (t *PrysmClient) Command() (command []string) {
	return
}

// Image returns prysm docker image
func (t *PrysmClient) Image() string {
	if os.Getenv(EnvPrysmImage) == "" {
		return DefaultPrysmImage
	}
	return os.Getenv(EnvPrysmImage)
}
