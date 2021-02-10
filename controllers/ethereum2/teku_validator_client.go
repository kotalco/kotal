package controllers

import (
	"os"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
)

// TekuValidatorClient is Teku validator client
type TekuValidatorClient struct{}

const (
	// EnvTekuValidatorImage is the environment variable used for PegaSys Teku validator client image
	EnvTekuValidatorImage = "TEKU_VALIDATOR_CLIENT_IMAGE"
	// DefaultTekuValidatorImage is PegaSys Teku validator client image
	DefaultTekuValidatorImage = "consensys/teku:20.12.1"
)

// Args returns command line arguments required for client
func (t *TekuValidatorClient) Args(validator *ethereum2v1alpha1.Validator) (args []string) {

	args = append(args, "vc")

	args = append(args, TekuNetwork, validator.Spec.Network)

	return args
}

// Command returns command for running the client
func (t *TekuValidatorClient) Command() (command []string) {
	return
}

// Image returns teku docker image
func (t *TekuValidatorClient) Image() string {
	if os.Getenv(EnvTekuValidatorImage) == "" {
		return DefaultTekuValidatorImage
	}
	return os.Getenv(EnvTekuValidatorImage)
}
