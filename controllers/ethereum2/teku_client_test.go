package controllers

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Teku Ethereum 2.0 client arguments", func() {

	cases := []struct {
		title  string
		node   *ethereum2v1alpha1.Node
		result []string
	}{
		{
			title: "beacon node syncing mainnet",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client: ethereum2v1alpha1.TekuClient,
					Join:   "mainnet",
				},
			},
			result: []string{
				TekuDataPath,
				PathBlockchainData,
				TekuNetwork,
				"mainnet",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:        ethereum2v1alpha1.TekuClient,
					Join:          "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
				},
			},
			result: []string{
				TekuDataPath,
				PathBlockchainData,
				TekuNetwork,
				"mainnet",
				TekuEth1Endpoint,
				"https://localhost:8545",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and http enabled",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:        ethereum2v1alpha1.TekuClient,
					Join:          "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					REST:          true,
				},
			},
			result: []string{
				TekuDataPath,
				PathBlockchainData,
				TekuNetwork,
				"mainnet",
				TekuEth1Endpoint,
				"https://localhost:8545",
				TekuRestEnabled,
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and http enabled with port",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:        ethereum2v1alpha1.TekuClient,
					Join:          "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					REST:          true,
					RESTPort:      3333,
				},
			},
			result: []string{
				TekuDataPath,
				PathBlockchainData,
				TekuNetwork,
				"mainnet",
				TekuEth1Endpoint,
				"https://localhost:8545",
				TekuRestEnabled,
				TekuRestPort,
				"3333",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and http enabled with port and host",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:        ethereum2v1alpha1.TekuClient,
					Join:          "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					REST:          true,
					RESTPort:      3333,
					RESTHost:      "0.0.0.0",
				},
			},
			result: []string{
				TekuDataPath,
				PathBlockchainData,
				TekuNetwork,
				"mainnet",
				TekuEth1Endpoint,
				"https://localhost:8545",
				TekuRestEnabled,
				TekuRestPort,
				"3333",
				TekuRestHost,
				"0.0.0.0",
			},
		},
		{
			title: "beacon node syncing mainnet with p2p port, eth1 endpoint, http enabled with port and host",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:        ethereum2v1alpha1.TekuClient,
					P2PPort:       7891,
					Join:          "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					REST:          true,
					RESTPort:      3333,
					RESTHost:      "0.0.0.0",
				},
			},
			result: []string{
				TekuDataPath,
				PathBlockchainData,
				TekuP2PPort,
				"7891",
				TekuNetwork,
				"mainnet",
				TekuEth1Endpoint,
				"https://localhost:8545",
				TekuRestEnabled,
				TekuRestPort,
				"3333",
				TekuRestHost,
				"0.0.0.0",
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
