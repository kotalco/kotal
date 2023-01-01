package ethereum2

import (
	"fmt"
	"os"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
	"github.com/kotalco/kotal/controllers/shared"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lighthouse beacon node", func() {

	node := ethereum2v1alpha1.BeaconNode{
		Spec: ethereum2v1alpha1.BeaconNodeSpec{
			Client:  ethereum2v1alpha1.LighthouseClient,
			Network: "mainnet",
		},
	}
	client, _ := NewClient(&node)

	It("Should get correct image", func() {
		// default image
		img := client.Image()
		Expect(img).To(Equal(DefaultLighthouseBeaconNodeImage))
		// after changing .spec.image
		testImage := "kotalco/lighthouse:spec"
		node.Spec.Image = &testImage
		img = client.Image()
		Expect(img).To(Equal(testImage))
		// after setting custom image
		testImage = "kotalco/lighthouse:test"
		os.Setenv(EnvLighthouseBeaconNodeImage, testImage)
		img = client.Image()
		Expect(img).To(Equal(testImage))
	})

	It("Should get correct command", func() {
		Expect(client.Command()).To(ConsistOf("lighthouse", "bn"))
	})

	It("Should get correct env", func() {
		Expect(client.Env()).To(BeNil())
	})

	It("Should get correct home dir", func() {
		Expect(client.HomeDir()).To(Equal(LighthouseHomeDir))
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
					Client:  ethereum2v1alpha1.LighthouseClient,
					Network: "mainnet",
					Logging: sharedAPI.TraceLogs,
				},
			},
			result: []string{
				LighthouseDataDir,
				LighthouseNetwork,
				"mainnet",
				LighthouseDebugLevel,
				string(sharedAPI.TraceLogs),
			},
		},
		{
			title: "beacon node syncing mainnet",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.LighthouseClient,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
				},
			},
			result: []string{
				LighthouseDataDir,
				LighthouseNetwork,
				"mainnet",
				LighthouseExecutionEngineEndpoint,
				"https://localhost:8551",
				LighthouseJwtSecretFile,
				fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir())),
			},
		},
		{
			title: "beacon node syncing mainnet and http enabled",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.LighthouseClient,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
					REST:                    true,
				},
			},
			result: []string{
				LighthouseDataDir,
				LighthouseNetwork,
				"mainnet",
				LighthouseExecutionEngineEndpoint,
				"https://localhost:8551",
				LighthouseJwtSecretFile,
				fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir())),
				LighthouseHTTP,
				LighthouseAllowOrigins,
				"*",
			},
		},
		{
			title: "beacon node syncing mainnet and http enabled with port",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.LighthouseClient,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
					REST:                    true,
					RESTPort:                4444,
				},
			},
			result: []string{
				LighthouseDataDir,
				LighthouseNetwork,
				"mainnet",
				LighthouseExecutionEngineEndpoint,
				"https://localhost:8551",
				LighthouseJwtSecretFile,
				fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir())),
				LighthouseHTTP,
				LighthouseHTTPPort,
				"4444",
				LighthouseAllowOrigins,
				"*",
			},
		},
		{
			title: "beacon node syncing mainnet with http enabled with port and host",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.LighthouseClient,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
					REST:                    true,
					RESTPort:                4444,
					RESTHost:                "0.0.0.0",
				},
			},
			result: []string{
				LighthouseDataDir,
				LighthouseNetwork,
				"mainnet",
				LighthouseExecutionEngineEndpoint,
				"https://localhost:8551",
				LighthouseJwtSecretFile,
				fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir())),
				LighthouseHTTP,
				LighthouseHTTPPort,
				"4444",
				LighthouseHTTPAddress,
				"0.0.0.0",
				LighthouseAllowOrigins,
				"*",
			},
		},
		{
			title: "beacon node syncing mainnet with p2p port, http enabled with port and host",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:                  ethereum2v1alpha1.LighthouseClient,
					P2PPort:                 7891,
					Network:                 "mainnet",
					ExecutionEngineEndpoint: "https://localhost:8551",
					JWTSecretName:           "jwt-secret",
					REST:                    true,
					RESTPort:                4444,
					RESTHost:                "0.0.0.0",
				},
			},
			result: []string{
				LighthouseDataDir,
				LighthousePort,
				"7891",
				LighthouseDiscoveryPort,
				"7891",
				LighthouseNetwork,
				"mainnet",
				LighthouseExecutionEngineEndpoint,
				"https://localhost:8551",
				LighthouseJwtSecretFile,
				fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir())),
				LighthouseHTTP,
				LighthouseHTTPPort,
				"4444",
				LighthouseHTTPAddress,
				"0.0.0.0",
				LighthouseAllowOrigins,
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
