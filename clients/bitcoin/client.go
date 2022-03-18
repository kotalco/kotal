package bitcoin

import (
	bitcoinv1alpha1 "github.com/kotalco/kotal/apis/bitcoin/v1alpha1"
	clients "github.com/kotalco/kotal/clients"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewClient(node *bitcoinv1alpha1.Node, client client.Client) clients.Interface {
	return &BitcoinCoreClient{node, client}
}
