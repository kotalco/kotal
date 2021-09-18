package polkadot

import (
	polkadotv1alpha1 "github.com/kotalco/kotal/apis/polkadot/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Polkadot client arguments", func() {

	It("Should generate correct client arguments", func() {
		node := &polkadotv1alpha1.Node{
			Spec: polkadotv1alpha1.NodeSpec{
				Network:  "kusama",
				SyncMode: "fast",
				Logging:  "warn",
			},
		}

		node.Default()
		client := NewClient(node)
		args := client.Args()

		Expect(args).To(ContainElements([]string{
			PolkadotArgBasePath,
			shared.PathData(client.HomeDir()),
			PolkadotArgChain,
			"kusama",
			PolkadotArgLogging,
			string(polkadotv1alpha1.WarnLogs),
		}))

	})

})
