package near

import (
	nearv1alpha1 "github.com/kotalco/kotal/apis/near/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("NEAR core client", func() {

	node := &nearv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "near-node",
			Namespace: "default",
		},
		// TODO: create test for rpc: false node
		Spec: nearv1alpha1.NodeSpec{
			Network:        "mainnet",
			MinPeers:       77,
			P2PPort:        3334,
			Archive:        true,
			RPC:            true,
			RPCPort:        7444,
			PrometheusPort: 9991,
			TelemetryURL:   "https://explorer.mainnet.near.org/api/nodes",
			Bootnodes: []string{
				"ed25519:86EtEy7epneKyrcJwSWP7zsisTkfDRH5CFVszt4qiQYw@35.195.32.249:24567",
				"ed25519:BFB78VTDBBfCY4jCP99zWxhXUcFAZqR22oSx2KEr8UM1@35.229.222.235:24567",
			},
		},
	}

	node.Default()
	client := NewClient(node)

	It("Should get correct command", func() {
		Expect(client.Command()).To(BeNil())
	})

	It("Should get correct home directory", func() {
		Expect(client.HomeDir()).To(Equal(NearHomeDir))
	})

	It("Should generate correct client arguments", func() {
		Expect(client.Args()).To(ContainElements([]string{
			"neard",
			NearArgHome,
			shared.PathData(client.HomeDir()),
			"run",
			NearArgMinimumPeers,
			"77",
			NearArgNetworkAddress,
			"0.0.0.0:3334",
			NearArgArchive,
			NearArgRPCAddress,
			"0.0.0.0:7444",
			NearArgPrometheusAddress,
			"0.0.0.0:9991",
			NearArgBootnodes,
			"ed25519:86EtEy7epneKyrcJwSWP7zsisTkfDRH5CFVszt4qiQYw@35.195.32.249:24567,ed25519:BFB78VTDBBfCY4jCP99zWxhXUcFAZqR22oSx2KEr8UM1@35.229.222.235:24567",
			NearArgTelemetryURL,
			"https://explorer.mainnet.near.org/api/nodes",
		}))

	})

})
