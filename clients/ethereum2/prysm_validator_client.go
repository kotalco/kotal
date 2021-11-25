package ethereum2

import (
	"fmt"
	"os"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// PrysmValidatorClient is Prysmatic labs validator client
type PrysmValidatorClient struct {
	validator *ethereum2v1alpha1.Validator
}

// Images
const (
	// EnvPrysmValidatorImage is the environment variable used for Prysmatic Labs validator client image
	EnvPrysmValidatorImage = "PRYSM_VALIDATOR_CLIENT_IMAGE"
	// DefaultPrysmValidatorImage is Prysmatic Labs validator client image
	DefaultPrysmValidatorImage = "kotalco/prysm:v2.0.2"
)

// HomeDir returns container home directory
func (t *PrysmValidatorClient) HomeDir() string {
	return PrysmHomeDir
}

// Command returns environment variables for the client
func (t *PrysmValidatorClient) Env() []corev1.EnvVar {
	return nil
}

// Args returns command line arguments required for client
func (t *PrysmValidatorClient) Args() (args []string) {

	validator := t.validator

	args = append(args, PrysmAcceptTermsOfUse)

	args = append(args, PrysmDataDir, shared.PathData(t.HomeDir()))

	args = append(args, PrysmLogging, string(t.validator.Spec.Logging))

	args = append(args, PrysmWalletDir, fmt.Sprintf("%s/prysm-wallet", shared.PathData(t.HomeDir())))

	args = append(args, PrysmWalletPasswordFile, fmt.Sprintf("%s/prysm-wallet/prysm-wallet-password.txt", shared.PathSecrets(t.HomeDir())))

	args = append(args, fmt.Sprintf("--%s", validator.Spec.Network))

	if len(validator.Spec.BeaconEndpoints) != 0 {
		args = append(args, PrysmBeaconRPCProvider, validator.Spec.BeaconEndpoints[0])
	}

	if validator.Spec.Graffiti != "" {
		args = append(args, PrysmGraffiti, validator.Spec.Graffiti)
	}

	if validator.Spec.CertSecretName != "" {
		args = append(args, PrysmTLSCert, fmt.Sprintf("%s/cert/tls.crt", shared.PathSecrets(t.HomeDir())))
	}

	return args
}

// Command returns command for running the client
func (t *PrysmValidatorClient) Command() (command []string) {
	command = []string{"validator"}
	return
}

// Image returns prysm docker image
func (t *PrysmValidatorClient) Image() string {
	if os.Getenv(EnvPrysmValidatorImage) == "" {
		return DefaultPrysmValidatorImage
	}
	return os.Getenv(EnvPrysmValidatorImage)
}
