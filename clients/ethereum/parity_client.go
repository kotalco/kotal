package ethereum

import (
	"fmt"
	"os"
	"strings"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// ParityClient is OpenEthereum (previously parity) client
type ParityClient struct {
	*ParityGenesis
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

	appendArg(ParityDataDir, shared.PathData(p.HomeDir()))
	appendArg(ParityP2PPort, fmt.Sprintf("%d", node.Spec.P2PPort))
	appendArg(ParitySyncMode, p.PrunningArgFromSyncMode(node.Spec.SyncMode))
	appendArg(ParityLogging, p.LoggingArgFromVerbosity(node.Spec.Logging))

	if node.Spec.NodePrivateKeySecretName != "" {
		appendArg(ParityNodeKey, fmt.Sprintf("%s/nodekey", shared.PathSecrets(p.HomeDir())))
	}

	if len(node.Spec.Bootnodes) != 0 {
		bootnodes := []string{}
		for _, bootnode := range node.Spec.Bootnodes {
			bootnodes = append(bootnodes, string(bootnode))
		}
		appendArg(ParityBootnodes, strings.Join(bootnodes, ","))
	}

	if len(node.Spec.StaticNodes) != 0 {
		appendArg(ParityReservedPeers, fmt.Sprintf("%s/static-nodes", shared.PathConfig(p.HomeDir())))
	}

	if node.Spec.Genesis == nil {
		if node.Spec.Network != ethereumv1alpha1.MainNetwork {
			appendArg(ParityNetwork, node.Spec.Network)
		}
	} else {
		appendArg(ParityNetwork, fmt.Sprintf("%s/genesis.json", shared.PathConfig(p.HomeDir())))
		appendArg(ParityNetworkID, fmt.Sprintf("%d", node.Spec.Genesis.NetworkID))
		appendArg(ParityNoDiscovery)
	}

	if node.Spec.Coinbase != "" {
		appendArg(ParityMinerCoinbase, string(node.Spec.Coinbase))
		appendArg(ParityUnlock, string(node.Spec.Coinbase))
		appendArg(ParityPassword, fmt.Sprintf("%s/account.password", shared.PathSecrets(p.HomeDir())))
		if node.Spec.Genesis.Clique != nil {
			appendArg(ParityEngineSigner, string(node.Spec.Coinbase))
		}
	}

	if node.Spec.RPC {
		appendArg(ParityRPCHTTPPort, fmt.Sprintf("%d", node.Spec.RPCPort))
		appendArg(ParityRPCHTTPHost, DefaultHost)
		// JSON-RPC API
		apis := []string{}
		for _, api := range node.Spec.RPCAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(ParityRPCHTTPAPI, commaSeperatedAPIs)
	} else {
		appendArg(ParityDisableRPC)
	}

	if node.Spec.WS {
		appendArg(ParityRPCWSPort, fmt.Sprintf("%d", node.Spec.WSPort))
		appendArg(ParityRPCWSHost, DefaultHost)
		// WebSocket API
		apis := []string{}
		for _, api := range node.Spec.WSAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(ParityRPCWSAPI, commaSeperatedAPIs)
	} else {
		appendArg(ParityDisableWS)
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

// Genesis returns genesis config parameter
func (p *ParityClient) Genesis() (string, error) {
	return p.ParityGenesis.Genesis(p.node)
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
