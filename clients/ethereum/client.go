package ethereum

import (
	"fmt"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	"github.com/kotalco/kotal/clients"
)

// EthereumClient is Ethereum client
type EthereumClient interface {
	clients.Interface
	Genesis() (string, error)
	EncodeStaticNodes() string
}

// NewClient returns an Ethereum client instance
func NewClient(node *ethereumv1alpha1.Node) (EthereumClient, error) {
	switch node.Spec.Client {
	case ethereumv1alpha1.BesuClient:
		return &BesuClient{node}, nil
	case ethereumv1alpha1.GethClient:
		return &GethClient{node}, nil
	case ethereumv1alpha1.NethermindClient:
		return &NethermindClient{&ParityGenesis{}, node}, nil
	default:
		return nil, fmt.Errorf("client %s is not supported", node.Spec.Client)
	}
}
