package ethereum2

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Nimbus Ethereum 2.0 client arguments", func() {

	cases := []struct {
		title  string
		node   *ethereum2v1alpha1.BeaconNode
		result []string
	}{
		{
			title: "beacon node syncing mainnet",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:  ethereum2v1alpha1.NimbusClient,
					Network: "mainnet",
					Logging: sharedAPI.DebugLogs,
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusLogging, string(sharedAPI.DebugLogs)),
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.NimbusClient,
					Network:       "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusEth1Endpoint, "https://localhost:8545"),
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and rpc",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.NimbusClient,
					Network:       "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					RPC:           true,
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusEth1Endpoint, "https://localhost:8545"),
				NimbusRPC,
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and rpc and rpc port",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.NimbusClient,
					Network:       "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					RPC:           true,
					RPCPort:       30303,
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusEth1Endpoint, "https://localhost:8545"),
				NimbusRPC,
				argWithVal(NimbusRPCPort, fmt.Sprintf("%d", 30303)),
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint and rpc with rpc port and host",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.NimbusClient,
					Network:       "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					RPC:           true,
					RPCPort:       30303,
					RPCHost:       "0.0.0.0",
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusEth1Endpoint, "https://localhost:8545"),
				NimbusRPC,
				argWithVal(NimbusRPCPort, fmt.Sprintf("%d", 30303)),
				argWithVal(NimbusRPCAddress, "0.0.0.0"),
			},
		},
		{
			title: "beacon node syncing mainnet with p2p port, eth1 endpoint and rpc with rpc port and host",
			node: &ethereum2v1alpha1.BeaconNode{
				Spec: ethereum2v1alpha1.BeaconNodeSpec{
					Client:        ethereum2v1alpha1.NimbusClient,
					P2PPort:       7891,
					Network:       "mainnet",
					Eth1Endpoints: []string{"https://localhost:8545"},
					RPC:           true,
					RPCPort:       30303,
					RPCHost:       "0.0.0.0",
				},
			},
			result: []string{
				NimbusNonInteractive,
				argWithVal(NimbusTCPPort, "7891"),
				argWithVal(NimbusUDPPort, "7891"),
				argWithVal(NimbusNetwork, "mainnet"),
				argWithVal(NimbusEth1Endpoint, "https://localhost:8545"),
				NimbusRPC,
				argWithVal(NimbusRPCPort, fmt.Sprintf("%d", 30303)),
				argWithVal(NimbusRPCAddress, "0.0.0.0"),
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
