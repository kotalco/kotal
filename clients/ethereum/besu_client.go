package ethereum

import (
	"encoding/json"
	"fmt"
	"strings"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// BesuClient is Hyperledger Besu client
// https://github.com/hyperledger/besu
type BesuClient struct {
	node *ethereumv1alpha1.Node
}

const (
	// BesuHomeDir is besu docker image home directory
	BesuHomeDir = "/opt/besu"
)

// HomeDir returns besu client home directory
func (b *BesuClient) HomeDir() string {
	return BesuHomeDir
}

func (b *BesuClient) Command() []string {
	return nil
}

func (b *BesuClient) Env() []corev1.EnvVar {
	return nil
}

// Args returns command line arguments required for client run
func (b *BesuClient) Args() (args []string) {

	node := b.node

	args = append(args, BesuNatMethod, "KUBERNETES")
	args = append(args, BesuDataPath, shared.PathData(b.HomeDir()))
	args = append(args, BesuP2PPort, fmt.Sprintf("%d", node.Spec.P2PPort))
	args = append(args, BesuSyncMode, string(node.Spec.SyncMode))
	args = append(args, BesuLogging, strings.ToUpper(string(node.Spec.Logging)))

	if node.Spec.NodePrivateKeySecretName != "" {
		args = append(args, BesuNodePrivateKey, fmt.Sprintf("%s/nodekey", shared.PathSecrets(b.HomeDir())))
	}

	if len(node.Spec.StaticNodes) != 0 {
		args = append(args, BesuStaticNodesFile, fmt.Sprintf("%s/static-nodes.json", shared.PathConfig(b.HomeDir())))
	}

	if len(node.Spec.Bootnodes) != 0 {
		bootnodes := []string{}
		for _, bootnode := range node.Spec.Bootnodes {
			bootnodes = append(bootnodes, string(bootnode))
		}
		args = append(args, BesuBootnodes, strings.Join(bootnodes, ","))
	}

	// public network
	if node.Spec.Genesis == nil {
		args = append(args, BesuNetwork, node.Spec.Network)
	} else { // private network
		args = append(args, BesuGenesisFile, fmt.Sprintf("%s/genesis.json", shared.PathConfig(b.HomeDir())))
		args = append(args, BesuNetworkID, fmt.Sprintf("%d", node.Spec.Genesis.NetworkID))
		args = append(args, BesuDiscoveryEnabled, "false")
	}

	if node.Spec.Miner {
		args = append(args, BesuMinerEnabled)
		args = append(args, BesuMinerCoinbase, string(node.Spec.Coinbase))
	}

	// convert spec rpc modules into format suitable for cli option
	normalizedAPIs := func(modules []ethereumv1alpha1.API) string {
		apis := []string{}
		for _, api := range modules {
			apis = append(apis, strings.ToUpper(string(api)))
		}
		return strings.Join(apis, ",")
	}

	if node.Spec.RPC {
		args = append(args, BesuRPCHTTPEnabled)
		args = append(args, BesuRPCHTTPHost, shared.Host(node.Spec.RPC))
		args = append(args, BesuRPCHTTPPort, fmt.Sprintf("%d", node.Spec.RPCPort))
		args = append(args, BesuRPCHTTPAPI, normalizedAPIs(node.Spec.RPCAPI))
	}

	if node.Spec.Engine {
		args = append(args, BesuEngineRpcEnabled)
		args = append(args, BesuEngineRpcPort, fmt.Sprintf("%d", node.Spec.EnginePort))
		jwtSecretPath := fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(b.HomeDir()))
		args = append(args, BesuEngineJwtSecret, jwtSecretPath)
	}

	if node.Spec.WS {
		args = append(args, BesuRPCWSEnabled)
		args = append(args, BesuRPCWSHost, shared.Host(node.Spec.WS))
		args = append(args, BesuRPCWSPort, fmt.Sprintf("%d", node.Spec.WSPort))
		args = append(args, BesuRPCWSAPI, normalizedAPIs(node.Spec.WSAPI))
	}

	if node.Spec.GraphQL {
		args = append(args, BesuGraphQLHTTPEnabled)
		args = append(args, BesuGraphQLHTTPHost, shared.Host(node.Spec.GraphQL))
		args = append(args, BesuGraphQLHTTPPort, fmt.Sprintf("%d", node.Spec.GraphQLPort))
	}

	if len(node.Spec.Hosts) != 0 {
		commaSeperatedHosts := strings.Join(node.Spec.Hosts, ",")
		args = append(args, BesuHostAllowlist, commaSeperatedHosts)
		if node.Spec.Engine {
			args = append(args, BesuEngineHostAllowList, commaSeperatedHosts)
		}
	}

	if len(node.Spec.CORSDomains) != 0 {
		commaSeperatedDomains := strings.Join(node.Spec.CORSDomains, ",")
		if node.Spec.RPC {
			args = append(args, BesuRPCHTTPCorsOrigins, commaSeperatedDomains)
		}
		if node.Spec.GraphQL {
			args = append(args, BesuGraphQLHTTPCorsOrigins, commaSeperatedDomains)
		}
		// no ws cors setting
	}

	return args
}

// Genesis returns genesis config parameter
func (b *BesuClient) Genesis() (content string, err error) {
	node := b.node
	genesis := node.Spec.Genesis
	mixHash := genesis.MixHash
	nonce := genesis.Nonce
	extraData := "0x00"
	difficulty := genesis.Difficulty
	result := map[string]interface{}{}

	var consensusConfig map[string]uint
	var engine string

	// ethash PoW settings
	if genesis.Ethash != nil {
		consensusConfig = map[string]uint{}

		if genesis.Ethash.FixedDifficulty != nil {
			consensusConfig["fixeddifficulty"] = *genesis.Ethash.FixedDifficulty
		}

		engine = "ethash"
	}

	// clique PoA settings
	if genesis.Clique != nil {
		consensusConfig = map[string]uint{
			"blockperiodseconds": genesis.Clique.BlockPeriod,
			"epochlength":        genesis.Clique.EpochLength,
		}
		engine = "clique"
		extraData = createExtraDataFromSigners(genesis.Clique.Signers)
	}

	// clique ibft2 settings
	if genesis.IBFT2 != nil {

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
		"berlinBlock":         genesis.Forks.Berlin,
		"londonBlock":         genesis.Forks.London,
		"arrowGlacierBlock":   genesis.Forks.ArrowGlacier,
		engine:                consensusConfig,
	}

	if genesis.Forks.DAO != nil {
		config["daoForkBlock"] = genesis.Forks.DAO
	}

	// If london fork is activated at genesis block
	// set baseFeePerGas to 0x3B9ACA00
	// https://discord.com/channels/697535391594446898/743193040197386451/900791897700859916
	if genesis.Forks.London == 0 {
		result["baseFeePerGas"] = "0x3B9ACA00"
	}

	result["config"] = config
	result["nonce"] = nonce
	result["timestamp"] = genesis.Timestamp
	result["gasLimit"] = genesis.GasLimit
	result["difficulty"] = difficulty
	result["coinbase"] = genesis.Coinbase
	result["mixHash"] = mixHash
	result["extraData"] = extraData

	alloc := genesisAccounts(false, genesis.Forks)
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

// EncodeStaticNodes returns the static nodes, one per line
func (b *BesuClient) EncodeStaticNodes() string {

	if len(b.node.Spec.StaticNodes) == 0 {
		return "[]"
	}

	encoded, _ := json.Marshal(b.node.Spec.StaticNodes)
	return string(encoded)
}
