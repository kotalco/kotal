package aptos

import (
	"os"

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
	// EnvAptosCoreImage is the environment variable used for Aptos Core client image
	EnvAptosCoreImage = "APTOS_CORE_IMAGE"
	// DefaultAptosCoreDevnetImage is the default Aptos core Devnet client image
	DefaultAptosCoreDevnetImage = "aptoslab/validator:devnet"
	// DefaultAptosCoreTestnetImage is the default Aptos core Testnet client image
	DefaultAptosCoreTestnetImage = "aptoslab/validator:testnet"
	// AptosCoreHomeDir is Aptos Core image home dir
	// TODO: create aptos image with non root user and /home/aptos home directory
	AptosCoreHomeDir = "/opt/aptos"
)

// Image returns Aptos Core client image
func (c *AptosCoreClient) Image() string {
	if img := c.node.Spec.Image; img != nil {
		return *img
	} else if c.node.Spec.Network == aptosv1alpha1.Devnet {
		return DefaultAptosCoreDevnetImage
	} else if c.node.Spec.Network == aptosv1alpha1.Testnet {
		return DefaultAptosCoreTestnetImage
	}

	return os.Getenv(EnvAptosCoreImage)
}

// Command returns environment variables for the client
func (c *AptosCoreClient) Env() (env []corev1.EnvVar) {
	return
}

// Command is Aptos Core client entrypoint
func (c *AptosCoreClient) Command() (command []string) {
	return
}

// Args returns Aptos Core client args
func (c *AptosCoreClient) Args() (args []string) {
	return
}

// HomeDir is the home directory of Aptos Core client image
func (c *AptosCoreClient) HomeDir() string {
	return AptosCoreHomeDir
}
