package controllers

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// Ethereum2Client is Ethereum client
type Ethereum2Client interface {
	Command() []string
	Image() string
}

// BeaconNodeClient is Ethereum 2.0 beacon node client
type BeaconNodeClient interface {
	Ethereum2Client
	Args(*ethereum2v1alpha1.BeaconNode) []string
}

// ValidatorClient is Ethereum 2.0 validator client
type ValidatorClient interface {
	Ethereum2Client
	Args(*ethereum2v1alpha1.Validator) []string
}

// NewBeaconNodeClient returns an Ethereum beacon node client instance
func NewBeaconNodeClient(name ethereum2v1alpha1.Ethereum2Client) (BeaconNodeClient, error) {
	switch name {
	case ethereum2v1alpha1.TekuClient:
		return &TekuBeaconNode{}, nil
	case ethereum2v1alpha1.PrysmClient:
		return &PrysmBeaconNode{}, nil
	case ethereum2v1alpha1.LighthouseClient:
		return &LighthouseBeaconNode{}, nil
	case ethereum2v1alpha1.NimbusClient:
		return &NimbusBeaconNode{}, nil
	default:
		return nil, fmt.Errorf("Client %s is not supported", name)
	}
}
