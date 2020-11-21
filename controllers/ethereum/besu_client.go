package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
)

// BesuClient is Hyperledger Besu client
type BesuClient struct{}

// LoggingArgFromVerbosity returns logging argument from node verbosity level
func (b *BesuClient) LoggingArgFromVerbosity(level ethereumv1alpha1.VerbosityLevel) string {
	return strings.ToUpper(string(level))
}

// GetArgs returns command line arguments required for client run
func (b *BesuClient) GetArgs(node *ethereumv1alpha1.XNode, network *ethereumv1alpha1.Network) (args []string) {

	// appendArg appends argument with optional value to the arguments array
	appendArg := func(arg ...string) {
		args = append(args, arg...)
	}

	appendArg(BesuNatMethod, "KUBERNETES")

	appendArg(BesuLogging, b.LoggingArgFromVerbosity(node.Logging))

	if network.Spec.ID != 0 {
		appendArg(BesuNetworkID, fmt.Sprintf("%d", network.Spec.ID))
	}

	if node.WithNodekey() {
		appendArg(BesuNodePrivateKey, fmt.Sprintf("%s/nodekey", PathSecrets))
	}

	if network.Spec.Genesis != nil {
		appendArg(BesuGenesisFile, fmt.Sprintf("%s/genesis.json", PathConfig))
	}

	appendArg(BesuDataPath, PathBlockchainData)

	if network.Spec.Genesis == nil {
		appendArg(BesuNetwork, network.Spec.Join)
	} else {
		appendArg(BesuDiscoveryEnabled, "false")
	}

	if node.P2PPort != 0 {
		appendArg(BesuP2PPort, fmt.Sprintf("%d", node.P2PPort))
	}

	if node.SyncMode != "" {
		appendArg(BesuSyncMode, string(node.SyncMode))
	}

	if node.Miner {
		appendArg(BesuMinerEnabled)
	}

	if node.Coinbase != "" {
		appendArg(BesuMinerCoinbase, string(node.Coinbase))
	}

	if node.RPC {
		appendArg(BesuRPCHTTPEnabled)
		appendArg(BesuRPCHTTPHost, DefaultHost)
	}

	if node.RPCPort != 0 {
		appendArg(BesuRPCHTTPPort, fmt.Sprintf("%d", node.RPCPort))
	}

	if len(node.RPCAPI) != 0 {
		apis := []string{}
		for _, api := range node.RPCAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(BesuRPCHTTPAPI, commaSeperatedAPIs)
	}

	if node.WS {
		appendArg(BesuRPCWSEnabled)
		appendArg(BesuRPCWSHost, DefaultHost)
	}

	if node.WSPort != 0 {
		appendArg(BesuRPCWSPort, fmt.Sprintf("%d", node.WSPort))
	}

	if len(node.WSAPI) != 0 {
		apis := []string{}
		for _, api := range node.WSAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(BesuRPCWSAPI, commaSeperatedAPIs)
	}

	if node.GraphQL {
		appendArg(BesuGraphQLHTTPEnabled)
		appendArg(BesuGraphQLHTTPHost, DefaultHost)
	}

	if node.GraphQLPort != 0 {
		appendArg(BesuGraphQLHTTPPort, fmt.Sprintf("%d", node.GraphQLPort))
	}

	if len(node.Hosts) != 0 {
		commaSeperatedHosts := strings.Join(node.Hosts, ",")
		appendArg(BesuHostAllowlist, commaSeperatedHosts)
	}

	if len(node.CORSDomains) != 0 {
		commaSeperatedDomains := strings.Join(node.CORSDomains, ",")
		if node.RPC {
			appendArg(BesuRPCHTTPCorsOrigins, commaSeperatedDomains)
		}
		if node.GraphQL {
			appendArg(BesuGraphQLHTTPCorsOrigins, commaSeperatedDomains)
		}
		// no ws cors setting
	}

	return args
}

// GetGenesisFile returns genesis config parameter
func (b *BesuClient) GetGenesisFile(network *ethereumv1alpha1.Network) (content string, err error) {
	genesis := network.Spec.Genesis
	consensus := network.Spec.Consensus
	mixHash := genesis.MixHash
	nonce := genesis.Nonce
	extraData := "0x00"
	difficulty := genesis.Difficulty
	result := map[string]interface{}{}

	var consensusConfig map[string]uint
	var engine string

	// ethash PoW settings
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
			"blockperiodseconds": genesis.Clique.BlockPeriod,
			"epochlength":        genesis.Clique.EpochLength,
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
