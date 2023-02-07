package stacks

import (
	"fmt"

	stacksv1alpha1 "github.com/kotalco/kotal/apis/stacks/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Stacks node client", func() {

	node := &stacksv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "stacks-node",
			Namespace: "default",
		},
		Spec: stacksv1alpha1.NodeSpec{
			Network: "mainnet",
		},
	}

	node.Default()

	client := NewClient(node)

	It("Should get correct command", func() {
		Expect(client.Command()).To(Equal(
			[]string{
				StacksNodeCommand,
				StacksStartCommand,
			},
		))
	})

	It("Should get correct home directory", func() {
		Expect(client.HomeDir()).To(Equal(StacksNodeHomeDir))
	})

	It("Should generate correct client arguments", func() {
		Expect(client.Args()).To(ContainElements(
			[]string{
				StacksArgConfig,
				fmt.Sprintf("%s/config.toml", shared.PathConfig(client.HomeDir())),
			},
		))
	})

})
