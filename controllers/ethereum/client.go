package controllers

import (
	"fmt"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
)

// EthereumClient is Ethereum client
type EthereumClient interface {
	Args(*ethereumv1alpha1.Node) []string
	Genesis(*ethereumv1alpha1.Node) (string, error)
	LoggingArgFromVerbosity(ethereumv1alpha1.VerbosityLevel) string
	EncodeStaticNodes(*ethereumv1alpha1.Node) string
}

// NewEthereumClient returns an Ethereum client instance
func NewEthereumClient(name ethereumv1alpha1.EthereumClient) (EthereumClient, error) {
	switch name {
	case ethereumv1alpha1.BesuClient:
		return &BesuClient{}, nil
	case ethereumv1alpha1.GethClient:
		return &GethClient{}, nil
	case ethereumv1alpha1.ParityClient:
		return &ParityClient{}, nil
	default:
		return nil, fmt.Errorf("client %s is not supported", name)
	}
}
