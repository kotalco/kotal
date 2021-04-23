package controllers

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prysm Ethereum 2.0 validator client arguments", func() {

	It("Should generate correct client arguments", func() {
		validator := &ethereum2v1alpha1.Validator{
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
		}

		validator.Default()
		client, _ := NewEthereum2Client(validator)
		args := client.Args()

		Expect(args).To(ContainElements([]string{
			PrysmAcceptTermsOfUse,
			PrysmDataDir,
			shared.PathData(client.HomeDir()),
			"--mainnet",
			PrysmBeaconRPCProvider,
			"http://localhost:8899",
			PrysmGraffiti,
			"Validated by Kotal",
			PrysmWalletDir,
			fmt.Sprintf("%s/prysm-wallet", shared.PathData(client.HomeDir())),
			PrysmWalletPasswordFile,
			fmt.Sprintf("%s/prysm-wallet/prysm-wallet-password.txt", shared.PathSecrets(client.HomeDir())),
		}))

	})

})
