package near

import (
	nearv1alpha1 "github.com/kotalco/kotal/apis/near/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NEAR core client", func() {

	It("Should generate correct client arguments", func() {
		node := &nearv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "near-node",
				Namespace: "default",
			},
			// TODO: create test for rpc: false node
			Spec: nearv1alpha1.NodeSpec{
				Network: "mainnet",
				RPC:     true,
				RPCPort: 7444,
				RPCHost: "127.0.0.1",
			},
		}

		node.Default()
		client := NewClient(node)
		args := client.Args()

		Expect(args).To(ContainElements([]string{
			"neard",
			NearArgHome,
			client.HomeDir(),
			"run",
			NearArgRPCAddress,
			"127.0.0.1:7444",
		}))

	})

})
