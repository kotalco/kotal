package ethereum2

import (
	"fmt"
	"net/url"
	"os"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// NimbusValidatorClient is Status Ethereum 2.0 client
type NimbusValidatorClient struct {
	validator *ethereum2v1alpha1.Validator
}

// Images
const (
	// EnvNimbusValidatorImage is the environment variable used for Status Ethereum 2.0 validator client image
	EnvNimbusValidatorImage = "NIMBUS_VALIDATOR_CLIENT_IMAGE"
	// DefaultNimbusValidatorImage is the default Status Ethereum 2.0 validator client image
	DefaultNimbusValidatorImage = "kotalco/nimbus:v1.5.1"
)

// HomeDir returns container home directory
func (t *NimbusValidatorClient) HomeDir() string {
	return NimbusHomeDir
}

// Args returns command line arguments required for client
func (t *NimbusValidatorClient) Args() (args []string) {

	validator := t.validator

	args = append(args, NimbusNonInteractive)

	args = append(args, argWithVal(NimbusDataDir, shared.PathData(t.HomeDir())))

	args = append(args, argWithVal(NimbusValidatorsDir, fmt.Sprintf("%s/kotal-validators/validator-keys", shared.PathData(t.HomeDir()))))

	args = append(args, argWithVal(NimbusSecretsDir, fmt.Sprintf("%s/kotal-validators/validator-secrets", shared.PathData(t.HomeDir()))))

	endpoint := validator.Spec.BeaconEndpoints[0]

	if endpoint != "" {
		// TODO: validate endpoint is valid from webhook
		u, _ := url.Parse(endpoint)
		port := u.Port()

		if port == "" {
			port = "80"
		} else {
			endpoint = u.Hostname()
		}

		// TODO: resolve host to ip

		args = append(args, argWithVal(NimbusRPCAddress, endpoint))
		args = append(args, argWithVal(NimbusRPCPort, port))
	}

	if validator.Spec.Graffiti != "" {
		args = append(args, argWithVal(NimbusGraffiti, validator.Spec.Graffiti))
	}

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
