package controllers

import ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"

// IPFSClient is IPFS client
type IPFSClient interface {
	Image() string
	Command() []string
	Args(*ipfsv1alpha1.Peer) []string
	HomeDir() string
}

// NewIPFSClient creates new ipfs client
func NewIPFSClient() IPFSClient {
	// TODO: update after multi-client support
	return &GoIPFSClient{}
}
