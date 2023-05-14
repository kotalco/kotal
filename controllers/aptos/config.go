package controllers

import (
	"context"
	"fmt"

	aptosv1alpha1 "github.com/kotalco/kotal/apis/aptos/v1alpha1"
	aptosClients "github.com/kotalco/kotal/clients/aptos"
	"github.com/kotalco/kotal/controllers/shared"
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Waypoint struct {
	FromConfig string `yaml:"from_config,omitempty"`
	FromFile   string `yaml:"from_file,omitempty"`
}

type Execution struct {
	GenesisFileLocation string `yaml:"genesis_file_location"`
}

type Base struct {
	Role     string   `yaml:"role"`
	DataDir  string   `yaml:"data_dir"`
	Waypoint Waypoint `yaml:"waypoint"`
}

type Identity struct {
	Type   string `yaml:"type"`
	Key    string `yaml:"key"`
	PeerId string `yaml:"peer_id"`
}

type Peer struct {
	Addresses []string `yaml:"addresses"`
	Role      string   `yaml:"role"`
}

type Network struct {
	NetworkId       string          `yaml:"network_id"`
	DiscoveryMethod string          `yaml:"discovery_method"`
	ListenAddress   string          `yaml:"listen_address"`
	Identity        Identity        `yaml:"identity,omitempty"`
	Seeds           map[string]Peer `yaml:"seeds,omitempty"`
}

type API struct {
	Enabled bool   `yaml:"enabled"`
	Address string `yaml:"address"`
}

type InspectionService struct {
	Port uint `yaml:"port"`
}

type Config struct {
	Base              Base              `yaml:"base"`
	Execution         Execution         `yaml:"execution"`
	FullNodeNetworks  []Network         `yaml:"full_node_networks,omitempty"`
	API               API               `yaml:"api"`
	InspectionService InspectionService `yaml:"inspection_service"`
}

// ConfigFromSpec generates config.toml file from node spec
func ConfigFromSpec(node *aptosv1alpha1.Node, client client.Client) (config string, err error) {
	var role string
	if node.Spec.Validator {
		role = "validator"
	} else {
		role = "full_node"
	}

	var nodePrivateKey string
	var identity Identity
	if node.Spec.NodePrivateKeySecretName != "" {
		key := types.NamespacedName{
			Name:      node.Spec.NodePrivateKeySecretName,
			Namespace: node.Namespace,
		}

		if nodePrivateKey, err = shared.GetSecret(context.Background(), client, key, "key"); err != nil {
			return
		}

		identity = Identity{
			Type:   "from_config",
			Key:    nodePrivateKey,
			PeerId: node.Spec.PeerId,
		}

	}

	seeds := map[string]Peer{}

	if len(node.Spec.SeedPeers) != 0 {
		for _, peer := range node.Spec.SeedPeers {
			seeds[peer.ID] = Peer{
				Addresses: peer.Addresses,
				Role:      "Upstream",
			}
		}
	}

	homeDir := aptosClients.NewClient(node).HomeDir()
	configDir := shared.PathConfig(homeDir)
	dataDir := shared.PathData(homeDir)

	waypoint := Waypoint{}
	var genesisFileLocation string

	if node.Spec.Waypoint == "" {
		waypoint.FromFile = fmt.Sprintf("%s/kotal_waypoint.txt", shared.PathData(homeDir))
	} else {
		waypoint.FromConfig = node.Spec.Waypoint
	}

	if node.Spec.GenesisConfigmapName == "" {
		genesisFileLocation = fmt.Sprintf("%s/kotal_genesis.blob", dataDir)
	} else {
		genesisFileLocation = fmt.Sprintf("%s/genesis.blob", configDir)
	}

	c := Config{
		Base: Base{
			Role:     role,
			DataDir:  shared.PathData(homeDir),
			Waypoint: waypoint,
		},
		Execution: Execution{
			GenesisFileLocation: genesisFileLocation,
		},
		FullNodeNetworks: []Network{
			{
				NetworkId:       "public",
				DiscoveryMethod: "onchain",
				ListenAddress:   fmt.Sprintf("/ip4/%s/tcp/%d", shared.Host(true), node.Spec.P2PPort),
				Identity:        identity,
				Seeds:           seeds,
			},
		},
		API: API{
			Enabled: node.Spec.API,
			Address: fmt.Sprintf("%s:%d", shared.Host(node.Spec.API), node.Spec.APIPort),
		},
		InspectionService: InspectionService{
			Port: node.Spec.MetricsPort,
		},
	}

	data, err := yaml.Marshal(&c)
	if err != nil {
		return
	}

	config = string(data)
	return
}
