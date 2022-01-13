package near

import (
	"fmt"
	"os"

	nearv1alpha1 "github.com/kotalco/kotal/apis/near/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// NearClient is NEAR core client
// https://github.com/near/nearcore/
type NearClient struct {
	node *nearv1alpha1.Node
}

// Images
const (
	// EnvNearImage is the environment variable used for NEAR core client image
	EnvNearImage = "NEAR_IMAGE"
	// DefaultNearImage is the default NEAR core client image
	DefaultNearImage = "nearprotocol/nearcore:1.23.1"
	// NearHomeDir is go ipfs image home dir
	// TODO: update home dir after building docker image with non-root user and home dir
	NearHomeDir = "/root/.near"
)

// Image returns NEAR core client image
func (c *NearClient) Image() string {
	if os.Getenv(EnvNearImage) == "" {
		return DefaultNearImage
	}
	return os.Getenv(EnvNearImage)
}

// Command returns environment variables for the client
func (c *NearClient) Env() []corev1.EnvVar {
	return nil
}

// Command is NEAR core client entrypoint
func (c *NearClient) Command() []string {
	return nil
}

// Args returns NEAR core client args
func (c *NearClient) Args() (args []string) {

	node := c.node

	args = append(args, "neard")
	args = append(args, NearArgHome, c.HomeDir())
	args = append(args, "run")

	if node.Spec.RPC {
		args = append(args, NearArgRPCAddress, fmt.Sprintf("%s:%d", node.Spec.RPCHost, node.Spec.RPCPort))
	} else {
		args = append(args, NearArgDisableRPC)
	}

	return
}

// HomeDir is the home directory of NEAR core client image
func (c *NearClient) HomeDir() string {
	return NearHomeDir
}
