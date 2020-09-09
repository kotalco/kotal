package controllers

import (
	"errors"
	"fmt"
	"strings"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
)

// ParityClient is Go-Ethereum client
type ParityClient struct{}

// LoggingArgFromVerbosity returns logging argument from node verbosity level
func (p *ParityClient) LoggingArgFromVerbosity(level ethereumv1alpha1.VerbosityLevel) string {
	return string(level)
}

// GetArgs returns command line arguments required for client run
func (p *ParityClient) GetArgs(node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network, bootnodes []string) (args []string) {
	// appendArg appends argument with optional value to the arguments array
	appendArg := func(arg ...string) {
		args = append(args, arg...)
	}

	appendArg(ParityLogging, p.LoggingArgFromVerbosity(node.Logging))

	if network.Spec.ID != 0 {
		appendArg(ParityNetworkID, fmt.Sprintf("%d", network.Spec.ID))
	}

	appendArg(ParityDataDir, PathBlockchainData)

	if network.Spec.Join != "" && network.Spec.Join != ethereumv1alpha1.MainNetwork {
		appendArg(ParityNetwork, network.Spec.Join)
	}

	if node.P2PPort != 0 {
		appendArg(ParityP2PPort, fmt.Sprintf("%d", node.P2PPort))
	}

	if len(bootnodes) != 0 {
		commaSeperatedBootnodes := strings.Join(bootnodes, ",")
		appendArg(ParityBootnodes, commaSeperatedBootnodes)
	}

	if node.SyncMode != "" {
		appendArg(ParitySyncMode, string(node.SyncMode))
	}

	if node.Coinbase != "" {
		appendArg(ParityMinerCoinbase, string(node.Coinbase))
		appendArg(ParityUnlock, string(node.Coinbase))
		appendArg(ParityPassword, fmt.Sprintf("%s/account.password", PathSecrets))
	}

	if node.RPCPort != 0 {
		appendArg(ParityRPCHTTPPort, fmt.Sprintf("%d", node.RPCPort))
	}

	if node.RPCHost != "" {
		appendArg(ParityRPCHTTPHost, node.RPCHost)
	}

	if len(node.RPCAPI) != 0 {
		apis := []string{}
		for _, api := range node.RPCAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(ParityRPCHTTPAPI, commaSeperatedAPIs)
	}

	if node.WSPort != 0 {
		appendArg(ParityRPCWSPort, fmt.Sprintf("%d", node.WSPort))
	}

	if node.WSHost != "" {
		appendArg(ParityRPCWSHost, node.WSHost)
	}

	if len(node.WSAPI) != 0 {
		apis := []string{}
		for _, api := range node.WSAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(ParityRPCWSAPI, commaSeperatedAPIs)
	}

	if len(node.Hosts) != 0 {
		commaSeperatedHosts := strings.Join(node.Hosts, ",")
		if node.RPC {
			appendArg(ParityRPCHostWhitelist, commaSeperatedHosts)
		}
	}

	if len(node.CORSDomains) != 0 {
		commaSeperatedDomains := strings.Join(node.CORSDomains, ",")
		if node.RPC {
			appendArg(ParityRPCHTTPCorsOrigins, commaSeperatedDomains)
		}
	}

	return args
}

// GetGenesisFile returns genesis config parameter
func (p *ParityClient) GetGenesisFile(genesis *ethereumv1alpha1.Genesis, consensus ethereumv1alpha1.ConsensusAlgorithm) (content string, err error) {
	// TODO: implement
	err = errors.New("not implemented")
	return
}
