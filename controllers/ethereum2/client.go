package controllers

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// BeaconNodeClient is Ethereum 2.0 beacon node client
type BeaconNodeClient interface {
	shared.Client
}

// ValidatorClient is Ethereum 2.0 validator client
type ValidatorClient interface {
	shared.Client
}

// NewBeaconNodeClient returns an Ethereum beacon node client instance
func NewBeaconNodeClient(node *ethereum2v1alpha1.BeaconNode) (BeaconNodeClient, error) {
	switch node.Spec.Client {
	case ethereum2v1alpha1.TekuClient:
		return &TekuBeaconNode{node}, nil
	case ethereum2v1alpha1.PrysmClient:
		return &PrysmBeaconNode{node}, nil
	case ethereum2v1alpha1.LighthouseClient:
		return &LighthouseBeaconNode{node}, nil
	case ethereum2v1alpha1.NimbusClient:
		return &NimbusBeaconNode{node}, nil
	default:
		return nil, fmt.Errorf("client %s is not supported", node.Spec.Client)
	}
}

// NewValidatorClient returns an Ethereum validator client instance
func NewValidatorClient(validator *ethereum2v1alpha1.Validator) (ValidatorClient, error) {
	switch validator.Spec.Client {
	case ethereum2v1alpha1.TekuClient:
		return &TekuValidatorClient{validator}, nil
	case ethereum2v1alpha1.PrysmClient:
		return &PrysmValidatorClient{validator}, nil
	case ethereum2v1alpha1.LighthouseClient:
		return &LighthouseValidatorClient{validator}, nil
	case ethereum2v1alpha1.NimbusClient:
		return &NimbusValidatorClient{validator}, nil
	default:
		return nil, fmt.Errorf("client %s is not supported", validator.Spec.Client)
	}
}
