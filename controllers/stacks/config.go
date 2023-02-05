package controllers

import (
	"bytes"
	"context"
	"fmt"

	"github.com/BurntSushi/toml"
	stacksv1alpha1 "github.com/kotalco/kotal/apis/stacks/v1alpha1"
	stacksClients "github.com/kotalco/kotal/clients/stacks"
	"github.com/kotalco/kotal/controllers/shared"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BurnChain struct {
	Chain    string `toml:"chain"`
	Mode     string `toml:"mode"`
	PeerHost string `toml:"peer_host"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	RPCPort  uint   `toml:"rpc_port"`
	PeerPort uint   `toml:"peer_port"`
}

type Node struct {
	WorkingDir      string `toml:"working_dir"`
	RPCBind         string `toml:"rpc_bind"`
	P2PBind         string `toml:"p2p_bind"`
	Seed            string `toml:"seed,omitempty"`
	LocalPeerSeed   string `toml:"local_peer_seed"`
	Miner           bool   `toml:"miner"`
	MineMicroblocks bool   `toml:"mine_microblocks,omitempty"`
}

type Config struct {
	Node      Node      `toml:"node"`
	BurnChain BurnChain `toml:"burnchain"`
}

// ConfigFromSpec generates config.toml file from node spec
func ConfigFromSpec(node *stacksv1alpha1.Node, client client.Client) (config string, err error) {
	c := &Config{}

	c.Node = Node{
		WorkingDir: shared.PathData(stacksClients.StacksNodeHomeDir),
		RPCBind:    fmt.Sprintf("%s:%d", shared.Host(node.Spec.RPC), node.Spec.RPCPort),
		P2PBind:    fmt.Sprintf("%s:%d", shared.Host(true), node.Spec.P2PPort),
		Miner:      node.Spec.Miner,
	}

	if node.Spec.Miner {
		var seedPrivateKey string
		name := types.NamespacedName{
			Name:      node.Spec.SeedPrivateKeySecretName,
			Namespace: node.Namespace,
		}
		seedPrivateKey, err = shared.GetSecret(context.Background(), client, name, "key")
		if err != nil {
			return
		}

		c.Node.Seed = seedPrivateKey
		c.Node.MineMicroblocks = node.Spec.MineMicroblocks
	}

	if node.Spec.NodePrivateKeySecretName != "" {
		var nodePrivateKey string
		name := types.NamespacedName{
			Name:      node.Spec.NodePrivateKeySecretName,
			Namespace: node.Namespace,
		}
		nodePrivateKey, err = shared.GetSecret(context.Background(), client, name, "key")
		if err != nil {
			return
		}

		c.Node.LocalPeerSeed = nodePrivateKey
	}

	name := types.NamespacedName{
		Name:      node.Spec.BitcoinNode.RpcPasswordSecretName,
		Namespace: node.Namespace,
	}
	password, err := shared.GetSecret(context.Background(), client, name, "password")
	if err != nil {
		return
	}

	c.BurnChain = BurnChain{
		Chain:    "bitcoin",
		Mode:     string(node.Spec.Network),
		PeerHost: node.Spec.BitcoinNode.Endpoint,
		Username: node.Spec.BitcoinNode.RpcUsername,
		Password: password,
		RPCPort:  node.Spec.BitcoinNode.RpcPort,
		PeerPort: node.Spec.BitcoinNode.P2pPort,
	}

	var buff bytes.Buffer
	enc := toml.NewEncoder(&buff)
	err = enc.Encode(c)
	if err != nil {
		return
	}

	config = buff.String()

	return
}
