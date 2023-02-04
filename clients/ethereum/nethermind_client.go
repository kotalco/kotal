package ethereum

import (
	"encoding/json"
	"fmt"
	"strings"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

const (
	// NethermindHomeDir is nethermind docker image home directory
	NethermindHomeDir = "/home/nethermind"
)

// NethermindClient is nethermind client
// https://github.com/NethermindEth/nethermind
type NethermindClient struct {
	*ParityGenesis
	node *ethereumv1alpha1.Node
}

// HomeDir returns besu client home directory
func (n *NethermindClient) HomeDir() string {
	return NethermindHomeDir
}

func (n *NethermindClient) Command() []string {
	return nil
}

func (n *NethermindClient) Env() []corev1.EnvVar {
	return nil
}

// Args returns command line arguments required for client run
// NOTE:
// - Network ID can be set in genesis config
func (n *NethermindClient) Args() (args []string) {

	node := n.node

	args = append(args, NethermindDataPath, shared.PathData(n.HomeDir()))
	args = append(args, NethermindP2PPort, fmt.Sprintf("%d", node.Spec.P2PPort))
	args = append(args, NethermindLogging, strings.ToUpper(string(node.Spec.Logging)))

	if node.Spec.NodePrivateKeySecretName != "" {
		// use enode private key in binary format
		// that has been converted using nethermind_convert_enode_privatekey.sh script
		args = append(args, NethermindNodePrivateKey, fmt.Sprintf("%s/kotal_nodekey", shared.PathData(n.HomeDir())))
	}

	if len(node.Spec.StaticNodes) != 0 {
		args = append(args, NethermindStaticNodesFile, fmt.Sprintf("%s/static-nodes.json", shared.PathConfig(n.HomeDir())))
	}

	if len(node.Spec.Bootnodes) != 0 {
		bootnodes := []string{}
		for _, bootnode := range node.Spec.Bootnodes {
			bootnodes = append(bootnodes, string(bootnode))
		}
		args = append(args, NethermindBootnodes, strings.Join(bootnodes, ","))
	}

	if node.Spec.Genesis == nil {
		args = append(args, NethermindNetwork, node.Spec.Network)
	} else {
		// use empty config, because nethermind uses mainnet.cfg by default which can shadow some settings here
		args = append(args, NethermindNetwork, fmt.Sprintf("%s/empty.cfg", shared.PathConfig(n.HomeDir())))
		args = append(args, NethermindGenesisFile, fmt.Sprintf("%s/genesis.json", shared.PathConfig(n.HomeDir())))
		args = append(args, NethermindDiscoveryEnabled, "false")
	}

	switch node.Spec.SyncMode {
	case ethereumv1alpha1.FullSynchronization:
		args = append(args, NethermindFastSync, "false")
		args = append(args, NethermindFastBlocks, "false")
		args = append(args, NethermindDownloadBodiesInFastSync, "false")
		args = append(args, NethermindDownloadReceiptsInFastSync, "false")
	case ethereumv1alpha1.FastSynchronization:
		args = append(args, NethermindFastSync, "true")
		args = append(args, NethermindFastBlocks, "true")
		args = append(args, NethermindDownloadBodiesInFastSync, "true")
		args = append(args, NethermindDownloadReceiptsInFastSync, "true")
	}

	if node.Spec.Miner {
		args = append(args, NethermindMiningEnabled, "true")
		args = append(args, NethermindMinerCoinbase, string(node.Spec.Coinbase))
		args = append(args, NethermindUnlockAccounts, fmt.Sprintf("[%s]", node.Spec.Coinbase))
		args = append(args, NethermindPasswordFiles, fmt.Sprintf("[%s/account.password]", shared.PathSecrets(n.HomeDir())))
	}

	if node.Spec.RPC {
		args = append(args, NethermindRPCHTTPEnabled, "true")
		args = append(args, NethermindRPCHTTPPort, fmt.Sprintf("%d", node.Spec.RPCPort))
		args = append(args, NethermindRPCHTTPHost, shared.Host(node.Spec.RPC))
		// JSON-RPC API
		apis := []string{}
		for _, api := range node.Spec.RPCAPI {
			apis = append(apis, string(api))
		}
		commaSeperatedAPIs := strings.Join(apis, ",")
		args = append(args, NethermindRPCHTTPAPI, commaSeperatedAPIs)
	}

	if node.Spec.Engine {
		args = append(args, NethermindRPCEnginePort, fmt.Sprintf("%d", node.Spec.EnginePort))
		jwtSecretPath := fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(n.HomeDir()))
		args = append(args, NethermindRPCJwtSecretFile, jwtSecretPath)
	}
	args = append(args, NethermindRPCEngineHost, shared.Host(node.Spec.Engine))

	if node.Spec.WS {
		args = append(args, NethermindRPCWSEnabled, "true")
		args = append(args, NethermindRPCWSPort, fmt.Sprintf("%d", node.Spec.WSPort))
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
