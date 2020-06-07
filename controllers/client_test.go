package controllers

import (
	"fmt"

	ethereumv1alpha1 "github.com/mfarghaly/kotal/api/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ethereum client arguments", func() {

	// var noNetwork string
	var genesis bool
	var bootnodes []string
	rinkeby := "rinkeby"
	bootnode := "enode://publickey@ip:port"
	coinbase := ethereumv1alpha1.EthereumAddress("0x5A0b54D5dc17e0AadC383d2db43B0a0D3E029c4c")
	nodekey := ethereumv1alpha1.PrivateKey("0x608e9b6f67c65e47531e08e8e501386dfae63a540fa3c48802c8aad854510b4e")

	cases := []struct {
		title     string
		join      string
		genesis   bool
		bootnodes []string
		node      *ethereumv1alpha1.Node
		result    []string
	}{
		{
			"node joining rinkeby",
			rinkeby,
			genesis,
			bootnodes,
			&ethereumv1alpha1.Node{},
			[]string{
				ArgNatMethod,
				ArgNetwork,
				rinkeby,
				ArgDataPath,
				blockchainDataPath,
			},
		},
		{
			"bootnode joining rinkeby",
			rinkeby,
			genesis,
			bootnodes,
			&ethereumv1alpha1.Node{
				Bootnode: true,
				Nodekey:  nodekey,
			},
			[]string{
				ArgNatMethod,
				ArgNetwork,
				rinkeby,
				ArgNodePrivateKey,
				ArgDataPath,
				blockchainDataPath,
			},
		},
		{
			"bootnode joining rinkeby with rpc",
			rinkeby,
			genesis,
			bootnodes,
			&ethereumv1alpha1.Node{
				Bootnode: true,
				Nodekey:  nodekey,
				RPC:      true,
			},
			[]string{
				ArgNatMethod,
				ArgNetwork,
				rinkeby,
				ArgNodePrivateKey,
				ArgDataPath,
				blockchainDataPath,
				ArgRPCHTTPEnabled,
			},
		},
		{
			"bootnode joining rinkeby with rpc settings",
			rinkeby,
			genesis,
			bootnodes,
			&ethereumv1alpha1.Node{
				Bootnode: true,
				Nodekey:  nodekey,
				RPC:      true,
				RPCPort:  8599,
				RPCAPI: []ethereumv1alpha1.API{
					ethereumv1alpha1.ETHAPI,
					ethereumv1alpha1.Web3API,
					ethereumv1alpha1.NetworkAPI,
				},
			},
			[]string{
				ArgNatMethod,
				ArgNetwork,
				rinkeby,
				ArgNodePrivateKey,
				ArgDataPath,
				blockchainDataPath,
				ArgRPCHTTPEnabled,
				ArgRPCHTTPPort,
				"8599",
				ArgRPCHTTPAPI,
				"eth,web3,net",
			},
		},
		{
			"bootnode joining rinkeby with rpc, ws settings",
			rinkeby,
			genesis,
			bootnodes,
			&ethereumv1alpha1.Node{
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
			},
			[]string{
				ArgNatMethod,
				ArgNetwork,
				rinkeby,
				ArgNodePrivateKey,
				ArgDataPath,
				blockchainDataPath,
				ArgRPCHTTPEnabled,
				ArgRPCHTTPPort,
				"8599",
				ArgRPCHTTPHost,
				"0.0.0.0",
				ArgRPCHTTPAPI,
				"eth,web3,net",
				ArgRPCWSEnabled,
				ArgRPCWSHost,
				"127.0.0.1",
				ArgRPCWSPort,
				"8588",
				ArgRPCWSAPI,
				"web3,eth",
			},
		},
		{
			"bootnode joining rinkeby with rpc, ws, graphql settings and cors domains",
			rinkeby,
			genesis,
			bootnodes,
			&ethereumv1alpha1.Node{
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
			},
			[]string{
				ArgNatMethod,
				ArgNetwork,
				rinkeby,
				ArgNodePrivateKey,
				ArgDataPath,
				blockchainDataPath,
				ArgRPCHTTPCorsOrigins,
				ArgRPCHTTPEnabled,
				ArgRPCHTTPPort,
				"8599",
				ArgRPCHTTPHost,
				"0.0.0.0",
				ArgRPCHTTPAPI,
				"eth,web3,net",
				ArgRPCWSEnabled,
				ArgRPCWSHost,
				"127.0.0.1",
				ArgRPCWSPort,
				"8588",
				ArgRPCWSAPI,
				"web3,eth",
				ArgGraphQLHTTPEnabled,
				ArgGraphQLHTTPHost,
				"127.0.0.2",
				ArgGraphQLHTTPPort,
				"8511",
				ArgGraphQLHTTPCorsOrigins,
				"cors.example.com",
			},
		},
		{
			"miner node of private network that connects to bootnode",
			"",   // no network
			true, // genesis
			[]string{bootnode},
			&ethereumv1alpha1.Node{
				Miner:    true,
				Coinbase: coinbase,
			},
			[]string{
				ArgNatMethod,
				ArgDataPath,
				blockchainDataPath,
				ArgBootnodes,
				bootnode,
				ArgMinerEnabled,
				ArgMinerCoinbase,
				string(coinbase),
			},
		},
	}

	for _, c := range cases {
		func() {
			cc := c
			It(fmt.Sprintf("Should create correct client arguments for %s", cc.title), func() {
				args := reconciler.createArgsForClient(cc.node, cc.join, cc.bootnodes, cc.genesis)
				Expect(args).To(ContainElements(cc.result))
			})
		}()
	}

})
