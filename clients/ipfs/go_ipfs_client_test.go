package ipfs

import (
	"os"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Go IPFS Client", func() {
	peer := &ipfsv1alpha1.Peer{
		Spec: ipfsv1alpha1.PeerSpec{
			Routing: ipfsv1alpha1.DHTClientRouting,
		},
	}

	client, _ := NewClient(peer)

	It("Should get correct image", func() {
		// default image
		img := client.Image()
		Expect(img).To(Equal(DefaultGoIPFSImage))
		// after changing .spec.image
		testImage := "kotalco/go-ipfs:spec"
		peer.Spec.Image = &testImage
		img = client.Image()
		Expect(img).To(Equal(testImage))
		// after setting custom image
		testImage = "kotalco/go-ipfs:test"
		os.Setenv(EnvGoIPFSImage, testImage)
		img = client.Image()
		Expect(img).To(Equal(testImage))
	})

	It("Should get correct command", func() {
		Expect(client.Command()).To(ConsistOf("ipfs"))
	})

	It("Should get correct env", func() {
		Expect(client.Env()).To(Equal(
			[]corev1.EnvVar{
				{
					Name:  EnvIPFSPath,
					Value: shared.PathData(client.HomeDir()),
				},
				{
					Name:  EnvIPFSLogging,
					Value: string(peer.Spec.Logging),
				},
			},
		))
	})

	It("Should get correct home dir", func() {
		Expect(client.HomeDir()).To(Equal(GoIPFSHomeDir))
	})

	It("Should get correct args", func() {
		Expect(client.Args()).To(ContainElements(
			GoIPFSDaemonArg,
			GoIPFSRoutingArg,
			string(ipfsv1alpha1.DHTClientRouting),
		))
	})

})
