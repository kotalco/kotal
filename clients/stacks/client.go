package stacks

import (
	stacksv1alpha1 "github.com/kotalco/kotal/apis/stacks/v1alpha1"
	clients "github.com/kotalco/kotal/clients"
)

func NewClient(node *stacksv1alpha1.Node) clients.Interface {
	return &StacksNodeClient{node}
}
