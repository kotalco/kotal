package graph

import (
	graphv1alpha1 "github.com/kotalco/kotal/apis/graph/v1alpha1"
	clients "github.com/kotalco/kotal/clients"
)

// NewClient creates new graph node client
func NewClient(node *graphv1alpha1.Node) clients.Interface {
	return &GraphNodeClient{node}
}
