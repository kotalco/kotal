package filecoin

import (
	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
	clients "github.com/kotalco/kotal/clients"
)

func NewClient(node *filecoinv1alpha1.Node) clients.Interface {
	return &LotusClient{node}
}
