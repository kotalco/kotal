package chainlink

import (
	chainlinkv1alpha1 "github.com/kotalco/kotal/apis/chainlink/v1alpha1"
	clients "github.com/kotalco/kotal/clients"
)

// NewClient returns chainlink client for the given node
func NewClient(node *chainlinkv1alpha1.Node) clients.Interface {
	return &ChainlinkClient{node}
}
