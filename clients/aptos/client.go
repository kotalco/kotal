package aptos

import (
	aptosv1alpha1 "github.com/kotalco/kotal/apis/aptos/v1alpha1"
	clients "github.com/kotalco/kotal/clients"
)

// NewClient returns new Aptos client
func NewClient(node *aptosv1alpha1.Node) clients.Interface {
	return &AptosCoreClient{node}
}
