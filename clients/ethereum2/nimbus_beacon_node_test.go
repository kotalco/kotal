package ethereum2

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nimbus beacon node", func() {

	node := ethereum2v1alpha1.BeaconNode{
		Spec: ethereum2v1alpha1.BeaconNodeSpec{
			Client:  ethereum2v1alpha1.NimbusClient,
			Network: "mainnet",
		},
	}
	client, _ := NewClient(&node)

	It("Should get correct command", func() {
		Expect(client.Command()).To(ConsistOf("nimbus_beacon_node"))
	})

	It("Should get correct env", func() {
		Expect(client.Env()).To(BeNil())
	})

	It("Should get correct home dir", func() {
		Expect(client.HomeDir()).To(Equal(NimbusHomeDir))
	})

	cases := []struct {
		title  string
		node   *ethereum2v1alpha1.BeaconNode
		result []string
	}{
		{
			title: "beacon node syncing mainnet",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:  ethereum2v1alpha1.NimbusClient,
					Network: "mainnet",
					Logging: sharedAPI.DebugLogs,
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusLogging, string(sharedAPI.DebugLogs)),
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.NimbusClient,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
					REST:                    true,
					RESTPort:                8957,
					CORSDomains:             []string{"kotal.pro", "kotal.cloud"},
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusExecutionEngineEndpoint, "https://localhost:8551"),
				argWithVal(NimbusJwtSecretFile, fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir()))),
				argWithVal(NimbusRESTAddress, "0.0.0.0"),
				argWithVal(NimbusRESTPort, "8957"),
				argWithVal(NimbusRESTAllowOrigin, "kotal.pro,kotal.cloud"),
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.NimbusClient,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusExecutionEngineEndpoint, "https://localhost:8551"),
				argWithVal(NimbusJwtSecretFile, fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir()))),
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.NimbusClient,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusExecutionEngineEndpoint, "https://localhost:8551"),
				argWithVal(NimbusJwtSecretFile, fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir()))),
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.NimbusClient,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusExecutionEngineEndpoint, "https://localhost:8551"),
				argWithVal(NimbusJwtSecretFile, fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir()))),
			},
		},
		{
			title: "beacon node syncing mainnet with p2p port, eth1 endpoint",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.NimbusClient,
					P2PPort:                 7891,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
					FeeRecipient:            "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusTCPPort, "7891"),
				argWithVal(NimbusUDPPort, "7891"),
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusExecutionEngineEndpoint, "https://localhost:8551"),
				argWithVal(NimbusJwtSecretFile, fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir()))),
				argWithVal(NimbusFeeRecipient, "0xd8da6bf26964af9d7eed9e03e53415d37aa96045"),
			},
		},
	}

	for _, c := range cases {
		func() {
			cc := c
			It(fmt.Sprintf("Should create correct client arguments for %s", cc.title), func() {
				cc.node.Default()
				client, _ := NewClient(cc.node)
				args := client.Args()
				Expect(args).To(ContainElements(cc.result))
			})
		}()
	}

})
