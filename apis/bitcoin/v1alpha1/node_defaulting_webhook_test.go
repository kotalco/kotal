package v1alpha1

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Bitcoin node defaulting", func() {
	It("Should default Bitcoin node", func() {
		node := Node{
			ObjectMeta: metav1.ObjectMeta{},
			Spec: NodeSpec{
				Network: Mainnet,
			},
		}

		node.Default()

		Expect(node.Spec.RPCPort).To(Equal(DefaultMainnetRPCPort))
		Expect(node.Spec.RPCHost).To(Equal(DefaultRPCHost))

	})
})
