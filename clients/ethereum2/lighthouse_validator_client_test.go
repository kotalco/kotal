package ethereum2

import (
	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
	"github.com/kotalco/kotal/controllers/shared"

	. "github.com/onsi/ginkgo/v2"
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
			Graffiti:     "Validated by Kotal",
			Logging:      sharedAPI.WarnLogs,
			FeeRecipient: "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
		},
	}

	validator.Default()
	client, _ := NewClient(validator)

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
			LighthouseFeeRecipient,
			"0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
		}))
	})

})
