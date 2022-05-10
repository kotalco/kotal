package ethereum2

import (
	"os"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
	"github.com/kotalco/kotal/controllers/shared"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lighthouse validator client", func() {

	validator := &ethereum2v1alpha1.Validator{
		Spec: ethereum2v1alpha1.ValidatorSpec{
			Client:  ethereum2v1alpha1.LighthouseClient,
			Network: "mainnet",
			BeaconEndpoints: []string{
				"http://localhost:8899",
				"http://localhost:9988",
			},
			Graffiti: "Validated by Kotal",
			Logging:  sharedAPI.WarnLogs,
		},
	}

	validator.Default()
	client, _ := NewClient(validator)

	It("Should get correct image", func() {
		// default image
		img := client.Image()
		Expect(img).To(Equal(DefaultLighthouseValidatorImage))
		// after changing .spec.image
		testImage := "kotalco/lighthouse:spec"
		validator.Spec.Image = &testImage
		img = client.Image()
		Expect(img).To(Equal(testImage))
		// after setting custom image
		testImage = "kotalco/lighthouse:test"
		os.Setenv(EnvLighthouseValidatorImage, testImage)
		img = client.Image()
		Expect(img).To(Equal(testImage))
	})

	It("Should get correct command", func() {
		Expect(client.Command()).To(ConsistOf("lighthouse", "vc"))
	})

	It("Should get correct env", func() {
		Expect(client.Env()).To(BeNil())
	})

	It("Should get correct home dir", func() {
		Expect(client.HomeDir()).To(Equal(LighthouseHomeDir))
	})

	It("Should generate correct client arguments", func() {
		args := client.Args()

		Expect(args).To(ContainElements([]string{
			LighthouseDataDir,
			shared.PathData(client.HomeDir()),
			LighthouseNetwork,
			"mainnet",
			LighthouseBeaconNodeEndpoints,
			"http://localhost:8899,http://localhost:9988",
			LighthouseGraffiti,
			"Validated by Kotal",
			LighthouseDebugLevel,
			string(sharedAPI.WarnLogs),
		}))
	})

})
