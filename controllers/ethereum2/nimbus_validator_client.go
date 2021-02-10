package controllers

import (
	"os"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// NimbusValidatorClient is Status Ethereum 2.0 client
type NimbusValidatorClient struct{}

// Images
const (
	// EnvNimbusValidatorImage is the environment variable used for Status Ethereum 2.0 validator client image
	EnvNimbusValidatorImage = "NIMBUS_VALIDATOR_CLIENT_IMAGE"
	// DefaultNimbusValidatorImage is the default Status Ethereum 2.0 validator client image
	DefaultNimbusValidatorImage = "kotalco/nimbus:v1.0.4"
)

// Args returns command line arguments required for client
func (t *NimbusValidatorClient) Args(validator *ethereum2v1alpha1.Validator) (args []string) {

	args = append(args, NimbusNonInteractive)

	return
}

// Command returns command for running the client
func (t *NimbusValidatorClient) Command() (command []string) {
	command = []string{"nimbus_validator_client"}
	return
}

// Image returns prysm docker image
func (t *NimbusValidatorClient) Image() string {
	if os.Getenv(EnvNimbusValidatorImage) == "" {
		return DefaultNimbusValidatorImage
	}
	return os.Getenv(EnvNimbusValidatorImage)
}
