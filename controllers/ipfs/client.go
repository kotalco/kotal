package controllers

import (
	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
)

// IPFSClient is IPFS peer client
type IPFSClient interface {
	shared.Client
}

// IPFSClusterClient is IPFS cluster peer client
type IPFSClusterClient interface {
	shared.Client
}

// NewIPFSClient creates new ipfs client
func NewIPFSClient(peer *ipfsv1alpha1.Peer) IPFSClient {
	// TODO: update after multi-client support
	return &GoIPFSClient{peer}
}

// NewIPFSClusterClient creates new ipfs cluster client
func NewIPFSClusterClient(peer *ipfsv1alpha1.ClusterPeer) IPFSClusterClient {
	// TODO: update after multi-client support
	return &GoIPFSClusterClient{peer}
}
