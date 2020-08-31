package controllers

import (
	"fmt"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum client arguments", func() {

	var bootnodes []string
	besuClient, _ := NewEthereumClient(ethereumv1alpha1.BesuClient)
	gethClient, _ := NewEthereumClient(ethereumv1alpha1.GethClient)
	rinkeby := "rinkeby"
	bootnode := "enode://publickey@ip:port"
	coinbase := ethereumv1alpha1.EthereumAddress("0x5A0b54D5dc17e0AadC383d2db43B0a0D3E029c4c")
	nodekey := ethereumv1alpha1.PrivateKey("0x608e9b6f67c65e47531e08e8e501386dfae63a540fa3c48802c8aad854510b4e")

	cases := []struct {
		title     string
		bootnodes []string
		network   *ethereumv1alpha1.Network
		result    []string
	}{
		{
			"node joining rinkeby",
			bootnodes,
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
			bootnodes,
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
			},
		},
		{
			"bootnode joining rinkeby",
			bootnodes,
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
			bootnodes,
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.GethClient,
							Bootnode: true,
							Nodekey:  nodekey,
							Logging:  ethereumv1alpha1.NoLogs,
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
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.NoLogs),
			},
		},
		{
			"bootnode joining rinkeby with rpc",
			bootnodes,
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
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.FatalLogs),
			},
		},
		{
			"geth bootnode joining rinkeby with rpc",
			bootnodes,
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
							Logging:  ethereumv1alpha1.FatalLogs,
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
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.FatalLogs),
			},
		},
		{
			"bootnode joining rinkeby with rpc settings",
			bootnodes,
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
			},
		},
		{
			"geth bootnode joining rinkeby with rpc settings",
			bootnodes,
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
			},
		},
		{
			"bootnode joining rinkeby with rpc, ws settings",
			bootnodes,
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Bootnode: true,
							Nodekey:  nodekey,
							RPC:      true,
							RPCHost:  "0.0.0.0",
							RPCPort:  8599,
							RPCAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.ETHAPI,
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.NetworkAPI,
							},
							WS:     true,
							WSHost: "127.0.0.1",
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
				"0.0.0.0",
				BesuRPCHTTPAPI,
				"eth,web3,net",
				BesuRPCWSEnabled,
				BesuRPCWSHost,
				"127.0.0.1",
				BesuRPCWSPort,
				"8588",
				BesuRPCWSAPI,
				"web3,eth",
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.WarnLogs),
			},
		},
		{
			"geth bootnode joining rinkeby with rpc, ws settings",
			bootnodes,
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
							RPCHost:  "0.0.0.0",
							RPCPort:  8599,
							RPCAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.ETHAPI,
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.NetworkAPI,
							},
							WS:     true,
							WSHost: "127.0.0.1",
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
				"0.0.0.0",
				GethRPCHTTPAPI,
				"eth,web3,net",
				GethRPCWSEnabled,
				GethRPCWSHost,
				"127.0.0.1",
				GethRPCWSPort,
				"8588",
				GethRPCWSAPI,
				"web3,eth",
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.WarnLogs),
			},
		},
		{
			"bootnode joining rinkeby with rpc, ws, graphql settings and cors domains",
			bootnodes,
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					Join: rinkeby,
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Bootnode: true,
							Nodekey:  nodekey,
							RPC:      true,
							RPCHost:  "0.0.0.0",
							RPCPort:  8599,
							RPCAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.ETHAPI,
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.NetworkAPI,
							},
							CORSDomains: []string{"cors.example.com"},
							WS:          true,
							WSHost:      "127.0.0.1",
							WSPort:      8588,
							WSAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.ETHAPI,
							},
							GraphQL:     true,
							GraphQLHost: "127.0.0.2",
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
				"0.0.0.0",
				BesuRPCHTTPAPI,
				"eth,web3,net",
				BesuRPCWSEnabled,
				BesuRPCWSHost,
				"127.0.0.1",
				BesuRPCWSPort,
				"8588",
				BesuRPCWSAPI,
				"web3,eth",
				BesuGraphQLHTTPEnabled,
				BesuGraphQLHTTPHost,
				"127.0.0.2",
				BesuGraphQLHTTPPort,
				"8511",
				BesuGraphQLHTTPCorsOrigins,
				"cors.example.com",
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.InfoLogs),
			},
		},
		{
			"geth bootnode joining rinkeby with rpc, ws, graphql settings and cors domains",
			bootnodes,
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
							RPCHost:  "0.0.0.0",
							RPCPort:  8599,
							RPCAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.ETHAPI,
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.NetworkAPI,
							},
							CORSDomains: []string{"cors.example.com"},
							WS:          true,
							WSHost:      "127.0.0.1",
							WSPort:      8588,
							WSAPI: []ethereumv1alpha1.API{
								ethereumv1alpha1.Web3API,
								ethereumv1alpha1.ETHAPI,
							},
							GraphQL:     true,
							GraphQLHost: "127.0.0.2",
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
				"0.0.0.0",
				GethRPCHTTPAPI,
				"eth,web3,net",
				GethRPCWSEnabled,
				GethRPCWSHost,
				"127.0.0.1",
				GethRPCWSPort,
				"8588",
				GethRPCWSAPI,
				"web3,eth",
				GethGraphQLHTTPEnabled,
				GethGraphQLHTTPHost,
				"127.0.0.2",
				GethGraphQLHTTPPort,
				"8511",
				GethGraphQLHTTPCorsOrigins,
				"cors.example.com",
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.InfoLogs),
			},
		},
		{
			"miner node of private network that connects to bootnode",
			[]string{bootnode},
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					ID:      8888,
					Genesis: &ethereumv1alpha1.Genesis{},
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
				BesuBootnodes,
				bootnode,
				BesuMinerEnabled,
				BesuMinerCoinbase,
				string(coinbase),
				BesuLogging,
				besuClient.LoggingArgFromVerbosity(ethereumv1alpha1.DebugLogs),
			},
		},
		{
			"geth miner node of private network that connects to bootnode",
			[]string{bootnode},
			&ethereumv1alpha1.Network{
				Spec: ethereumv1alpha1.NetworkSpec{
					ID:      7777,
					Genesis: &ethereumv1alpha1.Genesis{},
					Nodes: []ethereumv1alpha1.Node{
						{
							Name:     "node-1",
							Client:   ethereumv1alpha1.GethClient,
							Miner:    true,
							Coinbase: coinbase,
							Logging:  ethereumv1alpha1.DebugLogs,
						},
					},
				},
			},
			[]string{
				GethDataDir,
				GethNetworkID,
				"7777",
				PathBlockchainData,
				GethBootnodes,
				bootnode,
				GethMinerEnabled,
				GethMinerCoinbase,
				string(coinbase),
				GethLogging,
				gethClient.LoggingArgFromVerbosity(ethereumv1alpha1.DebugLogs),
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
				args := client.GetArgs(&cc.network.Spec.Nodes[0], cc.network, cc.bootnodes)
				Expect(args).To(ContainElements(cc.result))
			})
		}()
	}

})
