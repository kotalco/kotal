package controllers

import (
	"fmt"
	"os"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"

	"gopkg.in/yaml.v2"
)

// LighthouseValidatorClient is SigmaPrime Ethereum 2.0 validator client
type LighthouseValidatorClient struct{}

// Images
const (
	// EnvLighthouseValidatorImage is the environment variable used for SigmaPrime Ethereum 2.0 validator client image
	EnvLighthouseValidatorImage = "LIGHTHOUSE_VALIDATOR_CLIENT_IMAGE"
	// DefaultLighthouseValidatorImage is the default SigmaPrime Ethereum 2.0 validator client image
	DefaultLighthouseValidatorImage = "sigp/lighthouse:v1.1.3"
)

// HomeDir returns container home directory
func (t *LighthouseValidatorClient) HomeDir() string {
	return LighthouseHomeDir
}

// Args returns command line arguments required for client
func (t *LighthouseValidatorClient) Args(validator *ethereum2v1alpha1.Validator) (args []string) {

	args = append(args, LighthouseDataDir, shared.PathData(t.HomeDir()))

	args = append(args, LighthouseNetwork, validator.Spec.Network)

	args = append(args, LighthouseDisableAutoDiscover)

	args = append(args, LighthouseInitSlashingProtection)

	if validator.Spec.BeaconEndpoint != "" {
		args = append(args, LighthouseBeaconNodeEndpoint, validator.Spec.BeaconEndpoint)
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
	if os.Getenv(EnvLighthouseValidatorImage) == "" {
		return DefaultLighthouseValidatorImage
	}
	return os.Getenv(EnvLighthouseValidatorImage)
}

// ValidatorDefinition is a validator definition
// https://lighthouse-book.sigmaprime.io/validator-management.html
type ValidatorDefinition struct {
	VotingPublicKey            string `yaml:"voting_public_key"`
	Type                       string `yaml:"type"`
	Enabled                    bool   `yaml:"enabled"`
	VotingKeystorePath         string `yaml:"voting_keystore_path"`
	VotingKeystorePasswordPath string `yaml:"voting_keystore_password_path"`
}

// CreateValidatorDefinitions create validator definitions yaml file
// https://lighthouse-book.sigmaprime.io/validator-management.html
func (t *LighthouseValidatorClient) CreateValidatorDefinitions(validator *ethereum2v1alpha1.Validator) (data string, err error) {
	definitions := []ValidatorDefinition{}

	for i, keystore := range validator.Spec.Keystores {

		keystorePath := fmt.Sprintf("%s/validator-keys/%s", shared.PathData(t.HomeDir()), keystore.SecretName)

		definitions = append(definitions, ValidatorDefinition{
			VotingPublicKey:            keystore.PublicKey,
			Type:                       "local_keystore",
			Enabled:                    true,
			VotingKeystorePath:         fmt.Sprintf("%s/keystore-%d.json", keystorePath, i),
			VotingKeystorePasswordPath: fmt.Sprintf("%s/password.txt", keystorePath),
		})

	}

	out, err := yaml.Marshal(definitions)
	if err != nil {
		return
	}

	return string(out), nil
}
