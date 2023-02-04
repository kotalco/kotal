package ethereum

import (
	"encoding/json"
	"fmt"
	"strings"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// GethClient is Go-Ethereum client
// https://github.com/ethereum/go-ethereum
type GethClient struct {
	node *ethereumv1alpha1.Node
}

const (
	// GethHomeDir is go-ethereum docker image home directory
	GethHomeDir = "/home/ethereum"
)

var (
	verbosityLevels = map[sharedAPI.VerbosityLevel]string{
		sharedAPI.NoLogs:    "0",
		sharedAPI.ErrorLogs: "1",
		sharedAPI.WarnLogs:  "2",
		sharedAPI.InfoLogs:  "3",
		sharedAPI.DebugLogs: "4",
		sharedAPI.AllLogs:   "5",
	}
)

// HomeDir returns go-ethereum docker image home directory
func (g *GethClient) HomeDir() string {
	return GethHomeDir
}

func (g *GethClient) Command() []string {
	return nil
}

func (g *GethClient) Env() []corev1.EnvVar {
	return nil
}

// Args returns command line arguments required for client run
func (g *GethClient) Args() (args []string) {

	node := g.node

	args = append(args, GethDataDir, shared.PathData(g.HomeDir()))
	args = append(args, GethDisableIPC)
	args = append(args, GethP2PPort, fmt.Sprintf("%d", node.Spec.P2PPort))
	args = append(args, GethSyncMode, string(node.Spec.SyncMode))
	if g.node.Spec.SyncMode == ethereumv1alpha1.FullSynchronization {
		args = append(args, GethGcMode, "archive")
		args = append(args, GethTxLookupLimit, "0")
		args = append(args, GethCachePreImages)
	}
	args = append(args, GethLogging, verbosityLevels[node.Spec.Logging])

	// config.toml holding static nodes
	if len(node.Spec.StaticNodes) != 0 {
		args = append(args, GethConfig, fmt.Sprintf("%s/config.toml", shared.PathConfig(g.HomeDir())))
	}

	if node.Spec.NodePrivateKeySecretName != "" {
		args = append(args, GethNodeKey, fmt.Sprintf("%s/nodekey", shared.PathSecrets(g.HomeDir())))
	}

	if len(node.Spec.Bootnodes) != 0 {
		bootnodes := []string{}
		for _, bootnode := range node.Spec.Bootnodes {
			bootnodes = append(bootnodes, string(bootnode))
		}
		args = append(args, GethBootnodes, strings.Join(bootnodes, ","))
	}

	if node.Spec.Genesis == nil {
		args = append(args, fmt.Sprintf("--%s", node.Spec.Network))
	} else {
		args = append(args, GethNoDiscovery)
		args = append(args, GethNetworkID, fmt.Sprintf("%d", node.Spec.Genesis.NetworkID))
	}

	if node.Spec.Miner {
		args = append(args, GethMinerEnabled)
		args = append(args, GethMinerCoinbase, string(node.Spec.Coinbase))
		args = append(args, GethUnlock, string(node.Spec.Coinbase))
		args = append(args, GethPassword, fmt.Sprintf("%s/account.password", shared.PathSecrets(g.HomeDir())))
	}

	if node.Spec.RPC {
		args = append(args, GethRPCHTTPEnabled)
		args = append(args, GethRPCHTTPHost, shared.Host(node.Spec.RPC))
		args = append(args, GethRPCHTTPPort, fmt.Sprintf("%d", node.Spec.RPCPort))
		// JSON-RPC API
		apis := []string{}
		for _, api := range node.Spec.RPCAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		args = append(args, GethRPCHTTPAPI, commaSeperatedAPIs)
	}

	if node.Spec.Engine {
		args = append(args, GethAuthRPCPort, fmt.Sprintf("%d", node.Spec.EnginePort))
		jwtSecretPath := fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(g.HomeDir()))
		args = append(args, GethAuthRPCJwtSecret, jwtSecretPath)
	}
	args = append(args, GethAuthRPCAddress, shared.Host(node.Spec.Engine))

	if node.Spec.WS {
		args = append(args, GethRPCWSEnabled)
		args = append(args, GethRPCWSHost, shared.Host(node.Spec.WS))
		args = append(args, GethRPCWSPort, fmt.Sprintf("%d", node.Spec.WSPort))
		// WebSocket API
		apis := []string{}
		for _, api := range node.Spec.WSAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		args = append(args, GethRPCWSAPI, commaSeperatedAPIs)
	}

	if node.Spec.GraphQL {
		args = append(args, GethGraphQLHTTPEnabled)
		//NOTE: .GraphQLPort is ignored because rpc port will be used by graphql server
		// .GraphQLPort will be used in the service that point to the pod
	}

	if len(node.Spec.Hosts) != 0 {
		commaSeperatedHosts := strings.Join(node.Spec.Hosts, ",")
		if node.Spec.RPC {
			args = append(args, GethRPCHostWhitelist, commaSeperatedHosts)
		}
		if node.Spec.GraphQL {
			args = append(args, GethGraphQLHostWhitelist, commaSeperatedHosts)
		}
		if node.Spec.Engine {
			args = append(args, GethAuthRPCHosts, commaSeperatedHosts)
		}
		// no ws hosts settings
	}

	if len(node.Spec.CORSDomains) != 0 {
		commaSeperatedDomains := strings.Join(node.Spec.CORSDomains, ",")
		if node.Spec.RPC {
			args = append(args, GethRPCHTTPCorsOrigins, commaSeperatedDomains)
		}
		if node.Spec.GraphQL {
			args = append(args, GethGraphQLHTTPCorsOrigins, commaSeperatedDomains)
		}
		if node.Spec.WS {
			args = append(args, GethWSOrigins, commaSeperatedDomains)
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
		engine = "ethash"
	}

	// clique PoA settings
	if genesis.Clique != nil {
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
		"berlinBlock":         genesis.Forks.Berlin,
		"londonBlock":         genesis.Forks.London,
		"arrowGlacierBlock":   genesis.Forks.ArrowGlacier,
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
