package ethereum2

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
				Client:          ethereum2v1alpha1.PrysmClient,
				Network:         "mainnet",
				BeaconEndpoints: []string{"http://localhost:8899"},
				Graffiti:        "Validated by Kotal",
				Keystores: []ethereum2v1alpha1.Keystore{
					{
						SecretName: "my-validator",
					},
				},
				WalletPasswordSecret: "wallet-password",
				CertSecretName:       "my-cert",
				Logging:              ethereum2v1alpha1.ErrorLogs,
			},
		}

		validator.Default()
		client, _ := NewClient(validator)
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
			PrysmLogging,
			string(ethereum2v1alpha1.ErrorLogs),
			PrysmTLSCert,
			fmt.Sprintf("%s/cert/tls.crt", shared.PathSecrets(client.HomeDir())),
		}))

	})

})
