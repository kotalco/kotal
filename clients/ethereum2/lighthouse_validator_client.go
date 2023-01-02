package ethereum2

import (
	"os"
	"strings"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// LighthouseValidatorClient is SigmaPrime Ethereum 2.0 validator client
// https://github.com/sigp/lighthouse
type LighthouseValidatorClient struct {
	validator *ethereum2v1alpha1.Validator
}

// Images
const (
	// EnvLighthouseValidatorImage is the environment variable used for SigmaPrime Ethereum 2.0 validator client image
	EnvLighthouseValidatorImage = "LIGHTHOUSE_VALIDATOR_CLIENT_IMAGE"
	// DefaultLighthouseValidatorImage is the default SigmaPrime Ethereum 2.0 validator client image
	DefaultLighthouseValidatorImage = "kotalco/lighthouse:v3.3.0"
)

// HomeDir returns container home directory
func (t *LighthouseValidatorClient) HomeDir() string {
	return LighthouseHomeDir
}

// Command returns environment variables for the client
func (t *LighthouseValidatorClient) Env() []corev1.EnvVar {
	return nil
}

// Args returns command line arguments required for client
func (t *LighthouseValidatorClient) Args() (args []string) {

	validator := t.validator

	args = append(args, LighthouseDataDir, shared.PathData(t.HomeDir()))

	args = append(args, LighthouseDebugLevel, string(t.validator.Spec.Logging))

	args = append(args, LighthouseNetwork, validator.Spec.Network)

	args = append(args, LighthouseFeeRecipient, string(validator.Spec.FeeRecipient))

	if len(validator.Spec.BeaconEndpoints) != 0 {
		args = append(args, LighthouseBeaconNodeEndpoints, strings.Join(validator.Spec.BeaconEndpoints, ","))
	}

	if validator.Spec.Graffiti != "" {
		args = append(args, LighthouseGraffiti, validator.Spec.Graffiti)
	}

	return
}

// Command returns command for running the client
func (t *LighthouseValidatorClient) Command() (command []string) {
	command = []string{"lighthouse", "vc"}
	return
}

// Image returns prysm docker image
func (t *LighthouseValidatorClient) Image() string {
	if img := t.validator.Spec.Image; img != nil {
		return *img
	} else if os.Getenv(EnvLighthouseValidatorImage) == "" {
		return DefaultLighthouseValidatorImage
	}
	return os.Getenv(EnvLighthouseValidatorImage)
}
