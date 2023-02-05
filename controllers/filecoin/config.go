package controllers

import (
	"bytes"
	"fmt"

	"github.com/BurntSushi/toml"
	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

type API struct {
	ListenAddress  string
	RequestTimeout string
}

type Backup struct {
	DisableMetadataLog bool
}

type LibP2P struct {
	ListenAddresses []string
}

type Client struct {
	UseIpfs             bool
	IpfsOnlineMode      bool
	IpfsMAddr           string
	IpfsUseForRetrieval bool
}

type Config struct {
	API    *API `toml:"API,omitempty"`
	Backup Backup
	LibP2P LibP2P `toml:"Libp2p"`
	Client *Client
}

// ConfigFromSpec generates config.toml file from node spec
func ConfigFromSpec(node *filecoinv1alpha1.Node) (config string, err error) {
	c := &Config{}

	if node.Spec.API {
		c.API = &API{
			ListenAddress: fmt.Sprintf("/ip4/%s/tcp/%d/http", shared.Host(node.Spec.API), node.Spec.APIPort),
		}
		c.API.RequestTimeout = fmt.Sprintf("%ds", node.Spec.APIRequestTimeout)
	}

	c.Backup.DisableMetadataLog = node.Spec.DisableMetadataLog

	c.LibP2P.ListenAddresses = []string{fmt.Sprintf("/ip4/%s/tcp/%d", shared.Host(true), node.Spec.P2PPort)}

	if node.Spec.IPFSPeerEndpoint != "" {
		c.Client = &Client{
			UseIpfs:             true,
			IpfsMAddr:           node.Spec.IPFSPeerEndpoint,
			IpfsOnlineMode:      node.Spec.IPFSOnlineMode,
			IpfsUseForRetrieval: node.Spec.IPFSForRetrieval,
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
