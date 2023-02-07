package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NEAR node defaulting", func() {
	It("Should default NEAR node", func() {
		node := Node{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: NodeSpec{
				Network: "mainnet",
				RPC:     true,
			},
		}

		node.Default()

		Expect(node.Spec.Image).To(Equal(DefaultNearImage))
		Expect(node.Spec.RPCPort).To(Equal(DefaultRPCPort))
		Expect(node.Spec.P2PPort).To(Equal(DefaultP2PPort))
		Expect(node.Spec.MinPeers).To(Equal(DefaultMinPeers))
		Expect(node.Spec.PrometheusPort).To(Equal(DefaultPrometheusPort))

		Expect(node.Spec.Resources.CPU).To(Equal(DefaultNodeCPURequest))
		Expect(node.Spec.Resources.CPULimit).To(Equal(DefaultNodeCPULimit))
		Expect(node.Spec.Resources.Memory).To(Equal(DefaultNodeMemoryRequest))
		Expect(node.Spec.Resources.MemoryLimit).To(Equal(DefaultNodeMemoryLimit))
		Expect(node.Spec.Resources.Storage).To(Equal(DefaultNodeStorageRequest))
	})
})
