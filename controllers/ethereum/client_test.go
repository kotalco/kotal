package controllers

import (
	"fmt"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum client arguments", func() {

	besuClient, _ := NewEthereumClient(ethereumv1alpha1.BesuClient)
	gethClient, _ := NewEthereumClient(ethereumv1alpha1.GethClient)
	parityClient, _ := NewEthereumClient(ethereumv1alpha1.ParityClient)
	coinbase := ethereumv1alpha1.EthereumAddress("0x2b3430337f12Ce89EaBC7b0d865F4253c7744c0d")
	accountKey := ethereumv1alpha1.PrivateKey("0x5df5eff7ef9e4e82739b68a34c6b23608d79ee8daf3b598a01ffb0dd7aa3a2fd")
	accountPassword := "secret"
	rinkeby := "rinkeby"
	nodekey := ethereumv1alpha1.PrivateKey("0x608e9b6f67c65e47531e08e8e501386dfae63a540fa3c48802c8aad854510b4e")

	cases := []struct {
		title   string
		network *ethereumv1alpha1.Network
		result  []string
	}{
		{
			"node joining rinkeby",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name: "node-1",
						},
					},
				},
			},
			[]string{
				BesuNatMethod,
				BesuNetwork,
				rinkeby,
				BesuDataPath,
				PathBlockchainData,
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.DefaultLogging),
			},
		},
		{
			"geth bootnode joining rinkeby",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.GethClient,
							Bootnode: true,
							Nodekey:  nodekey,
						},
					},
				},
			},
			[]string{
				"--rinkeby",
				GethNodeKey,
				GethDataDir,
				PathBlockchainData,
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.DefaultLogging),
				GethConfig,
			},
		},
		{
			"parity bootnode joining rinkeby",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.ParityClient,
							Bootnode: true,
							Nodekey:  nodekey,
						},
					},
				},
			},
			[]string{
				rinkeby,
				ParityNodeKey,
				ParityDataDir,
				PathBlockchainData,
				ParityLogging,
				parityClient.LoggingArgFromVerbosity(ethereumv1alpha1.DefaultLogging),
				ParityDisableRPC,
				ParityDisableWS,
			},
		},
		{
			"bootnode joining rinkeby",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Bootnode: true,
							Nodekey:  nodekey,
							Logging:  ethereumv1alpha1.NoLogs,
						},
					},
				},
			},
			[]string{
				BesuNatMethod,
				BesuNetwork,
				rinkeby,
				BesuNodePrivateKey,
				BesuDataPath,
				PathBlockchainData,
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.NoLogs),
			},
		},
		{
			"geth bootnode joining rinkeby",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.GethClient,
							Bootnode: true,
							Nodekey:  nodekey,
							Logging:  ethereumv1alpha1.AllLogs,
						},
					},
				},
			},
			[]string{
				"--rinkeby",
				GethNodeKey,
				GethDataDir,
				PathBlockchainData,
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.AllLogs),
				GethConfig,
			},
		},
		{
			"parity bootnode joining rinkeby",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.ParityClient,
							Bootnode: true,
							Nodekey:  nodekey,
							Logging:  ethereumv1alpha1.ErrorLogs,
						},
					},
				},
			},
			[]string{
				ParityNetwork,
				rinkeby,
				ParityDataDir,
				PathBlockchainData,
				ParityNodeKey,
				ParityLogging,
				parityClient.LoggingArgFromVerbosity(ethereumv1alpha1.ErrorLogs),
				ParityDisableRPC,
				ParityDisableWS,
			},
		},
		{
			"bootnode joining rinkeby with rpc",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Bootnode: true,
							Nodekey:  nodekey,
							RPC:      true,
							Logging:  ethereumv1alpha1.FatalLogs,
						},
					},
				},
			},
			[]string{
				BesuNatMethod,
				BesuNetwork,
				rinkeby,
				BesuNodePrivateKey,
				BesuDataPath,
				PathBlockchainData,
				BesuRPCHTTPEnabled,
				BesuRPCHTTPCorsOrigins,
				BesuHostAllowlist,
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.FatalLogs),
			},
		},
		{
			"geth bootnode joining rinkeby with rpc",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.GethClient,
							Bootnode: true,
							Nodekey:  nodekey,
							RPC:      true,
							Logging:  ethereumv1alpha1.WarnLogs,
						},
					},
				},
			},
			[]string{
				"--rinkeby",
				GethNodeKey,
				GethDataDir,
				PathBlockchainData,
				GethRPCHTTPEnabled,
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.WarnLogs),
				GethConfig,
				GethRPCHostWhitelist,
				GethRPCHTTPCorsOrigins,
			},
		},
		{
			"parity bootnode joining rinkeby with rpc",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.ParityClient,
							Bootnode: true,
							Nodekey:  nodekey,
							RPC:      true,
							Logging:  ethereumv1alpha1.WarnLogs,
						},
					},
				},
			},
			[]string{
				ParityNetwork,
				rinkeby,
				ParityDataDir,
				PathBlockchainData,
				ParityNodeKey,
				ParityLogging,
				parityClient.LoggingArgFromVerbosity(ethereumv1alpha1.WarnLogs),
				ParityDisableWS,
				ParityRPCHostWhitelist,
				ParityRPCHTTPCorsOrigins,
			},
		},
		{
			"bootnode joining rinkeby with rpc settings",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Bootnode: true,
							Nodekey:  nodekey,
							RPC:      true,
							RPCPort:  8599,
							RPCAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.ETHAPI,
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.NetworkAPI,
							},
							Logging: ethereumv1alpha1.ErrorLogs,
						},
					},
				},
			},
			[]string{
				BesuNatMethod,
				BesuNetwork,
				rinkeby,
				BesuNodePrivateKey,
				BesuDataPath,
				PathBlockchainData,
				BesuRPCHTTPEnabled,
				BesuRPCHTTPPort,
				"8599",
				BesuRPCHTTPAPI,
				"eth,web3,net",
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.ErrorLogs),
				BesuHostAllowlist,
				BesuRPCHTTPCorsOrigins,
			},
		},
		{
			"geth bootnode joining rinkeby with rpc settings",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.GethClient,
							Bootnode: true,
							Nodekey:  nodekey,
							RPC:      true,
							RPCPort:  8599,
							RPCAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.ETHAPI,
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.NetworkAPI,
							},
							Logging: ethereumv1alpha1.ErrorLogs,
						},
					},
				},
			},
			[]string{
				"--rinkeby",
				GethNodeKey,
				GethDataDir,
				PathBlockchainData,
				GethRPCHTTPEnabled,
				GethRPCHTTPPort,
				"8599",
				GethRPCHTTPAPI,
				"eth,web3,net",
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.ErrorLogs),
				GethConfig,
				GethRPCHostWhitelist,
				GethRPCHTTPCorsOrigins,
			},
		},
		{
			"parity bootnode joining rinkeby with rpc settings",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.ParityClient,
							Bootnode: true,
							Nodekey:  nodekey,
							RPC:      true,
							RPCPort:  8599,
							RPCAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.ETHAPI,
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.NetworkAPI,
							},
							Logging: ethereumv1alpha1.DebugLogs,
						},
					},
				},
			},
			[]string{
				ParityNetwork,
				rinkeby,
				ParityNodeKey,
				ParityDataDir,
				PathBlockchainData,
				ParityRPCHTTPPort,
				"8599",
				ParityRPCHTTPAPI,
				"eth,web3,net",
				ParityLogging,
				parityClient.LoggingArgFromVerbosity(ethereumv1alpha1.DebugLogs),
				ParityDisableWS,
				ParityRPCHostWhitelist,
				ParityRPCHTTPCorsOrigins,
			},
		},
		{
			"bootnode joining rinkeby with rpc, ws settings",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Bootnode: true,
							Nodekey:  nodekey,
							RPC:      true,
							RPCPort:  8599,
							RPCAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.ETHAPI,
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.NetworkAPI,
							},
							WS:     true,
							WSPort: 8588,
							WSAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.ETHAPI,
							},
							Logging: ethereumv1alpha1.WarnLogs,
						},
					},
				},
			},
			[]string{
				BesuNatMethod,
				BesuNetwork,
				rinkeby,
				BesuNodePrivateKey,
				BesuDataPath,
				PathBlockchainData,
				BesuRPCHTTPEnabled,
				BesuRPCHTTPPort,
				"8599",
				BesuRPCHTTPHost,
				DefaultHost,
				BesuRPCHTTPAPI,
				"eth,web3,net",
				BesuRPCWSEnabled,
				BesuRPCWSHost,
				DefaultHost,
				BesuRPCWSPort,
				"8588",
				BesuRPCWSAPI,
				"web3,eth",
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.WarnLogs),
				BesuHostAllowlist,
				BesuRPCHTTPCorsOrigins,
			},
		},
		{
			"geth bootnode joining rinkeby with rpc, ws settings",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.GethClient,
							Bootnode: true,
							Nodekey:  nodekey,
							RPC:      true,
							RPCPort:  8599,
							RPCAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.ETHAPI,
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.NetworkAPI,
							},
							WS:     true,
							WSPort: 8588,
							WSAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.ETHAPI,
							},
							Logging: ethereumv1alpha1.WarnLogs,
						},
					},
				},
			},
			[]string{
				"--rinkeby",
				GethNodeKey,
				GethDataDir,
				PathBlockchainData,
				GethRPCHTTPEnabled,
				GethRPCHTTPPort,
				"8599",
				GethRPCHTTPHost,
				DefaultHost,
				GethRPCHTTPAPI,
				"eth,web3,net",
				GethRPCWSEnabled,
				GethRPCWSHost,
				DefaultHost,
				GethRPCWSPort,
				"8588",
				GethRPCWSAPI,
				"web3,eth",
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.WarnLogs),
				GethConfig,
				GethRPCHostWhitelist,
				GethRPCHTTPCorsOrigins,
				GethWSOrigins,
			},
		},
		{
			"parity bootnode joining rinkeby with rpc, ws settings",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.ParityClient,
							Bootnode: true,
							Nodekey:  nodekey,
							RPC:      true,
							RPCPort:  8599,
							RPCAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.ETHAPI,
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.NetworkAPI,
							},
							WS:     true,
							WSPort: 8588,
							WSAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.ETHAPI,
							},
							Logging: ethereumv1alpha1.TraceLogs,
						},
					},
				},
			},
			[]string{
				ParityNetwork,
				rinkeby,
				ParityNodeKey,
				ParityDataDir,
				PathBlockchainData,
				ParityRPCHTTPPort,
				"8599",
				ParityRPCHTTPHost,
				DefaultHost,
				ParityRPCHTTPAPI,
				"eth,web3,net",
				ParityRPCWSHost,
				DefaultHost,
				ParityRPCWSPort,
				"8588",
				ParityRPCWSAPI,
				"web3,eth",
				ParityLogging,
				parityClient.LoggingArgFromVerbosity(ethereumv1alpha1.TraceLogs),
				ParityRPCHostWhitelist,
				ParityRPCHTTPCorsOrigins,
				ParityRPCWSWhitelist,
				ParityRPCWSCorsOrigins,
			},
		},
		{
			"bootnode joining rinkeby with rpc, ws, graphql settings and cors domains",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Bootnode: true,
							Nodekey:  nodekey,
							RPC:      true,
							RPCPort:  8599,
							RPCAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.ETHAPI,
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.NetworkAPI,
							},
							CORSDomains: []string{"cors.example.com"},
							WS:          true,
							WSPort:      8588,
							WSAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.ETHAPI,
							},
							GraphQL:     true,
							GraphQLPort: 8511,
							Logging:     ethereumv1alpha1.InfoLogs,
						},
					},
				},
			},
			[]string{
				BesuNatMethod,
				BesuNetwork,
				rinkeby,
				BesuNodePrivateKey,
				BesuDataPath,
				PathBlockchainData,
				BesuRPCHTTPCorsOrigins,
				BesuRPCHTTPEnabled,
				BesuRPCHTTPPort,
				"8599",
				BesuRPCHTTPHost,
				DefaultHost,
				BesuRPCHTTPAPI,
				"eth,web3,net",
				BesuRPCWSEnabled,
				BesuRPCWSHost,
				DefaultHost,
				BesuRPCWSPort,
				"8588",
				BesuRPCWSAPI,
				"web3,eth",
				BesuGraphQLHTTPEnabled,
				BesuGraphQLHTTPHost,
				DefaultHost,
				BesuGraphQLHTTPPort,
				"8511",
				BesuGraphQLHTTPCorsOrigins,
				"cors.example.com",
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.InfoLogs),
				BesuHostAllowlist,
			},
		},
		{
			"geth bootnode joining rinkeby with rpc, ws, graphql settings and cors domains",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.GethClient,
							Bootnode: true,
							Nodekey:  nodekey,
							RPC:      true,
							RPCPort:  8599,
							RPCAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.ETHAPI,
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.NetworkAPI,
							},
							CORSDomains: []string{"cors.example.com"},
							WS:          true,
							WSPort:      8588,
							WSAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.ETHAPI,
							},
							GraphQL:     true,
							GraphQLPort: 8511,
							Logging:     ethereumv1alpha1.InfoLogs,
						},
					},
				},
			},
			[]string{
				"--rinkeby",
				GethNodeKey,
				GethDataDir,
				PathBlockchainData,
				GethRPCHTTPCorsOrigins,
				GethRPCHTTPEnabled,
				GethRPCHTTPPort,
				"8599",
				GethRPCHTTPHost,
				DefaultHost,
				GethRPCHTTPAPI,
				"eth,web3,net",
				GethRPCWSEnabled,
				GethRPCWSHost,
				DefaultHost,
				GethRPCWSPort,
				"8588",
				GethRPCWSAPI,
				"web3,eth",
				GethGraphQLHTTPEnabled,
				GethGraphQLHTTPCorsOrigins,
				"cors.example.com",
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.InfoLogs),
				GethConfig,
				GethRPCHostWhitelist,
				GethGraphQLHostWhitelist,
				GethWSOrigins,
			},
		},
		{
			"miner node of private network that connects to bootnode",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					ID:        8888,
					Consensus: ethereumv1alpha1.ProofOfAuthority,
					Genesis: &ethereumv1alpha1.Genesis{
						ChainID: 5555,
					},
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Miner:    true,
							Coinbase: coinbase,
							Logging:  ethereumv1alpha1.DebugLogs,
						},
					},
				},
			},
			[]string{
				BesuNatMethod,
				BesuNetworkID,
				"8888",
				BesuDataPath,
				PathBlockchainData,
				BesuMinerEnabled,
				BesuMinerCoinbase,
				string(coinbase),
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.DebugLogs),
				BesuDiscoveryEnabled,
				"false",
			},
		},
		{
			"geth miner node of private network that connects to bootnode",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					ID:        7777,
					Consensus: ethereumv1alpha1.ProofOfAuthority,
					Genesis: &ethereumv1alpha1.Genesis{
						ChainID: 5555,
					},
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.GethClient,
							Miner:    true,
							Coinbase: coinbase,
							Import: &ethereumv1alpha1.ImportedAccount{
								PrivateKey: accountKey,
								Password:   accountPassword,
							},
							Logging: ethereumv1alpha1.DebugLogs,
						},
					},
				},
			},
			[]string{
				GethDataDir,
				GethNetworkID,
				"7777",
				PathBlockchainData,
				GethMinerEnabled,
				GethMinerCoinbase,
				GethUnlock,
				GethPassword,
				string(coinbase),
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.DebugLogs),
				GethNoDiscovery,
				GethConfig,
			},
		},
		{
			"parity node of private network that connects to bootnode",
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					ID:        8888,
					Consensus: ethereumv1alpha1.ProofOfAuthority,
					Genesis: &ethereumv1alpha1.Genesis{
						ChainID: 5555,
					},
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.ParityClient,
							Miner:    true,
							Coinbase: coinbase,
							Logging:  ethereumv1alpha1.InfoLogs,
						},
					},
				},
			},
			[]string{
				ParityNetworkID,
				"8888",
				ParityDataDir,
				PathBlockchainData,
				ParityMinerCoinbase,
				string(coinbase),
				ParityLogging,
				ParityUnlock,
				ParityPassword,
				ParityEngineSigner,
				parityClient.LoggingArgFromVerbosity(ethereumv1alpha1.InfoLogs),
				ParityDisableRPC,
				ParityDisableWS,
				ParityNoDiscovery,
			},
		},
	}

	for _, c := range cases {
		func() {
			cc := c
			It(fmt.Sprintf("Should create correct client arguments for %s", cc.title), func() {
				cc.network.Default()
				client, err := NewEthereumClient(cc.network.Spec.Nodes[0].Client)
				Expect(err).To(BeNil())
				args := client.GetArgs(&cc.network.Spec.Nodes[0], cc.network)
				Expect(args).To(ContainElements(cc.result))
			})
		}()
	}

})
