package controllers

import (
	"os"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
)

// GoIPFSClient is go-ipfs client
// https://github.com/ipfs/go-ipfs
type GoIPFSClient struct{}

// Images
const (
	// EnvGoIPFSImage is the environment variable used for go ipfs client image
	EnvGoIPFSImage = "GO_IPFS_IMAGE"
	// DefaultGoIPFSImage is the default go ipfs client image
	DefaultGoIPFSImage = "ipfs/go-ipfs:v0.8.0"
	//  GoIPFSHomeDir is go ipfs image home dir
	GoIPFSHomeDir = "/root"
)

// Image returns go-ipfs image
func (c *GoIPFSClient) Image() string {
	if os.Getenv(EnvGoIPFSImage) == "" {
		return DefaultGoIPFSImage
	}
	return os.Getenv(EnvGoIPFSImage)
}

// Command is go-ipfs entrypoint
func (c *GoIPFSClient) Command() []string {
	return []string{"ipfs"}
}

// Args returns go-ipfs args
func (c *GoIPFSClient) Args(peer *ipfsv1alpha1.Peer) (args []string) {

	args = append(args, GoIPFSDaemonArg)

	args = append(args, GoIPFSRoutingArg, string(peer.Spec.Routing))

	return
}

func (c *GoIPFSClient) HomeDir() string {
	return GoIPFSHomeDir
}
