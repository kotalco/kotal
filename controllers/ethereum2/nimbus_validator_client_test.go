package controllers

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nimbus Ethereum 2.0 validator client arguments", func() {

	client, _ := NewValidatorClient(ethereum2v1alpha1.NimbusClient)

	cases := []struct {
		title     string
		validator *ethereum2v1alpha1.Validator
		result    []string
	}{
		{
			title: "mainnet validator client",
			validator: &ethereum2v1alpha1.Validator{
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
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusDataDir, shared.PathData(client.HomeDir())),
				argWithVal(NimbusRPCAddress, "http://10.0.0.11"),
				argWithVal(NimbusRPCPort, "80"),
				argWithVal(NimbusGraffiti, "Validated by Kotal"),
				argWithVal(NimbusValidatorsDir, fmt.Sprintf("%s/kotal-validators/validator-keys", shared.PathData(client.HomeDir()))),
				argWithVal(NimbusSecretsDir, fmt.Sprintf("%s/kotal-validators/validator-secrets", shared.PathData(client.HomeDir()))),
			},
		},
	}

	for _, c := range cases {
		func() {
			cc := c
			It(fmt.Sprintf("Should create correct client arguments for %s", cc.title), func() {
				cc.validator.Default()
				args := client.Args(cc.validator)
				Expect(args).To(ContainElements(cc.result))
			})
		}()
	}

})
