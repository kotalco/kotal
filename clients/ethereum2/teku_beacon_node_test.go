package ethereum2

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Teku Ethereum 2.0 client arguments", func() {

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
				},
			},
			result: []string{
				TekuDataPath,
				TekuNetwork,
				"mainnet",
			},
		},
		{
			title: "beacon node syncing mainnet with multiple eth1 endpoints",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:  ethereum2v1alpha1.TekuClient,
					Network: "mainnet",
					Eth1Endpoints: []string{
						"https://localhost:8545",
						"https://localhost:8546",
					},
				},
			},
			result: []string{
				TekuDataPath,
				TekuNetwork,
				"mainnet",
				TekuEth1Endpoints,
				"https://localhost:8545,https://localhost:8546",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and http enabled",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.TekuClient,
					Network:       "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					REST:          true,
				},
			},
			result: []string{
				TekuDataPath,
				TekuNetwork,
				"mainnet",
				TekuEth1Endpoints,
				"https://localhost:8545",
				TekuRestEnabled,
				TekuRESTAPICorsOrigins,
				"*",
				TekuRESTAPIHostAllowlist,
				"*",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and http enabled with port",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.TekuClient,
					Network:       "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					REST:          true,
					RESTPort:      3333,
				},
			},
			result: []string{
				TekuDataPath,
				TekuNetwork,
				"mainnet",
				TekuEth1Endpoints,
				"https://localhost:8545",
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
			title: "beacon node syncing mainnet with multiple eth1 endpoints and http enabled with port and host",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:  ethereum2v1alpha1.TekuClient,
					Network: "mainnet",
					Eth1Endpoints: []string{
						"https://localhost:8545",
						"https://localhost:8546",
					},
					REST:     true,
					RESTPort: 3333,
					RESTHost: "0.0.0.0",
				},
			},
			result: []string{
				TekuDataPath,
				TekuNetwork,
				"mainnet",
				TekuEth1Endpoints,
				"https://localhost:8545,https://localhost:8546",
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
			title: "beacon node syncing mainnet with p2p port, eth1 endpoint, http enabled with port and host",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.TekuClient,
					P2PPort:       7891,
					Network:       "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					REST:          true,
					RESTPort:      3333,
					RESTHost:      "0.0.0.0",
				},
			},
			result: []string{
				TekuDataPath,
				TekuP2PPort,
				"7891",
				TekuNetwork,
				"mainnet",
				TekuEth1Endpoints,
				"https://localhost:8545",
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
