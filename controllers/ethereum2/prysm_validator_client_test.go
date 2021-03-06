package controllers

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prysm Ethereum 2.0 validator client arguments", func() {

	client, _ := NewValidatorClient(ethereum2v1alpha1.PrysmClient)

	cases := []struct {
		title     string
		validator *ethereum2v1alpha1.Validator
		result    []string
	}{
		{
			title: "mainnet validator client",
			validator: &ethereum2v1alpha1.Validator{
				Spec: ethereum2v1alpha1.ValidatorSpec{
					Client:         ethereum2v1alpha1.PrysmClient,
					Network:        "mainnet",
					BeaconEndpoint: "http://localhost:8899",
					Graffiti:       "Validated by Kotal",
					Keystores: []ethereum2v1alpha1.Keystore{
						{
							SecretName: "my-validator",
						},
					},
					WalletPasswordSecret: "wallet-password",
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				PathBlockchainData(client.HomeDir()),
				"--mainnet",
				PrysmBeaconRPCProvider,
				"http://localhost:8899",
				PrysmGraffiti,
				"Validated by Kotal",
				PrysmWalletDir,
				fmt.Sprintf("%s/prysm-wallet", PathBlockchainData(client.HomeDir())),
				PrysmWalletPasswordFile,
				fmt.Sprintf("%s/prysm-wallet/prysm-wallet-password.txt", PathSecrets(client.HomeDir())),
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
