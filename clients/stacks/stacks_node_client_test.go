package stacks

import (
	"os"

	stacksv1alpha1 "github.com/kotalco/kotal/apis/stacks/v1alpha1"
	. "github.com/onsi/ginkgo"
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

	// TODO: default node

	client := NewClient(node)

	It("Should get correct image", func() {
		// default image
		img := client.Image()
		Expect(img).To(Equal(DefaultStacksNodeImage))
		// after setting custom image
		testImage := "kotalco/stacks-node:test"
		os.Setenv(EnvStacksNodeImage, testImage)
		img = client.Image()
		Expect(img).To(Equal(testImage))
	})

	It("Should get correct command", func() {
		Expect(client.Command()).To(BeNil())
	})

	It("Should get correct home directory", func() {
		Expect(client.HomeDir()).To(Equal(StacksNodeHomeDir))
	})

	It("Should generate correct client arguments", func() {
		Expect(client.Args()).To(ContainElements([]string{}))
	})

})
