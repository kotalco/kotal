package stacks

import (
	"fmt"
	"os"

	stacksv1alpha1 "github.com/kotalco/kotal/apis/stacks/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// StacksNodeClient is Stacks blockchain node client
// https://github.com/stacks-network/stacks-blockchain
type StacksNodeClient struct {
	node *stacksv1alpha1.Node
}

// Images
const (
	// EnvStacksNodeImage is the environment variable used for Stacks node client image
	EnvStacksNodeImage = "STACKS_NODE_IMAGE"
	// DefaultStacksNodeImage is the default Stacks node client image
	DefaultStacksNodeImage = "kotalco/stacks:2.05.0.1.0"
	// StacksNodeHomeDir is Stacks node image home dir
	// TODO: update home dir after creating a new docker image
	StacksNodeHomeDir = "/home/stacks"
)

// Image returns Stacks node client image
func (c *StacksNodeClient) Image() string {
	if img := c.node.Spec.Image; img != nil {
		return *img
	} else if os.Getenv(EnvStacksNodeImage) == "" {
		return DefaultStacksNodeImage
	}
	return os.Getenv(EnvStacksNodeImage)
}

// Command returns environment variables for the client
func (c *StacksNodeClient) Env() (env []corev1.EnvVar) {
	return
}

// Command is Stacks node client entrypoint
func (c *StacksNodeClient) Command() (command []string) {

	command = append(command, StacksNodeCommand, StacksStartCommand)

	return
}

// Args returns Stacks node client args
func (c *StacksNodeClient) Args() (args []string) {
	_ = c.node

	args = append(args, StacksArgConfig, fmt.Sprintf("%s/config.toml", shared.PathConfig(c.HomeDir())))

	return
}

// HomeDir is the home directory of Stacks node client image
func (c *StacksNodeClient) HomeDir() string {
	return StacksNodeHomeDir
}
