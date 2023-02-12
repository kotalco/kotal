package ethereum2

import (
	"fmt"
	"strings"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Teku beacon node", func() {

	node := ethereum2v1alpha1.BeaconNode{
		Spec: ethereum2v1alpha1.BeaconNodeSpec{
			Client:  ethereum2v1alpha1.TekuClient,
			Network: "mainnet",
		},
	}
	client, _ := NewClient(&node)

	It("Should get correct command", func() {
		Expect(client.Command()).To(BeNil())
	})

	It("Should get correct env", func() {
		Expect(client.Env()).To(BeNil())
	})

	It("Should get correct home dir", func() {
		Expect(client.HomeDir()).To(Equal(TekuHomeDir))
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
					Client:  ethereum2v1alpha1.TekuClient,
					Network: "mainnet",
					Logging: sharedAPI.ErrorLogs,
				},
			},
			result: []string{
				TekuDataPath,
				TekuNetwork,
				"mainnet",
				TekuLogging,
				strings.ToUpper(string(sharedAPI.ErrorLogs)),
			},
		},
		{
			title: "beacon node syncing mainnet with checkpoint syncing",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.TekuClient,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
					FeeRecipient:            "0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
					CheckpointSyncURL:       "https://kotal.cloud/eth2/beacon/checkpoint",
				},
			},
			result: []string{
				TekuDataPath,
				TekuNetwork,
				"mainnet",
				TekuExecutionEngineEndpoint,
				"https://localhost:8551",
				TekuJwtSecretFile,
				fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir())),
				TekuFeeRecipient,
				"0xd8da6bf26964af9d7eed9e03e53415d37aa96045",
				TekuInitialState,
				"https://kotal.cloud/eth2/beacon/checkpoint",
			},
		},
		{
			title: "beacon node syncing mainnet with http enabled",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.TekuClient,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
					REST:                    true,
				},
			},
			result: []string{
				TekuDataPath,
				TekuNetwork,
				"mainnet",
				TekuExecutionEngineEndpoint,
				"https://localhost:8551",
				TekuJwtSecretFile,
				fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir())),
				TekuRestEnabled,
				TekuRESTAPICorsOrigins,
				"*",
				TekuRESTAPIHostAllowlist,
				"*",
			},
		},
		{
			title: "beacon node syncing mainnet with http enabled with port",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.TekuClient,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
					REST:                    true,
					RESTPort:                3333,
				},
			},
			result: []string{
				TekuDataPath,
				TekuNetwork,
				"mainnet",
				TekuExecutionEngineEndpoint,
				"https://localhost:8551",
				TekuJwtSecretFile,
				fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir())),
				TekuRestEnabled,
				TekuRestPort,
				"3333",
				TekuRESTAPICorsOrigins,
				"*",
				TekuRESTAPIHostAllowlist,
				"*",
			},
		},
		{
			title: "beacon node syncing mainnet with http enabled with port and host",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.TekuClient,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
					REST:                    true,
					RESTPort:                3333,
				},
			},
			result: []string{
				TekuDataPath,
				TekuNetwork,
				"mainnet",
				TekuExecutionEngineEndpoint,
				"https://localhost:8551",
				TekuJwtSecretFile,
				fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir())),
				TekuRestEnabled,
				TekuRestPort,
				"3333",
				TekuRestHost,
				"0.0.0.0",
				TekuRESTAPICorsOrigins,
				"*",
				TekuRESTAPIHostAllowlist,
				"*",
			},
		},
		{
			title: "beacon node syncing mainnet with p2p port, http enabled with port and host",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.TekuClient,
					P2PPort:                 7891,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
					REST:                    true,
					RESTPort:                3333,
				},
			},
			result: []string{
				TekuDataPath,
				TekuP2PPort,
				"7891",
				TekuNetwork,
				"mainnet",
				TekuExecutionEngineEndpoint,
				"https://localhost:8551",
				TekuJwtSecretFile,
				fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir())),
				TekuRestEnabled,
				TekuRestPort,
				"3333",
				TekuRestHost,
				"0.0.0.0",
				TekuRESTAPICorsOrigins,
				"*",
				TekuRESTAPIHostAllowlist,
				"*",
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
