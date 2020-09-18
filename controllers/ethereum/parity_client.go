package controllers

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
)

// ParityClient is Go-Ethereum client
type ParityClient struct {
	NetworkID uint
}

// LoggingArgFromVerbosity returns logging argument from node verbosity level
func (p *ParityClient) LoggingArgFromVerbosity(level ethereumv1alpha1.VerbosityLevel) string {
	return string(level)
}

// PrunningArgFromSyncMode returns prunning arg from sync mode
func (p *ParityClient) PrunningArgFromSyncMode(mode ethereumv1alpha1.SynchronizationMode) string {
	m := map[ethereumv1alpha1.SynchronizationMode]string{
		ethereumv1alpha1.FullSynchronization: "archive",
		ethereumv1alpha1.FastSynchronization: "fast",
	}
	return m[mode]
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

	if network.Spec.Genesis == nil {
		if network.Spec.Join != ethereumv1alpha1.MainNetwork {
			appendArg(ParityNetwork, network.Spec.Join)
		}
	} else {
		appendArg(ParityNetwork, fmt.Sprintf("%s/genesis.json", PathConfig))
	}

	if node.P2PPort != 0 {
		appendArg(ParityP2PPort, fmt.Sprintf("%d", node.P2PPort))
	}

	if len(bootnodes) != 0 {
		commaSeperatedBootnodes := strings.Join(bootnodes, ",")
		appendArg(ParityBootnodes, commaSeperatedBootnodes)
	}

	if node.SyncMode != "" {
		appendArg(ParitySyncMode, p.PrunningArgFromSyncMode(node.SyncMode))
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

// NormalizeNonce normalizes nonce to be 8 bytes (16 hex digits)
func (p *ParityClient) NormalizeNonce(data string) string {
	n := new(big.Int)
	i, _ := n.SetString(data, 16)
	return fmt.Sprintf("%#0.16x", i)
}

// GetGenesisFile returns genesis config parameter
func (p *ParityClient) GetGenesisFile(network *ethereumv1alpha1.Network) (content string, err error) {
	genesis := network.Spec.Genesis
	consensus := network.Spec.Consensus
	extraData := "0x0"
	var engineConfig map[string]interface{}

	// clique PoA settings
	if consensus == ethereumv1alpha1.ProofOfAuthority {
		extraData = createExtraDataFromSigners(genesis.Clique.Signers)
		engineConfig = map[string]interface{}{
			"clique": map[string]interface{}{
				"params": map[string]interface{}{
					"period": genesis.Clique.BlockPeriod,
					"epoch":  genesis.Clique.EpochLength,
				},
			},
		}
	}

	// ethash PoW settings
	if consensus == ethereumv1alpha1.ProofOfWork {
	}

	genesisConfig := map[string]interface{}{
		"seal": map[string]interface{}{
			"ethereum": map[string]interface{}{
				"nonce":   p.NormalizeNonce(string(genesis.Nonce)[2:]),
				"mixHash": genesis.MixHash,
			},
		},
		"parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
		"timestamp":  genesis.Timestamp,
		"gasLimit":   genesis.GasLimit,
		"difficulty": genesis.Difficulty,
		"author":     genesis.Coinbase,
		"extraData":  extraData,
	}

	spuriousDragonBlock := fmt.Sprintf("%#x", genesis.Forks.EIP155)
	byzantiumBlock := fmt.Sprintf("%#x", genesis.Forks.Byzantium)
	constantinopleBlock := fmt.Sprintf("%#x", genesis.Forks.Constantinople)
	istanbulBlock := fmt.Sprintf("%#x", genesis.Forks.Istanbul)

	paramsConfig := map[string]interface{}{
		// other non fork parameters
		"chainID":              fmt.Sprintf("%#x", genesis.ChainID),
		"accountStartNonce":    "0x0",
		"gasLimitBoundDivisor": "0x400",
		"maximumExtraDataSize": "0xffff",
		"minGasLimit":          "0x1388",
		"networkID":            fmt.Sprintf("%#x", network.Spec.ID),
		// Tingerine Whistle
		"eip150Transition": fmt.Sprintf("%#x", genesis.Forks.EIP150),
		// Spurious Dragon
		"eip155Transition":      spuriousDragonBlock,
		"eip160Transition":      spuriousDragonBlock,
		"eip161abcTransition":   spuriousDragonBlock,
		"eip161dTransition":     spuriousDragonBlock,
		"maxCodeSizeTransition": spuriousDragonBlock, //eip170
		"maxCodeSize":           "0x6000",
		// Byzantium
		"eip140Transition": byzantiumBlock,
		"eip211Transition": byzantiumBlock,
		"eip214Transition": byzantiumBlock,
		"eip658Transition": byzantiumBlock,
		// Constantinople
		"eip145Transition":  constantinopleBlock,
		"eip1014Transition": constantinopleBlock,
		"eip1052Transition": constantinopleBlock,
		"eip1283Transition": constantinopleBlock,
		// PetersBurg
		"eip1283DisableTransition": genesis.Forks.Petersburg,
		// Istanbul
		"eip1283ReenableTransition": istanbulBlock,
		"eip1344Transition":         istanbulBlock,
		"eip1706Transition":         istanbulBlock,
		"eip1884Transition":         istanbulBlock,
		"eip2028Transition":         istanbulBlock,
	}

	alloc := genesisAccounts(true)
	for _, account := range genesis.Accounts {
		m := map[string]interface{}{
			"balance": account.Balance,
		}

		if account.Code != "" {
			m["code"] = account.Code
		}

		if account.Storage != nil {
			m["storage"] = account.Storage
		}

		alloc[string(account.Address)] = m
	}

	result := map[string]interface{}{
		"name":     "network",
		"genesis":  genesisConfig,
		"params":   paramsConfig,
		"engine":   engineConfig,
		"accounts": alloc,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return
	}

	content = string(data)

	return
}
