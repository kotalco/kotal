package chainlink

import (
	"os"

	chainlinkv1alpha1 "github.com/kotalco/kotal/apis/chainlink/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Chainlink Client", func() {
	node := &chainlinkv1alpha1.Node{
		Spec: chainlinkv1alpha1.NodeSpec{
			EthereumChainId:     1,
			EthereumWSEndpoint:  "ws://my-eth-node:8546",
			LinkContractAddress: "0x01BE23585060835E02B77ef475b0Cc51aA1e0709",
		},
	}

	client := NewClient(node)

	It("Should get correct image", func() {
		// default image
		img := client.Image()
		Expect(img).To(Equal(DefaultChainlinkImage))
		// after setting custom image
		testImage := "kotalco/chainlink:test"
		os.Setenv(EnvChainlinkImage, testImage)
		img = client.Image()
		Expect(img).To(Equal(testImage))
	})

	It("Should get correct command", func() {
		Expect(client.Command()).To(ConsistOf("chainlink"))
	})

	It("Should get correct environment variables", func() {
		Expect(client.Env()).To(ContainElements(
			corev1.EnvVar{
				Name:  EnvRoot,
				Value: "/",
			},
			corev1.EnvVar{
				Name:  EnvChainID,
				Value: "1",
			},
			corev1.EnvVar{
				Name:  EnvEthereumURL,
				Value: "ws://my-eth-node:8546",
			},
			corev1.EnvVar{
				Name:  EnvLinkContractAddress,
				Value: "0x01BE23585060835E02B77ef475b0Cc51aA1e0709",
			},
		))
	})

	It("Should get correct home dir", func() {
		Expect(client.HomeDir()).To(Equal(ChainlinkHomeDir))
	})

	It("Should get correct args", func() {
		Expect(client.Args()).To(ContainElements(
			"local",
			"node",
		))
	})

})
