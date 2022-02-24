package bitcoin

import (
	bitcoinv1alpha1 "github.com/kotalco/kotal/apis/bitcoin/v1alpha1"
	clients "github.com/kotalco/kotal/clients"
)

func NewClient(node *bitcoinv1alpha1.Node) clients.Interface {
	return &BitcoinCoreClient{node}
}
