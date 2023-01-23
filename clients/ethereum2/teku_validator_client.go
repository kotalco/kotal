package ethereum2

import (
	"fmt"
	"os"
	"strings"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// TekuValidatorClient is Teku validator client
// https://github.com/Consensys/teku/
type TekuValidatorClient struct {
	validator *ethereum2v1alpha1.Validator
}

const (
	// EnvTekuValidatorImage is the environment variable used for PegaSys Teku validator client image
	EnvTekuValidatorImage = "TEKU_VALIDATOR_CLIENT_IMAGE"
)

// HomeDir returns container home directory
func (t *TekuValidatorClient) HomeDir() string {
	return TekuHomeDir
}

// Command returns environment variables for running the client
func (t *TekuValidatorClient) Env() []corev1.EnvVar {
	return nil
}

// Args returns command line arguments required for client
func (t *TekuValidatorClient) Args() (args []string) {

	validator := t.validator

	args = append(args, TekuVC)

	args = append(args, TekuDataPath, shared.PathData(t.HomeDir()))

	args = append(args, TekuNetwork, "auto")

	args = append(args, TekuValidatorsKeystoreLockingEnabled, "false")

	args = append(args, TekuFeeRecipient, string(validator.Spec.FeeRecipient))

	if len(validator.Spec.BeaconEndpoints) != 0 {
		args = append(args, TekuBeaconNodeEndpoint, validator.Spec.BeaconEndpoints[0])
	}

	if validator.Spec.Graffiti != "" {
		args = append(args, TekuGraffiti, validator.Spec.Graffiti)
	}

	keyPass := []string{}
	for i, keystore := range validator.Spec.Keystores {
		path := fmt.Sprintf("%s/validator-keys/%s", shared.PathSecrets(t.HomeDir()), keystore.SecretName)
		keyPass = append(keyPass, fmt.Sprintf("%s/keystore-%d.json:%s/password.txt", path, i, path))
	}

	args = append(args, TekuValidatorKeys, strings.Join(keyPass, ","))

	return args
}

// Command returns command for running the client
func (t *TekuValidatorClient) Command() (command []string) {
	return
}

// Image returns teku docker image
func (t *TekuValidatorClient) Image() string {
	return os.Getenv(EnvTekuValidatorImage)
}
