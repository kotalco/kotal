package controllers

import ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"

// TekuClient is ConsenSys Pegasys Ethereum 2.0 client
type TekuClient struct{}

// GetArgs returns command line arguments required for client
func (t *TekuClient) GetArgs(node *ethereum2v1alpha1.Node) (args []string) {

	args = append(args, "--network", node.Spec.Join)

	return
}
