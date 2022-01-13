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
				P2PHost: "127.0.0.1",
				P2PPort: 3334,
				RPC:     true,
				RPCPort: 7444,
				RPCHost: "127.0.0.1",
				Bootnodes: []string{
					"ed25519:86EtEy7epneKyrcJwSWP7zsisTkfDRH5CFVszt4qiQYw@35.195.32.249:24567",
					"ed25519:BFB78VTDBBfCY4jCP99zWxhXUcFAZqR22oSx2KEr8UM1@35.229.222.235:24567",
				},
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
			NearArgNetworkAddress,
			"127.0.0.1:3334",
			NearArgRPCAddress,
			"127.0.0.1:7444",
			NearArgBootnodes,
			"ed25519:86EtEy7epneKyrcJwSWP7zsisTkfDRH5CFVszt4qiQYw@35.195.32.249:24567,ed25519:BFB78VTDBBfCY4jCP99zWxhXUcFAZqR22oSx2KEr8UM1@35.229.222.235:24567",
		}))

	})

})
