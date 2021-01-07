package controllers

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Prysm Ethereum 2.0 client arguments", func() {

	cases := []struct {
		title  string
		node   *ethereum2v1alpha1.Node
		result []string
	}{
		{
			title: "beacon node syncing mainnet",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client: ethereum2v1alpha1.PrysmClient,
					Join:   "mainnet",
					RPC:    true,
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				PathBlockchainData,
				"--mainnet",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:       ethereum2v1alpha1.PrysmClient,
					Join:         "mainnet",
					Eth1Endpoint: "https://localhost:8545",
					RPC:          true,
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				PathBlockchainData,
				"--mainnet",
				PrysmWeb3Provider,
				"https://localhost:8545",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and rpc port",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:       ethereum2v1alpha1.PrysmClient,
					Join:         "mainnet",
					Eth1Endpoint: "https://localhost:8545",
					RPC:          true,
					RPCPort:      9976,
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				PathBlockchainData,
				"--mainnet",
				PrysmWeb3Provider,
				"https://localhost:8545",
				PrysmRPCPort,
				"9976",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and rpc port and host",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:       ethereum2v1alpha1.PrysmClient,
					Join:         "mainnet",
					Eth1Endpoint: "https://localhost:8545",
					RPC:          true,
					RPCPort:      9976,
					RPCHost:      "0.0.0.0",
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				PathBlockchainData,
				"--mainnet",
				PrysmWeb3Provider,
				"https://localhost:8545",
				PrysmRPCPort,
				"9976",
				PrysmRPCHost,
				"0.0.0.0",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and grpc disabled",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:       ethereum2v1alpha1.PrysmClient,
					Join:         "mainnet",
					Eth1Endpoint: "https://localhost:8545",
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				PathBlockchainData,
				"--mainnet",
				PrysmWeb3Provider,
				"https://localhost:8545",
				PrysmDisableGRPC,
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and grpc",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:       ethereum2v1alpha1.PrysmClient,
					Join:         "mainnet",
					Eth1Endpoint: "https://localhost:8545",
					GRPC:         true,
					GRPCPort:     4445,
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				PathBlockchainData,
				"--mainnet",
				PrysmWeb3Provider,
				"https://localhost:8545",
				PrysmGRPCPort,
				"4445",
			},
		},
	}

	for _, c := range cases {
		func() {
			cc := c
			It(fmt.Sprintf("Should create correct client arguments for %s", cc.title), func() {
				cc.node.Default()
				client, err := NewEthereum2Client(cc.node.Spec.Client)
				Expect(err).To(BeNil())
				args := client.Args(cc.node)
				Expect(args).To(ContainElements(cc.result))
			})
		}()
	}

})
