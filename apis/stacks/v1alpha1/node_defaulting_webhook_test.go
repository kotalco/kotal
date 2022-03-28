package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Stacks node defaulting", func() {
	It("Should default Stacks node", func() {
		node := Node{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: NodeSpec{
				Network: Mainnet,
			},
		}

		node.Default()

		Expect(node.Spec.RPCHost).To(Equal(DefaultRPCHost))
		Expect(node.Spec.RPCPort).To(Equal(DefaultRPCPort))
		Expect(node.Spec.CPU).To(Equal(DefaultNodeCPURequest))
		Expect(node.Spec.CPULimit).To(Equal(DefaultNodeCPULimit))
		Expect(node.Spec.Memory).To(Equal(DefaultNodeMemoryRequest))
		Expect(node.Spec.MemoryLimit).To(Equal(DefaultNodeMemoryLimit))
		Expect(node.Spec.Storage).To(Equal(DefaultNodeStorageRequest))

	})
})
