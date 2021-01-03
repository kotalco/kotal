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
		node   *ethereum2v1alpha1.Node
		result []string
	}{
		{
			title: "beacon node syncing mainnet",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client: ethereum2v1alpha1.LighthouseClient,
					Join:   "mainnet",
				},
			},
			result: []string{
				LighthouseDataDir,
				PathBlockchainData,
				LighthouseNetwork,
				"mainnet",
			},
		},
		{
			title: "beacon node syncing mainnet with eth1 endpoint",
			node: &ethereum2v1alpha1.Node{
				Spec: ethereum2v1alpha1.NodeSpec{
					Client:       ethereum2v1alpha1.LighthouseClient,
					Join:         "mainnet",
					Eth1Endpoint: "https://localhost:8545",
				},
			},
			result: []string{
				LighthouseDataDir,
				PathBlockchainData,
				LighthouseNetwork,
				"mainnet",
				LighthouseEth1Endpoints,
				"https://localhost:8545",
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
