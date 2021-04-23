package controllers

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
				Client:         ethereum2v1alpha1.NimbusClient,
				Network:        "mainnet",
				BeaconEndpoint: "http://10.0.0.11",
				Graffiti:       "Validated by Kotal",
				Keystores: []ethereum2v1alpha1.Keystore{
					{
						SecretName: "my-validator",
					},
				},
			},
		}

		validator.Default()
		client, _ := NewEthereum2Client(validator)
		args := client.Args()

		Expect(args).To(ContainElements([]string{
			NimbusNonInteractive,
			argWithVal(NimbusDataDir, shared.PathData(client.HomeDir())),
			argWithVal(NimbusRPCAddress, "http://10.0.0.11"),
			argWithVal(NimbusRPCPort, "80"),
			argWithVal(NimbusGraffiti, "Validated by Kotal"),
			argWithVal(NimbusValidatorsDir, fmt.Sprintf("%s/kotal-validators/validator-keys", shared.PathData(client.HomeDir()))),
			argWithVal(NimbusSecretsDir, fmt.Sprintf("%s/kotal-validators/validator-secrets", shared.PathData(client.HomeDir()))),
		}))

	})

})
