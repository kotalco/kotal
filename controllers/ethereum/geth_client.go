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

// GetArgs returns command line arguments required for client run
func (g *GethClient) GetArgs(node *ethereumv1alpha1.Node, network *ethereumv1alpha1.Network, bootnodes []string) (args []string) {
	// appendArg appends argument with optional value to the arguments array
	appendArg := func(arg ...string) {
		args = append(args, arg...)
	}

	appendArg("--nousb")

	appendArg(GethLogging, g.LoggingArgFromVerbosity(node.Logging))

	if network.Spec.ID != 0 {
		appendArg(GethNetworkID, fmt.Sprintf("%d", network.Spec.ID))
	}

	if node.WithNodekey() {
		appendArg(GethNodeKey, fmt.Sprintf("%s/nodekey", PathSecrets))
	}

	appendArg(GethDataDir, PathBlockchainData)

	if network.Spec.Join != "" && network.Spec.Join != ethereumv1alpha1.MainNetwork {
		appendArg(fmt.Sprintf("--%s", network.Spec.Join))
	}

	if node.P2PPort != 0 {
		appendArg(GethP2PPort, fmt.Sprintf("%d", node.P2PPort))
	}

	if len(bootnodes) != 0 {
		commaSeperatedBootnodes := strings.Join(bootnodes, ",")
		appendArg(GethBootnodes, commaSeperatedBootnodes)
	}

	if node.SyncMode != "" {
		appendArg(GethSyncMode, string(node.SyncMode))
	}

	if node.Miner {
		appendArg(GethMinerEnabled)
	}

	if node.Coinbase != "" {
		appendArg(GethMinerCoinbase, string(node.Coinbase))
		appendArg(GethUnlock, string(node.Coinbase))
		appendArg(GethPassword, fmt.Sprintf("%s/account.password", PathSecrets))
	}

	if node.RPC {
		appendArg(GethRPCHTTPEnabled)
	}

	if node.RPCPort != 0 {
		appendArg(GethRPCHTTPPort, fmt.Sprintf("%d", node.RPCPort))
	}

	if node.RPCHost != "" {
		appendArg(GethRPCHTTPHost, node.RPCHost)
	}

	if len(node.RPCAPI) != 0 {
		apis := []string{}
		for _, api := range node.RPCAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(GethRPCHTTPAPI, commaSeperatedAPIs)
	}

	if node.WS {
		appendArg(GethRPCWSEnabled)
	}

	if node.WSPort != 0 {
		appendArg(GethRPCWSPort, fmt.Sprintf("%d", node.WSPort))
	}

	if node.WSHost != "" {
		appendArg(GethRPCWSHost, node.WSHost)
	}

	if len(node.WSAPI) != 0 {
		apis := []string{}
		for _, api := range node.WSAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(GethRPCWSAPI, commaSeperatedAPIs)
	}

	if node.GraphQL {
		appendArg(GethGraphQLHTTPEnabled)
	}

	if node.GraphQLPort != 0 {
		appendArg(GethGraphQLHTTPPort, fmt.Sprintf("%d", node.GraphQLPort))
	}

	if node.GraphQLHost != "" {
		appendArg(GethGraphQLHTTPHost, node.GraphQLHost)
	}

	if len(node.Hosts) != 0 {
		commaSeperatedHosts := strings.Join(node.Hosts, ",")
		if node.RPC {
			appendArg(GethRPCHostWhitelist, commaSeperatedHosts)
		}
		if node.GraphQL {
			appendArg(GethGraphQLHostWhitelist, commaSeperatedHosts)
		}
	}

	if len(node.CORSDomains) != 0 {
		commaSeperatedDomains := strings.Join(node.CORSDomains, ",")
		if node.RPC {
			appendArg(GethRPCHTTPCorsOrigins, commaSeperatedDomains)
		}
		if node.GraphQL {
			appendArg(GethGraphQLHTTPCorsOrigins, commaSeperatedDomains)
		}
	}

	return args
}

// GetGenesisFile returns genesis config parameter
func (g *GethClient) GetGenesisFile(network *ethereumv1alpha1.Network) (content string, err error) {
	genesis := network.Spec.Genesis
	consensus := network.Spec.Consensus
	mixHash := genesis.MixHash
	nonce := genesis.Nonce
	difficulty := genesis.Difficulty
	result := map[string]interface{}{}

	var consensusConfig map[string]uint
	var extraData string
	var engine string

	if consensus == ethereumv1alpha1.ProofOfWork {
		consensusConfig = map[string]uint{}

		if genesis.Ethash.FixedDifficulty != nil {
			consensusConfig["fixeddifficulty"] = *genesis.Ethash.FixedDifficulty
		}

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

	// clique ibft2 settings
	if consensus == ethereumv1alpha1.IstanbulBFT {

		consensusConfig = map[string]uint{
			"blockperiodseconds":        genesis.IBFT2.BlockPeriod,
			"epochlength":               genesis.IBFT2.EpochLength,
			"requesttimeoutseconds":     genesis.IBFT2.RequestTimeout,
			"messageQueueLimit":         genesis.IBFT2.MessageQueueLimit,
			"duplicateMessageLimit":     genesis.IBFT2.DuplicateMessageLimit,
			"futureMessagesLimit":       genesis.IBFT2.FutureMessagesLimit,
			"futureMessagesMaxDistance": genesis.IBFT2.FutureMessagesMaxDistance,
		}
		engine = "ibft2"
		mixHash = "0x63746963616c2062797a616e74696e65206661756c7420746f6c6572616e6365"
		nonce = "0x0"
		difficulty = "0x1"
		extraData, err = createExtraDataFromValidators(genesis.IBFT2.Validators)
		if err != nil {
			return
		}
	}

	config := map[string]interface{}{
		"chainId":             genesis.ChainID,
		"homesteadBlock":      genesis.Forks.Homestead,
		"eip150Block":         genesis.Forks.EIP150,
		"eip150Hash":          genesis.Forks.EIP150Hash,
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
