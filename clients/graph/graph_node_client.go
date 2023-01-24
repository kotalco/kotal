package graph

import (
	graphv1alpha1 "github.com/kotalco/kotal/apis/graph/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// GraphNodeClient is graph node client
// https://github.com/graphprotocol/graph-node
type GraphNodeClient struct {
	node *graphv1alpha1.Node
}

// Images
const (
	// GraphNodeHomeDir is Graph node image home dir
	// TODO: update home dir after creating a new docker image
	GraphNodeHomeDir = "/root"
)

// Command returns environment variables for the client
func (c *GraphNodeClient) Env() (env []corev1.EnvVar) {
	return
}

// Command is Graph node client entrypoint
func (c *GraphNodeClient) Command() (command []string) {

	command = append(command, GraphNodeCommand)

	return
}

// Args returns Graph node client args
func (c *GraphNodeClient) Args() (args []string) {
	_ = c.node

	return
}

// HomeDir is the home directory of Graph node client image
func (c *GraphNodeClient) HomeDir() string {
	return GraphNodeHomeDir
}
