package ethereum2

import (
	"fmt"
	"os"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Teku Ethereum 2.0 validator client arguments", func() {

	validator := &ethereum2v1alpha1.Validator{
		Spec: ethereum2v1alpha1.ValidatorSpec{
			Client:          ethereum2v1alpha1.TekuClient,
			Network:         "mainnet",
			BeaconEndpoints: []string{"http://localhost:9988"},
			Graffiti:        "Validated by Kotal",
			FeeRecipient:    "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
			Keystores: []ethereum2v1alpha1.Keystore{
				{
					SecretName: "my-validator",
				},
			},
		},
	}

	validator.Default()
	client, _ := NewClient(validator)

	It("Should get correct image", func() {
		// default image
		img := client.Image()
		Expect(img).To(Equal(DefaultTekuValidatorImage))
		// after changing .spec.image
		testImage := "kotalco/teku:spec"
		validator.Spec.Image = &testImage
		img = client.Image()
		Expect(img).To(Equal(testImage))
		// after setting custom image
		testImage = "kotalco/teku:test"
		os.Setenv(EnvTekuValidatorImage, testImage)
		img = client.Image()
		Expect(img).To(Equal(testImage))
	})

	It("Should get correct command", func() {
		Expect(client.Command()).To(BeNil())
	})

	It("Should get correct env", func() {
		Expect(client.Env()).To(BeNil())
	})

	It("Should get correct home dir", func() {
		Expect(client.HomeDir()).To(Equal(TekuHomeDir))
	})

	It("Should generate correct client arguments", func() {
		args := client.Args()

		Expect(args).To(ContainElements([]string{
			"vc",
			TekuDataPath,
			shared.PathData(client.HomeDir()),
			TekuNetwork,
			"auto",
			TekuBeaconNodeEndpoint,
			"http://localhost:9988",
			TekuGraffiti,
			"Validated by Kotal",
			TekuValidatorKeys,
			fmt.Sprintf(
				"%s/validator-keys/my-validator/keystore-0.json:%s/validator-keys/my-validator/password.txt",
				shared.PathSecrets(client.HomeDir()),
				shared.PathSecrets(client.HomeDir()),
			),
			TekuFeeRecipient,
			"0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
		}))

	})

})
