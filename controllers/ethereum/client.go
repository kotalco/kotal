package controllers

import (
	"fmt"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
)

// EthereumClient is Ethereum client
type EthereumClient interface {
	Image() string
	Args() []string
	HomeDir() string
	Genesis() (string, error)
	LoggingArgFromVerbosity(ethereumv1alpha1.VerbosityLevel) string
	EncodeStaticNodes() string
}

// NewEthereumClient returns an Ethereum client instance
func NewEthereumClient(node *ethereumv1alpha1.Node) (EthereumClient, error) {
	switch node.Spec.Client {
	case ethereumv1alpha1.BesuClient:
		return &BesuClient{node}, nil
	case ethereumv1alpha1.GethClient:
		return &GethClient{node}, nil
	case ethereumv1alpha1.ParityClient:
		return &ParityClient{node}, nil
	default:
		return nil, fmt.Errorf("client %s is not supported", node.Spec.Client)
	}
}
