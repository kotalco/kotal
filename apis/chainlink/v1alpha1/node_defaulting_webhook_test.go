package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Chainlink node defaulting", func() {
	It("Should default node", func() {

		node := Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-node",
			},
			Spec: NodeSpec{
				CertSecretName: "my-certificate",
			},
		}

		node.Default()

		Expect(node.Spec.Image).To(Equal(DefaultChainlinkImage))
		Expect(node.Spec.TLSPort).To(Equal(DefaultTLSPort))
		Expect(node.Spec.P2PPort).To(Equal(DefaultP2PPort))
		Expect(node.Spec.APIPort).To(Equal(DefaultAPIPort))
		Expect(node.Spec.Logging).To(Equal(shared.InfoLogs))
		Expect(node.Spec.CORSDomains).To(Equal(DefaultCorsDomains))
		// resources
		Expect(node.Spec.Resources.CPU).To(Equal(DefaultNodeCPURequest))
		Expect(node.Spec.Resources.CPULimit).To(Equal(DefaultNodeCPULimit))
		Expect(node.Spec.Resources.Memory).To(Equal(DefaultNodeMemoryRequest))
		Expect(node.Spec.Resources.MemoryLimit).To(Equal(DefaultNodeMemoryLimit))
		Expect(node.Spec.Resources.Storage).To(Equal(DefaultNodeStorageRequest))

	})
})
