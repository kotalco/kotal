package aptos

import (
	"fmt"

	aptosv1alpha1 "github.com/kotalco/kotal/apis/aptos/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Aptos core client", func() {

	node := &aptosv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aptos-node",
			Namespace: "default",
		},
		Spec: aptosv1alpha1.NodeSpec{
			Network: aptosv1alpha1.Testnet,
		},
	}

	node.Default()

	client := NewClient(node)

	It("Should get correct command", func() {
		Expect(client.Command()).To(ConsistOf("aptos-node"))
	})

	It("Should get correct environment variables", func() {
		Expect(client.Env()).To(BeNil())
	})

	It("Should get correct home directory", func() {
		Expect(client.HomeDir()).To(Equal(AptosCoreHomeDir))
	})

	It("Should generate correct client arguments", func() {

		Expect(client.Args()).To(ContainElements([]string{
			AptosArgConfig,
			fmt.Sprintf("%s/config.yaml", shared.PathConfig(client.HomeDir())),
		}))
	})

})
