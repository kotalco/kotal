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
		node   *ethereum2v1alpha1.BeaconNode
		result []string
	}{
		{
			title: "beacon node syncing mainnet",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client: ethereum2v1alpha1.PrysmClient,
					Join:   "mainnet",
					RPC:    true,
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				"--mainnet",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.PrysmClient,
					Join:          "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					RPC:           true,
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				"--mainnet",
				PrysmWeb3Provider,
				"https://localhost:8545",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and rpc port",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.PrysmClient,
					Join:          "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					RPC:           true,
					RPCPort:       9976,
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				"--mainnet",
				PrysmWeb3Provider,
				"https://localhost:8545",
				PrysmRPCPort,
				"9976",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and rpc port and host",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.PrysmClient,
					Join:          "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					RPC:           true,
					RPCPort:       9976,
					RPCHost:       "0.0.0.0",
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
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
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.PrysmClient,
					Join:          "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				"--mainnet",
				PrysmWeb3Provider,
				"https://localhost:8545",
				PrysmDisableGRPC,
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and grpc with port",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client: ethereum2v1alpha1.PrysmClient,
					Join:   "mainnet",
					Eth1Endpoints: []string{
						"https://localhost:8545",
						"https://localhost:8546",
						"https://localhost:8547",
					},
					GRPC:     true,
					GRPCPort: 4445,
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				"--mainnet",
				PrysmWeb3Provider,
				"https://localhost:8545",
				PrysmFallbackWeb3Provider,
				"https://localhost:8546",
				PrysmFallbackWeb3Provider,
				"https://localhost:8547",
				PrysmGRPCPort,
				"4445",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and grpc with port and host",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client: ethereum2v1alpha1.PrysmClient,
					Join:   "mainnet",
					Eth1Endpoints: []string{
						"https://localhost:8545",
						"https://localhost:8546",
					},
					GRPC:     true,
					GRPCPort: 4445,
					GRPCHost: "0.0.0.0",
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				"--mainnet",
				PrysmWeb3Provider,
				"https://localhost:8545",
				PrysmFallbackWeb3Provider,
				"https://localhost:8546",
				PrysmGRPCPort,
				"4445",
				PrysmGRPCHost,
				"0.0.0.0",
			},
		},
		{
			title: "beacon node syncing mainnet with p2p port, eth1 endpoint and grpc with port and host",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:  ethereum2v1alpha1.PrysmClient,
					P2PPort: 7891,
					Join:    "mainnet",
					Eth1Endpoints: []string{
						"https://localhost:8545",
						"https://localhost:8546",
					},
					GRPC:     true,
					GRPCPort: 4445,
					GRPCHost: "0.0.0.0",
				},
			},
			result: []string{
				PrysmAcceptTermsOfUse,
				PrysmDataDir,
				PrysmP2PTCPPort,
				"7891",
				PrysmP2PUDPPort,
				"7891",
				"--mainnet",
				PrysmWeb3Provider,
				"https://localhost:8545",
				PrysmFallbackWeb3Provider,
				"https://localhost:8546",
				PrysmGRPCPort,
				"4445",
				PrysmGRPCHost,
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
