package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("IPFS cluster peer defaulting", func() {
	It("Should default ipfs cluster peer", func() {
		peer := ClusterPeer{
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       ClusterPeerSpec{},
		}

		peer.Default()

		Expect(peer.Spec.Image).To(Equal(DefaultGoIPFSClusterImage))
		Expect(peer.Spec.Logging).To(Equal(DefaultLogging))
		Expect(peer.Spec.Resources.CPU).To(Equal(DefaultNodeCPURequest))
		Expect(peer.Spec.Resources.CPULimit).To(Equal(DefaultNodeCPULimit))
		Expect(peer.Spec.Resources.Memory).To(Equal(DefaultNodeMemoryRequest))
		Expect(peer.Spec.Resources.MemoryLimit).To(Equal(DefaultNodeMemoryLimit))
		Expect(peer.Spec.Resources.Storage).To(Equal(DefaultNodeStorageRequest))
		Expect(peer.Spec.Consensus).To(Equal(DefaultIPFSClusterConsensus))
	})
})
