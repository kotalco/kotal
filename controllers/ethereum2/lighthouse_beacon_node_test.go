package controllers

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Lighthouse Ethereum 2.0 client arguments", func() {

	cases := []struct {
		title  string
		node   *ethereum2v1alpha1.BeaconNode
		result []string
	}{
		{
			title: "beacon node syncing mainnet",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client: ethereum2v1alpha1.LighthouseClient,
					Join:   "mainnet",
				},
			},
			result: []string{
				LighthouseDataDir,
				LighthouseNetwork,
				"mainnet",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.LighthouseClient,
					Join:          "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
				},
			},
			result: []string{
				LighthouseDataDir,
				LighthouseNetwork,
				"mainnet",
				LighthouseEth1Endpoints,
				"https://localhost:8545",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and http enabled",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.LighthouseClient,
					Join:          "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					REST:          true,
				},
			},
			result: []string{
				LighthouseDataDir,
				LighthouseNetwork,
				"mainnet",
				LighthouseEth1Endpoints,
				"https://localhost:8545",
				LighthouseHTTP,
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and http enabled with port",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.LighthouseClient,
					Join:          "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					REST:          true,
					RESTPort:      4444,
				},
			},
			result: []string{
				LighthouseDataDir,
				LighthouseNetwork,
				"mainnet",
				LighthouseEth1Endpoints,
				"https://localhost:8545",
				LighthouseHTTP,
				LighthouseHTTPPort,
				"4444",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and http enabled with port and host",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client: ethereum2v1alpha1.LighthouseClient,
					Join:   "mainnet",
					Eth1Endpoints: []string{
						"https://localhost:8545",
						"https://localhost:8546",
					},
					REST:     true,
					RESTPort: 4444,
					RESTHost: "0.0.0.0",
				},
			},
			result: []string{
				LighthouseDataDir,
				LighthouseNetwork,
				"mainnet",
				LighthouseEth1Endpoints,
				"https://localhost:8545,https://localhost:8546",
				LighthouseHTTP,
				LighthouseHTTPPort,
				"4444",
				LighthouseHTTPAddress,
				"0.0.0.0",
			},
		},
		{
			title: "beacon node syncing mainnet with p2p port, eth1 endpoint, http enabled with port and host",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:  ethereum2v1alpha1.LighthouseClient,
					P2PPort: 7891,
					Join:    "mainnet",
					Eth1Endpoints: []string{
						"https://localhost:8545",
						"https://localhost:8546",
					},
					REST:     true,
					RESTPort: 4444,
					RESTHost: "0.0.0.0",
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
				LighthouseEth1Endpoints,
				"https://localhost:8545,https://localhost:8546",
				LighthouseHTTP,
				LighthouseHTTPPort,
				"4444",
				LighthouseHTTPAddress,
				"0.0.0.0",
			},
		},
	}

	for _, c := range cases {
		func() {
			cc := c
			It(fmt.Sprintf("Should create correct client arguments for %s", cc.title), func() {
				cc.node.Default()
				client, _ := NewEthereum2Client(cc.node)
				args := client.Args()
				Expect(args).To(ContainElements(cc.result))
			})
		}()
	}

})
