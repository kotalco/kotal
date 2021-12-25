package polkadot

import (
	"fmt"

	polkadotv1alpha1 "github.com/kotalco/kotal/apis/polkadot/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Polkadot client arguments", func() {

	It("Should generate correct client arguments", func() {
		t := false
		node := &polkadotv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kusama-node",
				Namespace: "default",
			},
			Spec: polkadotv1alpha1.NodeSpec{
				Network:                  "kusama",
				P2PPort:                  4444,
				NodePrivateKeySecretName: "kusama-node-key",
				Validator:                true,
				SyncMode:                 "fast",
				Logging:                  "warn",
				RPC:                      true,
				RPCPort:                  6789,
				WS:                       true,
				WSPort:                   3456,
				Telemetry:                true,
				TelemetryURL:             "wss://telemetry.kotal.io/submit/ 0",
				Prometheus:               true,
				PrometheusPort:           5432,
				Pruning:                  &t,
				CORSDomains:              []string{"kotal.com"},
				// TODO: create test for node with telemetry disabled
				// TODO: create test for node with prometheus disabled
				// TODO: create test for node with pruning true
			},
		}

		node.Default()
		client := NewClient(node)
		args := client.Args()

		Expect(args).To(ContainElements([]string{
			PolkadotArgBasePath,
			shared.PathData(client.HomeDir()),
			PolkadotArgChain,
			"kusama",
			PolkadotArgName,
			"kusama-node",
			PolkadotArgPort,
			"4444",
			PolkadotArgValidator,
			PolkadotArgLogging,
			string(sharedAPI.WarnLogs),
			PolkadotArgRPCExternal,
			PolkadotArgRPCPort,
			"6789",
			PolkadotArgWSExternal,
			PolkadotArgWSPort,
			"3456",
			PolkadotArgNodeKeyType,
			"Ed25519",
			PolkadotArgNodeKeyFile,
			fmt.Sprintf("%s/kotal_nodekey", shared.PathData(client.HomeDir())),
			PolkadotArgTelemetryURL,
			"wss://telemetry.kotal.io/submit/ 0",
			PolkadotArgPrometheusExternal,
			PolkadotArgPrometheusPort,
			"5432",
			PolkadotArgPruning,
			"archive",
			PolkadotArgRPCCors,
			"kotal.com",
		}))

	})

})
