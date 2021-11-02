package ipfs

import (
	"os"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// GoIPFSClient is go-ipfs client
// https://github.com/ipfs/go-ipfs
type GoIPFSClient struct {
	peer *ipfsv1alpha1.Peer
}

// Images
const (
	// EnvGoIPFSImage is the environment variable used for go ipfs client image
	EnvGoIPFSImage = "GO_IPFS_IMAGE"
	// DefaultGoIPFSImage is the default go ipfs client image
	DefaultGoIPFSImage = "kotalco/go-ipfs:v0.10.0"
	//  GoIPFSHomeDir is go ipfs image home dir
	GoIPFSHomeDir = "/home/ipfs"
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

// Command returns environment variables for the client
func (c *GoIPFSClient) Env() []corev1.EnvVar {
	return nil
}

// Args returns go-ipfs args
func (c *GoIPFSClient) Args() (args []string) {

	peer := c.peer

	args = append(args, GoIPFSDaemonArg)

	args = append(args, GoIPFSRoutingArg, string(peer.Spec.Routing))

	return
}

func (c *GoIPFSClient) HomeDir() string {
	return GoIPFSHomeDir
}
