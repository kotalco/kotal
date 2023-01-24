package ethereum2

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// PrysmValidatorClient is Prysmatic labs validator client
// https://github.com/prysmaticlabs/prysm
type PrysmValidatorClient struct {
	validator *ethereum2v1alpha1.Validator
}

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

	args = append(args, PrysmFeeRecipient, string(t.validator.Spec.FeeRecipient))

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
