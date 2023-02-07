package graph

import (
	graphv1alpha1 "github.com/kotalco/kotal/apis/graph/v1alpha1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Graph node client", func() {

	node := &graphv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "graph-node",
			Namespace: "default",
		},
		Spec: graphv1alpha1.NodeSpec{},
	}

	// TODO: default node

	client := NewClient(node)

	It("Should get correct command", func() {
		Expect(client.Command()).To(Equal(
			[]string{
				GraphNodeCommand,
			},
		))
	})

	It("Should get correct home directory", func() {
		Expect(client.HomeDir()).To(Equal(GraphNodeHomeDir))
	})

	It("Should generate correct client arguments", func() {
		Expect(client.Args()).To(ContainElements(
			[]string{},
		))
	})

})
