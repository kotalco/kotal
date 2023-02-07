package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("IPFS peer defaulting", func() {
	It("Should default ipfs peer", func() {
		peer := Peer{
			ObjectMeta: metav1.ObjectMeta{},
			Spec:       PeerSpec{},
		}

		peer.Default()

		Expect(peer.Spec.Image).To(Equal(DefaultGoIPFSImage))
		Expect(peer.Spec.Logging).To(Equal(DefaultLogging))
		Expect(peer.Spec.InitProfiles).To(ContainElements(DefaultDatastoreProfile))
		Expect(peer.Spec.APIPort).To(Equal(DefaultAPIPort))
		Expect(peer.Spec.GatewayPort).To(Equal(DefaultGatewayPort))
		Expect(peer.Spec.Routing).To(Equal(DefaultRoutingMode))
		Expect(peer.Spec.Resources.CPU).To(Equal(DefaultNodeCPURequest))
		Expect(peer.Spec.Resources.CPULimit).To(Equal(DefaultNodeCPULimit))
		Expect(peer.Spec.Resources.Memory).To(Equal(DefaultNodeMemoryRequest))
		Expect(peer.Spec.Resources.MemoryLimit).To(Equal(DefaultNodeMemoryLimit))
		Expect(peer.Spec.Resources.Storage).To(Equal(DefaultNodeStorageRequest))
	})
})
