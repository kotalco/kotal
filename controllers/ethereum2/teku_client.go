package controllers

import (
	"os"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// TekuClient is ConsenSys Pegasys Ethereum 2.0 client
type TekuClient struct{}

const (
	// EnvTekuImage is the environment variable used for PegaSys Teku client image
	EnvTekuImage = "TEKU_IMAGE"
	// DefaultTekuImage is PegaSys Teku client image
	DefaultTekuImage = "consensys/teku:20.12"
)

// Args returns command line arguments required for client
func (t *TekuClient) Args(node *ethereum2v1alpha1.Node) (args []string) {

	args = append(args, TekuDataPath, PathBlockchainData)

	args = append(args, TekuNetwork, node.Spec.Join)

	if node.Spec.Eth1Endpoint != "" {
		args = append(args, TekuEth1Endpoint, node.Spec.Eth1Endpoint)
	}

	if node.Spec.REST {
		args = append(args, TekuRestEnabled)
	}

	return
}

// Command returns command for running the client
func (t *TekuClient) Command() (command []string) {
	return
}

// Image returns teku docker image
func (t *TekuClient) Image() string {
	if os.Getenv(EnvTekuImage) == "" {
		return DefaultTekuImage
	}
	return os.Getenv(EnvTekuImage)
}
