package controllers

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/kotalco/kotal/controllers/shared"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/BurntSushi/toml"
	chainlinkv1alpha1 "github.com/kotalco/kotal/apis/chainlink/v1alpha1"
)

type WebServer struct {
	AllowOrigins  string
	SecureCookies bool
	HTTPPort      uint
	TLS           WebServerTLS
}

type WebServerTLS struct {
	HTTPSPort uint
	CertPath  string
	KeyPath   string
}

type EVM struct {
	ChainID             string
	LinkContractAddress string
	Nodes               []Node
}

type Node struct {
	Name    string
	WSURL   string
	HTTPURL string
}

type P2PV1 struct {
	ListenPort uint
}

type P2P struct {
	V1 P2PV1
}

type Log struct {
	Level string
}

type Config struct {
	RootDir   string
	P2P       P2P
	Log       Log
	WebServer WebServer
	// WebServerTLS WebServerTLS `toml:"WebServer.TLS,omitempty"`
	EVM   []EVM
	Nodes []Node `toml:"EVM.Nodes"`
}

// ConfigFromSpec generates config.toml file from node spec
func ConfigFromSpec(node *chainlinkv1alpha1.Node, homeDir string) (config string, err error) {
	c := &Config{}

	c.RootDir = shared.PathData(homeDir)

	c.Log = Log{
		Level: string(node.Spec.Logging),
	}

	c.EVM = []EVM{
		{
			ChainID:             fmt.Sprintf("%d", node.Spec.EthereumChainId),
			LinkContractAddress: node.Spec.LinkContractAddress,
			Nodes: []Node{
				{
					Name:    "node",
					WSURL:   node.Spec.EthereumWSEndpoint,
					HTTPURL: strings.Join(node.Spec.EthereumHTTPEndpoints, ","),
				},
			},
		},
	}

	c.P2P = P2P{
		V1: P2PV1{
			ListenPort: node.Spec.P2PPort,
		},
	}

	c.WebServer = WebServer{
		AllowOrigins:  strings.Join(node.Spec.CORSDomains, ","),
		SecureCookies: node.Spec.SecureCookies,
		HTTPPort:      node.Spec.APIPort,
	}

	if node.Spec.CertSecretName != "" {
		c.WebServer.TLS = WebServerTLS{
			HTTPSPort: node.Spec.TLSPort,
			KeyPath:   fmt.Sprintf("%s/tls.key", shared.PathSecrets(homeDir)),
			CertPath:  fmt.Sprintf("%s/tls.crt", shared.PathSecrets(homeDir)),
		}
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

type SecretsConfig struct {
	Database Database
	Password Password
}

type Database struct {
	URL string
}

type Password struct {
	Keystore string
}

// SecretsFromSpec generates config.toml file from node spec
func SecretsFromSpec(node *chainlinkv1alpha1.Node, homeDir string, client client.Client) (config string, err error) {
	c := &SecretsConfig{}

	c.Database = Database{
		URL: node.Spec.DatabaseURL,
	}

	key := types.NamespacedName{
		Name:      node.Spec.KeystorePasswordSecretName,
		Namespace: node.Namespace,
	}

	var password string
	if password, err = shared.GetSecret(context.Background(), client, key, "password"); err != nil {
		return
	}

	c.Password = Password{
		Keystore: password,
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
