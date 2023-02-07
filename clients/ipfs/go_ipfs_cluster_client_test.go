package ipfs

import (
	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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

	It("Should get correct env", func() {
		Expect(client.Env()).To(Equal(
			[]corev1.EnvVar{
				{
					Name:  EnvIPFSClusterPath,
					Value: shared.PathData(client.HomeDir()),
				},
				{
					Name:  EnvIPFSClusterPeerName,
					Value: peer.Name,
				},
				{
					Name:  EnvIPFSLogging,
					Value: string(peer.Spec.Logging),
				},
			},
		))
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
