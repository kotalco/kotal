package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
)

// GethClient is Go-Ethereum client
type GethClient struct{}

// LoggingArgFromVerbosity returns logging argument from node verbosity level
func (g *GethClient) LoggingArgFromVerbosity(level ethereumv1alpha1.VerbosityLevel) string {
	levels := map[ethereumv1alpha1.VerbosityLevel]string{
		ethereumv1alpha1.NoLogs:    "0",
		ethereumv1alpha1.ErrorLogs: "1",
		ethereumv1alpha1.WarnLogs:  "2",
		ethereumv1alpha1.InfoLogs:  "3",
		ethereumv1alpha1.DebugLogs: "4",
		ethereumv1alpha1.AllLogs:   "5",
	}

	return levels[level]
}

// Args returns command line arguments required for client run
func (g *GethClient) Args(node *ethereumv1alpha1.Node) (args []string) {
	// appendArg appends argument with optional value to the arguments array
	appendArg := func(arg ...string) {
		args = append(args, arg...)
	}

	appendArg(GethLogging, g.LoggingArgFromVerbosity(node.Spec.Logging))

	appendArg(GethConfig, fmt.Sprintf("%s/config.toml", PathConfig))

	if node.Spec.ID != 0 {
		appendArg(GethNetworkID, fmt.Sprintf("%d", node.Spec.ID))
	}

	if node.Spec.Nodekey != "" {
		appendArg(GethNodeKey, fmt.Sprintf("%s/nodekey", PathSecrets))
	}

	if len(node.Spec.Bootnodes) != 0 {
		bootnodes := []string{}
		for _, bootnode := range node.Spec.Bootnodes {
			bootnodes = append(bootnodes, string(bootnode))
		}
		appendArg(GethBootnodes, strings.Join(bootnodes, ","))
	}

	if node.Spec.Genesis != nil {
		appendArg(GethNoDiscovery)
	}

	appendArg(GethDataDir, PathBlockchainData)

	if node.Spec.Join != "" && node.Spec.Join != ethereumv1alpha1.MainNetwork {
		appendArg(fmt.Sprintf("--%s", node.Spec.Join))
	}

	if node.Spec.P2PPort != 0 {
		appendArg(GethP2PPort, fmt.Sprintf("%d", node.Spec.P2PPort))
	}

	if node.Spec.SyncMode != "" {
		appendArg(GethSyncMode, string(node.Spec.SyncMode))
	}

	if node.Spec.Miner {
		appendArg(GethMinerEnabled)
	}

	if node.Spec.Coinbase != "" {
		appendArg(GethMinerCoinbase, string(node.Spec.Coinbase))
		appendArg(GethUnlock, string(node.Spec.Coinbase))
		appendArg(GethPassword, fmt.Sprintf("%s/account.password", PathSecrets))
	}

	if node.Spec.RPC {
		appendArg(GethRPCHTTPEnabled)
		appendArg(GethRPCHTTPHost, DefaultHost)
	}

	if node.Spec.RPCPort != 0 {
		appendArg(GethRPCHTTPPort, fmt.Sprintf("%d", node.Spec.RPCPort))
	}

	if len(node.Spec.RPCAPI) != 0 {
		apis := []string{}
		for _, api := range node.Spec.RPCAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(GethRPCHTTPAPI, commaSeperatedAPIs)
	}

	if node.Spec.WS {
		appendArg(GethRPCWSEnabled)
		appendArg(GethRPCWSHost, DefaultHost)
	}

	if node.Spec.WSPort != 0 {
		appendArg(GethRPCWSPort, fmt.Sprintf("%d", node.Spec.WSPort))
	}

	if len(node.Spec.WSAPI) != 0 {
		apis := []string{}
		for _, api := range node.Spec.WSAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(GethRPCWSAPI, commaSeperatedAPIs)
	}

	if node.Spec.GraphQL {
		appendArg(GethGraphQLHTTPEnabled)
	}

	//NOTE: .GraphQLPort is ignored because rpc port will be used by graphql server
	// .GraphQLPort will be used in the service that point to the pod

	if len(node.Spec.Hosts) != 0 {
		commaSeperatedHosts := strings.Join(node.Spec.Hosts, ",")
		if node.Spec.RPC {
			appendArg(GethRPCHostWhitelist, commaSeperatedHosts)
		}
		if node.Spec.GraphQL {
			appendArg(GethGraphQLHostWhitelist, commaSeperatedHosts)
		}
		// no ws hosts settings
	}

	if len(node.Spec.CORSDomains) != 0 {
		commaSeperatedDomains := strings.Join(node.Spec.CORSDomains, ",")
		if node.Spec.RPC {
			appendArg(GethRPCHTTPCorsOrigins, commaSeperatedDomains)
		}
		if node.Spec.GraphQL {
			appendArg(GethGraphQLHTTPCorsOrigins, commaSeperatedDomains)
		}
		if node.Spec.WS {
			appendArg(GethWSOrigins, commaSeperatedDomains)
		}
	}

	return args
}

// EncodeStaticNodes returns the static nodes
// [Node.P2P]
// StaticNodes = [enodeURL1, enodeURL2 ...]
func (g *GethClient) EncodeStaticNodes(node *ethereumv1alpha1.Node) string {

	var encoded []byte

	if len(node.Spec.StaticNodes) == 0 {
		encoded = []byte("[]")
	} else {
		encoded, _ = json.Marshal(node.Spec.StaticNodes)
	}

	return fmt.Sprintf("[Node.P2P]\nStaticNodes = %s", string(encoded))
}

// Genesis returns genesis config parameter
func (g *GethClient) Genesis(node *ethereumv1alpha1.Node) (content string, err error) {
	genesis := node.Spec.Genesis
	consensus := node.Spec.Consensus
	mixHash := genesis.MixHash
	nonce := genesis.Nonce
	extraData := "0x00"
	difficulty := genesis.Difficulty
	result := map[string]interface{}{}

	var consensusConfig map[string]uint
	var engine string

	if consensus == ethereumv1alpha1.ProofOfWork {
		consensusConfig = map[string]uint{}
		engine = "ethash"
	}

	// clique PoA settings
	if consensus == ethereumv1alpha1.ProofOfAuthority {
		consensusConfig = map[string]uint{
			"period": genesis.Clique.BlockPeriod,
			"epoch":  genesis.Clique.EpochLength,
		}
		engine = "clique"
		extraData = createExtraDataFromSigners(genesis.Clique.Signers)
	}

	config := map[string]interface{}{
		"chainId":             genesis.ChainID,
		"homesteadBlock":      genesis.Forks.Homestead,
		"eip150Block":         genesis.Forks.EIP150,
		"eip155Block":         genesis.Forks.EIP155,
		"eip158Block":         genesis.Forks.EIP158,
		"byzantiumBlock":      genesis.Forks.Byzantium,
		"constantinopleBlock": genesis.Forks.Constantinople,
		"petersburgBlock":     genesis.Forks.Petersburg,
		"istanbulBlock":       genesis.Forks.Istanbul,
		"muirGlacierBlock":    genesis.Forks.MuirGlacier,
		engine:                consensusConfig,
	}

	if genesis.Forks.DAO != nil {
		config["daoForkBlock"] = genesis.Forks.DAO
		config["daoForkSupport"] = true
	}

	result["config"] = config

	result["nonce"] = nonce
	result["timestamp"] = genesis.Timestamp
	result["gasLimit"] = genesis.GasLimit
	result["difficulty"] = difficulty
	result["coinbase"] = genesis.Coinbase
	result["mixHash"] = mixHash
	result["extraData"] = extraData

	alloc := genesisAccounts(false)
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

	result["alloc"] = alloc

	data, err := json.Marshal(result)
	if err != nil {
		return
	}

	content = string(data)

	return
}
