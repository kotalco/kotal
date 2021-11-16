package ethereum2

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nimbus Ethereum 2.0 validator client arguments", func() {

	It("Should generate correct client arguments", func() {
		validator := &ethereum2v1alpha1.Validator{
			Spec: ethereum2v1alpha1.ValidatorSpec{
				Client:          ethereum2v1alpha1.NimbusClient,
				Network:         "mainnet",
				BeaconEndpoints: []string{"http://nimbus-beacon-node"},
				Graffiti:        "Validated by Kotal",
				Keystores: []ethereum2v1alpha1.Keystore{
					{
						SecretName: "my-validator",
					},
				},
				Logging: ethereum2v1alpha1.FatalLogs,
			},
		}

		validator.Default()
		client, _ := NewClient(validator)
		args := client.Args()

		Expect(args).To(ContainElements([]string{
			NimbusNonInteractive,
			argWithVal(NimbusLogging, string(validator.Spec.Logging)),
			argWithVal(NimbusDataDir, shared.PathData(client.HomeDir())),
			argWithVal(NimbusBeaconNodes, "http://nimbus-beacon-node"),
			argWithVal(NimbusGraffiti, "Validated by Kotal"),
			argWithVal(NimbusValidatorsDir, fmt.Sprintf("%s/kotal-validators/validator-keys", shared.PathData(client.HomeDir()))),
			argWithVal(NimbusSecretsDir, fmt.Sprintf("%s/kotal-validators/validator-secrets", shared.PathData(client.HomeDir()))),
		}))

	})

})
