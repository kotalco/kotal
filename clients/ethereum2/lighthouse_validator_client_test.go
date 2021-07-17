package ethereum2

import (
	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lighthouse Ethereum 2.0 validator client arguments", func() {

	It("Should generate correct client arguments", func() {
		validator := &ethereum2v1alpha1.Validator{
			Spec: ethereum2v1alpha1.ValidatorSpec{
				Client:  ethereum2v1alpha1.LighthouseClient,
				Network: "mainnet",
				BeaconEndpoints: []string{
					"http://localhost:8899",
					"http://localhost:9988",
				},
				Graffiti: "Validated by Kotal",
			},
		}

		validator.Default()
		client, _ := NewClient(validator)
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
		}))
	})

})
