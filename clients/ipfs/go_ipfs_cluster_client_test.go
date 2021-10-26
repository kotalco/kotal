package ipfs

import (
	"os"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Go IPFS Cluster Client", func() {
	peer := &ipfsv1alpha1.ClusterPeer{
		Spec: ipfsv1alpha1.ClusterPeerSpec{
			Consensus:    ipfsv1alpha1.Raft,
			PeerEndpoint: "/dns4/bare-peer/tcp/5001",
			BootstrapPeers: []string{
				"/ip4/95.111.253.236/tcp/4001/p2p/Qmd3FERyCvxvkC8su1DYhjybRaLueHveKysUVPxWAqR4U7",
			},
			ClusterSecretName: "cluster-secret",
		},
	}

	client, _ := NewClient(peer)

	It("Should get correct image", func() {
		// default image
		img := client.Image()
		Expect(img).To(Equal(DefaultGoIPFSClusterImage))
		// after setting custom image
		testImage := "kotalco/ipfs-cluster:test"
		os.Setenv(EnvGoIPFSClusterImage, testImage)
		img = client.Image()
		Expect(img).To(Equal(testImage))
	})

	It("Should get correct command", func() {
		Expect(client.Command()).To(ConsistOf("ipfs-cluster-service"))
	})

	It("Should get correct home dir", func() {
		Expect(client.HomeDir()).To(Equal(GoIPFSClusterHomeDir))
	})

	It("Should get correct args", func() {
		Expect(client.Args()).To(ContainElements(
			GoIPFSDaemonArg,
			GoIPFSClusterBootstrapArg,
			"/ip4/95.111.253.236/tcp/4001/p2p/Qmd3FERyCvxvkC8su1DYhjybRaLueHveKysUVPxWAqR4U7",
		))
	})

})
