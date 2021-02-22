package controllers

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Teku Ethereum 2.0 validator client arguments", func() {

	cases := []struct {
		title     string
		validator *ethereum2v1alpha1.Validator
		result    []string
	}{
		{
			title: "mainnet validator client",
			validator: &ethereum2v1alpha1.Validator{
				Spec: ethereum2v1alpha1.ValidatorSpec{
					Client:         ethereum2v1alpha1.TekuClient,
					Network:        "mainnet",
					BeaconEndpoint: "http://localhost:9988",
					Graffiti:       "Validated by Kotal",
					Secrets: []string{
						"my-validator",
					},
				},
			},
			result: []string{
				"vc",
				TekuDataPath,
				PathBlockchainData,
				TekuNetwork,
				"mainnet",
				TekuBeaconNodeEndpoint,
				"http://localhost:9988",
				TekuGraffiti,
				"Validated by Kotal",
				TekuValidatorKeys,
				"/mnt/secrets/validator-keys/my-validator/keystore-0.json:/mnt/secrets/validator-keys/my-validator/password.txt",
			},
		},
	}

	for _, c := range cases {
		func() {
			cc := c
			It(fmt.Sprintf("Should create correct client arguments for %s", cc.title), func() {
				cc.validator.Default()
				client, err := NewValidatorClient(cc.validator.Spec.Client)
				Expect(err).To(BeNil())
				args := client.Args(cc.validator)
				Expect(args).To(ContainElements(cc.result))
			})
		}()
	}

})
