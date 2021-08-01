package ethereum

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

const (
	// EnvNethermindImage is the environment variable used for nethermind image
	EnvNethermindImage = "NETHERMIND_IMAGE"
	// DefaultNethermindImage is nethermind image
	DefaultNethermindImage = "kotalco/nethermind:v1.10.79"
	// NethermindHomeDir is nethermind docker image home directory
	NethermindHomeDir = "/home/nethermind"
)

// NethermindClient is nethermind client
type NethermindClient struct {
	*ParityGenesis
	node *ethereumv1alpha1.Node
}

// LoggingArgFromVerbosity returns logging argument from node verbosity level
// Nethermind supports TRACE, DEBUG, INFO, WARN, ERROR verbosity levels
func (n *NethermindClient) LoggingArgFromVerbosity(level ethereumv1alpha1.VerbosityLevel) string {
	return strings.ToUpper(string(level))
}

// HomeDir returns besu client home directory
func (n *NethermindClient) HomeDir() string {
	return NethermindHomeDir
}

// Args returns command line arguments required for client run
// NOTE:
// - Network ID can be set in genesis config
// - Bootnodes can be set in genesis config
// TODO:
// - followup adding --bootnodes cli flag https://github.com/NethermindEth/nethermind/issues/3185
func (n *NethermindClient) Args() (args []string) {

	node := n.node

	// appendArg appends argument with optional value to the arguments array
	appendArg := func(arg ...string) {
		args = append(args, arg...)
	}

	appendArg(NethermindDataPath, shared.PathData(n.HomeDir()))
	appendArg(NethermindP2PPort, fmt.Sprintf("%d", node.Spec.P2PPort))
	appendArg(NethermindLogging, n.LoggingArgFromVerbosity(node.Spec.Logging))

	if node.Spec.NodePrivatekeySecretName != "" {
		// use enode private key in binary format
		// that has been converted using nethermind_convert_enode_privatekey.sh script
		appendArg(NethermindNodePrivateKey, fmt.Sprintf("%s/kotal_nodekey", shared.PathData(n.HomeDir())))
	}

	if len(node.Spec.StaticNodes) != 0 {
		appendArg(NethermindStaticNodesFile, fmt.Sprintf("%s/static-nodes.json", shared.PathConfig(n.HomeDir())))
	}

	if node.Spec.Genesis == nil {
		appendArg(NethermindNetwork, node.Spec.Network)
	} else {
		// use empty config, because nethermind uses mainnet.cfg by default which can shadow some settings here
		appendArg(NethermindNetwork, fmt.Sprintf("%s/empty.cfg", shared.PathConfig(n.HomeDir())))
		appendArg(NethermindGenesisFile, fmt.Sprintf("%s/genesis.json", shared.PathConfig(n.HomeDir())))
		appendArg(NethermindDiscoveryEnabled, "false")
	}

	switch node.Spec.SyncMode {
	case ethereumv1alpha1.FullSynchronization:
		appendArg(NethermindFastSync, "false")
		appendArg(NethermindFastBlocks, "false")
		appendArg(NethermindDownloadBodiesInFastSync, "false")
		appendArg(NethermindDownloadReceiptsInFastSync, "false")
	case ethereumv1alpha1.FastSynchronization:
		appendArg(NethermindFastSync, "true")
		appendArg(NethermindFastBlocks, "true")
		appendArg(NethermindDownloadBodiesInFastSync, "true")
		appendArg(NethermindDownloadReceiptsInFastSync, "true")
	case ethereumv1alpha1.LightSynchronization:
		appendArg(NethermindFastSync, "true")
		appendArg(NethermindBeamSync, "true")
		appendArg(NethermindFastBlocks, "true")
		appendArg(NethermindDownloadHeadersInFastSync, "false")
		appendArg(NethermindDownloadBodiesInFastSync, "false")
		appendArg(NethermindDownloadReceiptsInFastSync, "false")
	}

	if node.Spec.Miner {
		appendArg(NethermindMiningEnabled, "true")
		appendArg(NethermindMinerCoinbase, string(node.Spec.Coinbase))
		appendArg(NethermindUnlockAccounts, fmt.Sprintf("[%s]", node.Spec.Coinbase))
		appendArg(NethermindPasswordFiles, fmt.Sprintf("[%s/account.password]", shared.PathSecrets(n.HomeDir())))
	}

	if node.Spec.RPC {
		appendArg(NethermindRPCHTTPEnabled, "true")
		appendArg(NethermindRPCHTTPPort, fmt.Sprintf("%d", node.Spec.RPCPort))
		appendArg(NethermindRPCHTTPHost, DefaultHost)
		// JSON-RPC API
		apis := []string{}
		for _, api := range node.Spec.RPCAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		appendArg(NethermindRPCHTTPAPI, commaSeperatedAPIs)
	}

	if node.Spec.WS {
		appendArg(NethermindRPCWSEnabled, "true")
		appendArg(NethermindRPCWSPort, fmt.Sprintf("%d", node.Spec.WSPort))
		// no option for ws host, ws uses same http host as JSON-RPC
		// nethermind ws reuses enabled JSON-RPC modules
	}

	return args
}

// Genesis returns genesis config parameter
func (p *NethermindClient) Genesis() (string, error) {
	return p.ParityGenesis.Genesis(p.node)
}

// EncodeStaticNodes returns the static nodes, one per line
func (n *NethermindClient) EncodeStaticNodes() string {

	if len(n.node.Spec.StaticNodes) == 0 {
		return "[]"
	}

	encoded, _ := json.Marshal(n.node.Spec.StaticNodes)
	return string(encoded)
}

// Image returns nethermind docker image
func (n *NethermindClient) Image() string {
	if os.Getenv(EnvNethermindImage) == "" {
		return DefaultNethermindImage
	}
	return os.Getenv(EnvNethermindImage)
}
