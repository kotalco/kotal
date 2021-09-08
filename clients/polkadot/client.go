package polkadot

import (
	polkadotv1alpha1 "github.com/kotalco/kotal/apis/polkadot/v1alpha1"
	clients "github.com/kotalco/kotal/clients"
)

func NewClient(node *polkadotv1alpha1.Node) clients.Interface {
	return &PolkadotClient{node}
}
