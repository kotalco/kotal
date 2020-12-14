package controllers

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// Ethereum2Client is Ethereum client
type Ethereum2Client interface {
	GetArgs(*ethereum2v1alpha1.Node) []string
	Image() string
}

// NewEthereum2Client returns an Ethereum client instance
func NewEthereum2Client(name ethereum2v1alpha1.Ethereum2Client) (Ethereum2Client, error) {
	switch name {
	case ethereum2v1alpha1.TekuClient:
		return &TekuClient{}, nil
	case ethereum2v1alpha1.PrysmClient:
		return &PrysmClient{}, nil
	default:
		return nil, fmt.Errorf("Client %s is not supported", name)
	}
}