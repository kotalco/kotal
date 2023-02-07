package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Aptos node defaulting", func() {
	It("Should default Aptos node", func() {
		node := Node{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: NodeSpec{
				Network:   Devnet,
				Validator: true,
			},
		}
		// TODO: create a test for full node, p2p port

		node.Default()

		Expect(node.Spec.Image).To(Equal(DefaultAptosCoreDevnetImage))
		Expect(node.Spec.CPU).To(Equal(DefaultNodeCPURequest))
		Expect(node.Spec.CPULimit).To(Equal(DefaultNodeCPULimit))
		Expect(node.Spec.Memory).To(Equal(DefaultNodeMemoryRequest))
		Expect(node.Spec.MemoryLimit).To(Equal(DefaultNodeMemoryLimit))
		Expect(node.Spec.Storage).To(Equal(DefaultNodeStorageRequest))
		Expect(node.Spec.APIPort).To(Equal(DefaultAPIPort))
		Expect(node.Spec.P2PPort).To(Equal(DefaultValidatorP2PPort))

	})
})
