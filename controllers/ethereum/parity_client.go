package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// ParityClient is Go-Ethereum client
type ParityClient struct {
	node *ethereumv1alpha1.Node
}

const (
	// EnvParityImage is the environment variable used for parity (OpenEthereum)
	EnvParityImage = "PARITY_IMAGE"
	// DefaultParityImage is parity image
	DefaultParityImage = "openethereum/openethereum:v3.2.4"
	// ParityHomeDir is parity docker image home directory
	ParityHomeDir = "/home/openethereum"
)

// HomeDir returns parity docker image home directory
func (p *ParityClient) HomeDir() string {
	return ParityHomeDir
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

// Args returns command line arguments required for client run
func (p *ParityClient) Args() (args []string) {
	node := p.node

	// appendArg appends argument with optional value to the arguments array
	appendArg := func(arg ...string) {
		args = append(args, arg...)
	}

	appendArg(ParityLogging, p.LoggingArgFromVerbosity(node.Spec.Logging))

	if node.Spec.ID != 0 {
		appendArg(ParityNetworkID, fmt.Sprintf("%d", node.Spec.ID))
	}

	if node.Spec.NodekeySecretName != "" {
		appendArg(ParityNodeKey, fmt.Sprintf("%s/nodekey", shared.PathSecrets(p.HomeDir())))
	}

	if len(node.Spec.Bootnodes) != 0 {
		bootnodes := []string{}
		for _, bootnode := range node.Spec.Bootnodes {
			bootnodes = append(bootnodes, string(bootnode))
		}
		appendArg(ParityBootnodes, strings.Join(bootnodes, ","))
	}

	appendArg(ParityDataDir, shared.PathData(p.HomeDir()))

	appendArg(ParityReservedPeers, fmt.Sprintf("%s/static-nodes", shared.PathConfig(p.HomeDir())))

	if node.Spec.Genesis == nil {
		if node.Spec.Network != ethereumv1alpha1.MainNetwork {
			appendArg(ParityNetwork, node.Spec.Network)
		}
	} else {
		appendArg(ParityNetwork, fmt.Sprintf("%s/genesis.json", shared.PathConfig(p.HomeDir())))
		appendArg(ParityNoDiscovery)
	}

	if node.Spec.P2PPort != 0 {
		appendArg(ParityP2PPort, fmt.Sprintf("%d", node.Spec.P2PPort))
	}

	if node.Spec.SyncMode != "" {
		appendArg(ParitySyncMode, p.PrunningArgFromSyncMode(node.Spec.SyncMode))
	}

	if node.Spec.Coinbase != "" {
		appendArg(ParityMinerCoinbase, string(node.Spec.Coinbase))
		appendArg(ParityUnlock, string(node.Spec.Coinbase))
		appendArg(ParityPassword, fmt.Sprintf("%s/account.password", shared.PathSecrets(p.HomeDir())))
		if node.Spec.Consensus == ethereumv1alpha1.ProofOfAuthority {
			appendArg(ParityEngineSigner, string(node.Spec.Coinbase))
		}
	}

	if !node.Spec.RPC {
		appendArg(ParityDisableRPC)
	}

	if node.Spec.RPCPort != 0 {
		appendArg(ParityRPCHTTPPort, fmt.Sprintf("%d", node.Spec.RPCPort))
	}

	appendArg(ParityRPCHTTPHost, DefaultHost)

	if len(node.Spec.RPCAPI) != 0 {
		apis := []string{}
		for _, api := range node.Spec.RPCAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(ParityRPCHTTPAPI, commaSeperatedAPIs)
	}

	if !node.Spec.WS {
		appendArg(ParityDisableWS)
	}

	if node.Spec.WSPort != 0 {
		appendArg(ParityRPCWSPort, fmt.Sprintf("%d", node.Spec.WSPort))
	}

	appendArg(ParityRPCWSHost, DefaultHost)

	if len(node.Spec.WSAPI) != 0 {
		apis := []string{}
		for _, api := range node.Spec.WSAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(ParityRPCWSAPI, commaSeperatedAPIs)
	}

	if len(node.Spec.Hosts) != 0 {
		commaSeperatedHosts := strings.Join(node.Spec.Hosts, ",")
		if node.Spec.RPC {
			appendArg(ParityRPCHostWhitelist, commaSeperatedHosts)
		}
		if node.Spec.WS {
			appendArg(ParityRPCWSWhitelist, commaSeperatedHosts)
		}
	}

	if len(node.Spec.CORSDomains) != 0 {
		commaSeperatedDomains := strings.Join(node.Spec.CORSDomains, ",")
		if node.Spec.RPC {
			appendArg(ParityRPCHTTPCorsOrigins, commaSeperatedDomains)
		}
		if node.Spec.WS {
			appendArg(ParityRPCWSCorsOrigins, commaSeperatedDomains)
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

// Genesis returns genesis config parameter
func (p *ParityClient) Genesis() (content string, err error) {
	node := p.node
	genesis := node.Spec.Genesis
	consensus := node.Spec.Consensus
	extraData := "0x00"
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

	hex := func(n uint) string {
		return fmt.Sprintf("%#x", n)
	}

	tingerineWhistleBlock := hex(genesis.Forks.EIP150)
	spuriousDragonBlock := hex(genesis.Forks.EIP155)
	homesteadBlock := hex(genesis.Forks.Homestead)
	byzantiumBlock := hex(genesis.Forks.Byzantium)
	constantinopleBlock := hex(genesis.Forks.Constantinople)
	petersburgBlock := hex(genesis.Forks.Petersburg)
	istanbulBlock := hex(genesis.Forks.Istanbul)
	muirGlacierBlock := hex(genesis.Forks.MuirGlacier)
	berlinBlock := hex(genesis.Forks.Berlin)
	londonBlock := hex(genesis.Forks.London)

	// ethash PoW settings
	if consensus == ethereumv1alpha1.ProofOfWork {
		params := map[string]interface{}{
			"minimumDifficulty":      "0x020000",
			"difficultyBoundDivisor": "0x0800",
			"durationLimit":          "0x0d",
			"blockReward": map[string]string{
				tingerineWhistleBlock: "0x4563918244f40000",
				byzantiumBlock:        "0x29a2241af62c0000",
				constantinopleBlock:   "0x1bc16d674ec80000",
			},
			"homesteadTransition": homesteadBlock,
			"eip100bTransition":   byzantiumBlock,
			"difficultyBombDelays": map[string]string{
				byzantiumBlock:      "0x2dc6c0",
				constantinopleBlock: "0x1e8480",
				muirGlacierBlock:    "0x3d0900",
				londonBlock:         "0xaae60",
			},
		}

		if genesis.Forks.DAO != nil {
			params["daoHardforkTransition"] = hex(*genesis.Forks.DAO)
		}

		engineConfig = map[string]interface{}{
			"Ethash": map[string]interface{}{
				"params": params,
			},
		}
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

	paramsConfig := map[string]interface{}{
		// other non fork parameters
		"chainID":              hex(genesis.ChainID),
		"accountStartNonce":    "0x00",
		"gasLimitBoundDivisor": "0x0400",
		"maximumExtraDataSize": "0xffff",
		"minGasLimit":          "0x1388",
		"networkID":            hex(node.Spec.ID),
		// Tingerine Whistle
		"eip150Transition": tingerineWhistleBlock,
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
		"eip1283DisableTransition": petersburgBlock,
		// Istanbul
		"eip1283ReenableTransition": istanbulBlock,
		"eip1344Transition":         istanbulBlock,
		"eip1706Transition":         istanbulBlock,
		"eip1884Transition":         istanbulBlock,
		"eip2028Transition":         istanbulBlock,
		// Berlin
		"eip2315Transition": berlinBlock, // Simple Subroutines for the EVM
		"eip2929Transition": berlinBlock, // Gas cost increases for state access opcodes
		"eip2930Transition": berlinBlock, // Access lists. Requires eips 2718 (Typed Transaction Envelope), and 2929
		// London
		"eip1559Transition":                  londonBlock, // Fee market
		"eip3198Transition":                  londonBlock, // BASEFEE opcode
		"eip3541Transition":                  londonBlock, // Reject new contracts starting with the 0xEF byte
		"eip3529Transition":                  londonBlock, // Reduction in refunds
		"eip1559BaseFeeMaxChangeDenominator": "0x8",
		"eip1559ElasticityMultiplier":        "0x2",
		"eip1559BaseFeeInitialValue":         "0x3B9ACA00",
	}

	alloc := genesisAccounts(true, genesis.Forks)
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

// EncodeStaticNodes returns the static nodes, one per line
func (p *ParityClient) EncodeStaticNodes() string {
	nodes := []string{}
	for _, s := range p.node.Spec.StaticNodes {
		nodes = append(nodes, string(s))
	}
	return strings.Join(nodes, "\n")
}

// Image returns parity docker image
func (p *ParityClient) Image() string {
	if os.Getenv(EnvParityImage) == "" {
		return DefaultParityImage
	}
	return os.Getenv(EnvParityImage)
}

// KeyStoreFromPrivatekey generates key store from private key (hex without 0x)
func KeyStoreFromPrivatekey(key, password string) (content []byte, err error) {
	dir, err := ioutil.TempDir(os.TempDir(), "tmp")
	if err != nil {
		return
	}
	defer os.RemoveAll(dir)

	ks := keystore.NewKeyStore(dir, keystore.StandardScryptN, keystore.StandardScryptP)
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return
	}

	acc, err := ks.ImportECDSA(privateKey, password)
	if err != nil {
		return
	}

	content, err = ks.Export(acc, password, password)

	return
}
