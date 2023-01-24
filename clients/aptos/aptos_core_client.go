package aptos

import (
	aptosv1alpha1 "github.com/kotalco/kotal/apis/aptos/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// AptosCoreClient is Aptos core client
// https://github.com/aptos-labs/aptos-core
type AptosCoreClient struct {
	node *aptosv1alpha1.Node
}

// Images
const (
	// AptosCoreHomeDir is Aptos Core image home dir
	// TODO: create aptos image with non root user and /home/aptos home directory
	AptosCoreHomeDir = "/opt/aptos"
)

// Command returns environment variables for the client
func (c *AptosCoreClient) Env() (env []corev1.EnvVar) {
	return
}

// Command is Aptos Core client entrypoint
func (c *AptosCoreClient) Command() (command []string) {
	command = append(command, "/opt/aptos/bin/aptos-node")
	return
}

// Args returns Aptos Core client args
func (c *AptosCoreClient) Args() (args []string) {
	args = append(args, AptosArgConfig, "/opt/aptos/config/config.yaml")
	return
}

// HomeDir is the home directory of Aptos Core client image
func (c *AptosCoreClient) HomeDir() string {
	return AptosCoreHomeDir
}
