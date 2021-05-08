package controllers

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// GethClient is Go-Ethereum client
type GethClient struct {
	node *ethereumv1alpha1.Node
}

const (
	// EnvGethImage is the environment variable used for go ethereum image
	EnvGethImage = "GETH_IMAGE"
	// DefaultGethImage is go-ethereum image
	DefaultGethImage = "ethereum/client-go:v1.10.2"
	// GethHomeDir is go-ethereum docker image home directory
	GethHomeDir = "/root"
)

// HomeDir returns go-ethereum docker image home directory
func (g *GethClient) HomeDir() string {
	return GethHomeDir
}

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
func (g *GethClient) Args() (args []string) {

	node := g.node

	// appendArg appends argument with optional value to the arguments array
	appendArg := func(arg ...string) {
		args = append(args, arg...)
	}

	appendArg(GethLogging, g.LoggingArgFromVerbosity(node.Spec.Logging))

	// config.toml holding static nodes
	appendArg(GethConfig, fmt.Sprintf("%s/config.toml", shared.PathConfig(g.HomeDir())))

	if node.Spec.ID != 0 {
		appendArg(GethNetworkID, fmt.Sprintf("%d", node.Spec.ID))
	}

	if node.Spec.Nodekey != "" {
		appendArg(GethNodeKey, fmt.Sprintf("%s/nodekey", shared.PathSecrets(g.HomeDir())))
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

	appendArg(GethDataDir, shared.PathData(g.HomeDir()))

	if node.Spec.Join != "" {
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
		appendArg(GethPassword, fmt.Sprintf("%s/account.password", shared.PathSecrets(g.HomeDir())))
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
func (g *GethClient) EncodeStaticNodes() string {

	var encoded []byte

	if len(g.node.Spec.StaticNodes) == 0 {
		encoded = []byte("[]")
	} else {
		encoded, _ = json.Marshal(g.node.Spec.StaticNodes)
	}

	return fmt.Sprintf("[Node.P2P]\nStaticNodes = %s", string(encoded))
}

// Genesis returns genesis config parameter
func (g *GethClient) Genesis() (content string, err error) {
	node := g.node
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

// Image returns geth docker image
func (g *GethClient) Image() string {
	if os.Getenv(EnvGethImage) == "" {
		return DefaultGethImage
	}
	return os.Getenv(EnvGethImage)
}
