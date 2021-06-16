package controllers

import (
	"fmt"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum client arguments", func() {

	coinbase := ethereumv1alpha1.EthereumAddress("0x2b3430337f12Ce89EaBC7b0d865F4253c7744c0d")
	rinkeby := "rinkeby"
	bootnode := "enode://6f8a80d14311c39f35f516fa664deaaaa13e85b2f7493f37f6144d86991ec012937307647bd3b9a82abe2974e1407241d54947bbb39763a4cac9f77166ad92a0@10.3.58.6:30303"
	bootnodes := []ethereumv1alpha1.Enode{ethereumv1alpha1.Enode(bootnode)}

	cases := []struct {
		title  string
		node   *ethereumv1alpha1.Node
		result []string
	}{
		{
			"node joining rinkeby",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					Client: ethereumv1alpha1.BesuClient,
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Bootnodes: bootnodes,
				},
			},
			[]string{
				BesuNatMethod,
				BesuNetwork,
				rinkeby,
				BesuDataPath,
				BesuLogging,
				BesuBootnodes,
				bootnode,
			},
		},
		{
			"geth node joining rinkeby",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.GethClient,
					NodekeySecretName: "nodekey",
					Bootnodes:         bootnodes,
				},
			},
			[]string{
				"--rinkeby",
				GethNodeKey,
				GethDataDir,
				GethLogging,
				GethConfig,
				GethBootnodes,
				bootnode,
			},
		},
		{
			"parity joining rinkeby",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.ParityClient,
					NodekeySecretName: "nodekey",
					Bootnodes:         bootnodes,
				},
			},
			[]string{
				rinkeby,
				ParityNodeKey,
				ParityDataDir,
				ParityLogging,
				ParityDisableRPC,
				ParityDisableWS,
				ParityBootnodes,
				bootnode,
			},
		},
		{
			"besu node joining rinkeby",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					Client: ethereumv1alpha1.BesuClient,
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					NodekeySecretName: "nodekey",
					Logging:           ethereumv1alpha1.NoLogs,
				},
			},
			[]string{
				BesuNatMethod,
				BesuNetwork,
				rinkeby,
				BesuNodePrivateKey,
				BesuDataPath,
				BesuLogging,
			},
		},
		{
			"geth node joining rinkeby",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.GethClient,
					NodekeySecretName: "nodekey",
					Logging:           ethereumv1alpha1.AllLogs,
				},
			},
			[]string{
				"--rinkeby",
				GethNodeKey,
				GethDataDir,
				GethLogging,
				GethConfig,
			},
		},
		{
			"parity joining rinkeby",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.ParityClient,
					NodekeySecretName: "nodekey",
					Logging:           ethereumv1alpha1.ErrorLogs,
				},
			},
			[]string{
				ParityNetwork,
				rinkeby,
				ParityDataDir,
				ParityNodeKey,
				ParityLogging,
				ParityDisableRPC,
				ParityDisableWS,
			},
		},
		{
			"besu node joining rinkeby with rpc",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					Client: ethereumv1alpha1.BesuClient,
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					NodekeySecretName: "nodekey",
					RPC:               true,
					Logging:           ethereumv1alpha1.FatalLogs,
				},
			},
			[]string{
				BesuNatMethod,
				BesuNetwork,
				rinkeby,
				BesuNodePrivateKey,
				BesuDataPath,
				BesuRPCHTTPEnabled,
				BesuRPCHTTPCorsOrigins,
				BesuHostAllowlist,
				BesuLogging,
			},
		},
		{
			"geth node joining rinkeby with rpc",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.GethClient,
					NodekeySecretName: "nodekey",
					RPC:               true,
					Logging:           ethereumv1alpha1.WarnLogs,
				},
			},
			[]string{
				"--rinkeby",
				GethNodeKey,
				GethDataDir,
				GethRPCHTTPEnabled,
				GethLogging,
				GethConfig,
				GethRPCHostWhitelist,
				GethRPCHTTPCorsOrigins,
			},
		},
		{
			"parity joining rinkeby with rpc",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.ParityClient,
					NodekeySecretName: "nodekey",
					RPC:               true,
					Logging:           ethereumv1alpha1.WarnLogs,
				},
			},
			[]string{
				ParityNetwork,
				rinkeby,
				ParityDataDir,
				ParityNodeKey,
				ParityLogging,
				ParityDisableWS,
				ParityRPCHostWhitelist,
				ParityRPCHTTPCorsOrigins,
			},
		},
		{
			"besu node joining rinkeby with rpc settings",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.BesuClient,
					NodekeySecretName: "nodekey",
					RPC:               true,
					RPCPort:           8599,
					RPCAPI: []ethereumv1alpha1.API{
						ethereumv1alpha1.ETHAPI,
						ethereumv1alpha1.Web3API,
						ethereumv1alpha1.NetworkAPI,
					},
					Logging: ethereumv1alpha1.ErrorLogs,
				},
			},
			[]string{
				BesuNatMethod,
				BesuNetwork,
				rinkeby,
				BesuNodePrivateKey,
				BesuDataPath,
				BesuRPCHTTPEnabled,
				BesuRPCHTTPPort,
				"8599",
				BesuRPCHTTPAPI,
				"eth,web3,net",
				BesuLogging,
				BesuHostAllowlist,
				BesuRPCHTTPCorsOrigins,
			},
		},
		{
			"geth node joining rinkeby with rpc settings",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.GethClient,
					NodekeySecretName: "nodekey",
					RPC:               true,
					RPCPort:           8599,
					RPCAPI: []ethereumv1alpha1.API{
						ethereumv1alpha1.ETHAPI,
						ethereumv1alpha1.Web3API,
						ethereumv1alpha1.NetworkAPI,
					},
					Logging: ethereumv1alpha1.ErrorLogs,
				},
			},
			[]string{
				"--rinkeby",
				GethNodeKey,
				GethDataDir,
				GethRPCHTTPEnabled,
				GethRPCHTTPPort,
				"8599",
				GethRPCHTTPAPI,
				"eth,web3,net",
				GethLogging,
				GethConfig,
				GethRPCHostWhitelist,
				GethRPCHTTPCorsOrigins,
			},
		},
		{
			"parity joining rinkeby with rpc settings",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.ParityClient,
					NodekeySecretName: "nodekey",
					RPC:               true,
					RPCPort:           8599,
					RPCAPI: []ethereumv1alpha1.API{
						ethereumv1alpha1.ETHAPI,
						ethereumv1alpha1.Web3API,
						ethereumv1alpha1.NetworkAPI,
					},
					Logging: ethereumv1alpha1.DebugLogs,
				},
			},
			[]string{
				ParityNetwork,
				rinkeby,
				ParityNodeKey,
				ParityDataDir,
				ParityRPCHTTPPort,
				"8599",
				ParityRPCHTTPAPI,
				"eth,web3,net",
				ParityLogging,
				ParityDisableWS,
				ParityRPCHostWhitelist,
				ParityRPCHTTPCorsOrigins,
			},
		},
		{
			"besu node joining rinkeby with rpc, ws settings",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.BesuClient,
					NodekeySecretName: "nodekey",
					RPC:               true,
					RPCPort:           8599,
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
			[]string{
				BesuNatMethod,
				BesuNetwork,
				rinkeby,
				BesuNodePrivateKey,
				BesuDataPath,
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
				BesuHostAllowlist,
				BesuRPCHTTPCorsOrigins,
			},
		},
		{
			"geth node joining rinkeby with rpc, ws settings",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.GethClient,
					NodekeySecretName: "nodekey",
					RPC:               true,
					RPCPort:           8599,
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
			[]string{
				"--rinkeby",
				GethNodeKey,
				GethDataDir,
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
				GethConfig,
				GethRPCHostWhitelist,
				GethRPCHTTPCorsOrigins,
				GethWSOrigins,
			},
		},
		{
			"parity joining rinkeby with rpc, ws settings",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.ParityClient,
					NodekeySecretName: "nodekey",
					RPC:               true,
					RPCPort:           8599,
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
			[]string{
				ParityNetwork,
				rinkeby,
				ParityNodeKey,
				ParityDataDir,
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
				ParityRPCHostWhitelist,
				ParityRPCHTTPCorsOrigins,
				ParityRPCWSWhitelist,
				ParityRPCWSCorsOrigins,
			},
		},
		{
			"besu node joining rinkeby with rpc, ws, graphql settings and cors domains",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.BesuClient,
					NodekeySecretName: "nodekey",
					RPC:               true,
					RPCPort:           8599,
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
			[]string{
				BesuNatMethod,
				BesuNetwork,
				rinkeby,
				BesuNodePrivateKey,
				BesuDataPath,
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
				BesuHostAllowlist,
			},
		},
		{
			"geth node joining rinkeby with rpc, ws, graphql settings and cors domains",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						Join: rinkeby,
					},
					Client:            ethereumv1alpha1.GethClient,
					NodekeySecretName: "nodekey",
					RPC:               true,
					RPCPort:           8599,
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
			[]string{
				"--rinkeby",
				GethNodeKey,
				GethDataDir,
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
				GethConfig,
				GethRPCHostWhitelist,
				GethGraphQLHostWhitelist,
				GethWSOrigins,
			},
		},
		{
			"miner node of private network that connects to bootnode",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						ID:        8888,
						Consensus: ethereumv1alpha1.ProofOfAuthority,
						Genesis: &ethereumv1alpha1.Genesis{
							ChainID: 5555,
						},
					},
					Client:   ethereumv1alpha1.BesuClient,
					Miner:    true,
					Coinbase: coinbase,
					Logging:  ethereumv1alpha1.DebugLogs,
				},
			},
			[]string{
				BesuNatMethod,
				BesuNetworkID,
				"8888",
				BesuDataPath,
				BesuMinerEnabled,
				BesuMinerCoinbase,
				string(coinbase),
				BesuLogging,
				BesuDiscoveryEnabled,
				"false",
			},
		},
		{
			"geth miner node of private network that connects to bootnode",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						ID:        7777,
						Consensus: ethereumv1alpha1.ProofOfAuthority,
						Genesis: &ethereumv1alpha1.Genesis{
							ChainID: 5555,
						},
					},
					Client:   ethereumv1alpha1.GethClient,
					Miner:    true,
					Coinbase: coinbase,
					Import: &ethereumv1alpha1.ImportedAccount{
						PrivateKeySecretName: "my-account-privatekey",
						PasswordSecretName:   "my-account-password",
					},
					Logging: ethereumv1alpha1.DebugLogs,
				},
			},
			[]string{
				GethDataDir,
				GethNetworkID,
				"7777",
				GethMinerEnabled,
				GethMinerCoinbase,
				GethUnlock,
				GethPassword,
				string(coinbase),
				GethLogging,
				GethNoDiscovery,
				GethConfig,
			},
		},
		{
			"parity node of private network that connects to bootnode",
			&ethereumv1alpha1.Node{
				Spec: ethereumv1alpha1.NodeSpec{
					NetworkConfig: ethereumv1alpha1.NetworkConfig{
						ID:        8888,
						Consensus: ethereumv1alpha1.ProofOfAuthority,
						Genesis: &ethereumv1alpha1.Genesis{
							ChainID: 5555,
						},
					},
					Client:   ethereumv1alpha1.ParityClient,
					Miner:    true,
					Coinbase: coinbase,
					Logging:  ethereumv1alpha1.InfoLogs,
				},
			},
			[]string{
				ParityNetworkID,
				"8888",
				ParityDataDir,
				ParityMinerCoinbase,
				string(coinbase),
				ParityLogging,
				ParityUnlock,
				ParityPassword,
				ParityEngineSigner,
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
				cc.node.Default()
				client, err := NewEthereumClient(cc.node)
				Expect(err).To(BeNil())
				args := client.Args()
				Expect(args).To(ContainElements(cc.result))
			})
		}()
	}

})
