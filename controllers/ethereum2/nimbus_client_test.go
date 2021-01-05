package controllers

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nimbus Ethereum 2.0 client arguments", func() {

	cases := []struct {
		title  string
		node   *ethereum2v1alpha1.Node
		result []string
	}{
		{
			title: "beacon node syncing mainnet",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client: ethereum2v1alpha1.NimbusClient,
					Join:   "mainnet",
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusDataDir, PathBlockchainData),
				argWithVal(NimbusNetwork, "mainnet"),
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:       ethereum2v1alpha1.NimbusClient,
					Join:         "mainnet",
					Eth1Endpoint: "https://localhost:8545",
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusDataDir, PathBlockchainData),
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusEth1Endpoint, "https://localhost:8545"),
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and rpc",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:       ethereum2v1alpha1.NimbusClient,
					Join:         "mainnet",
					Eth1Endpoint: "https://localhost:8545",
					RPC:          true,
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusDataDir, PathBlockchainData),
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusEth1Endpoint, "https://localhost:8545"),
				NimbusRPC,
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and rpc and rpc port",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:       ethereum2v1alpha1.NimbusClient,
					Join:         "mainnet",
					Eth1Endpoint: "https://localhost:8545",
					RPC:          true,
					RPCPort:      30303,
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusDataDir, PathBlockchainData),
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusEth1Endpoint, "https://localhost:8545"),
				NimbusRPC,
				argWithVal(NimbusRPCPort, fmt.Sprintf("%d", 30303)),
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
