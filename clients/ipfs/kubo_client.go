package ipfs

import (
	"os"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// KuboClient is an ipfs implementation in golang
// https://github.com/ipfs/kubo
type KuboClient struct {
	peer *ipfsv1alpha1.Peer
}

// Images
const (
	// EnvGoIPFSImage is the environment variable used for go ipfs client image
	EnvGoIPFSImage = "GO_IPFS_IMAGE"
	// DefaultGoIPFSImage is the default go ipfs client image
	DefaultGoIPFSImage = "kotalco/kubo:v0.16.0"
	//  GoIPFSHomeDir is go ipfs image home dir
	GoIPFSHomeDir = "/home/ipfs"
)

// Image returns kubo image
func (c *KuboClient) Image() string {
	if img := c.peer.Spec.Image; img != nil {
		return *img
	} else if os.Getenv(EnvGoIPFSImage) == "" {
		return DefaultGoIPFSImage
	}
	return os.Getenv(EnvGoIPFSImage)
}

// Command is kubo entrypoint
func (c *KuboClient) Command() []string {
	return []string{"ipfs"}
}

// Command returns environment variables for the client
func (c *KuboClient) Env() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  EnvIPFSPath,
			Value: shared.PathData(c.HomeDir()),
		},
		{
			Name:  EnvIPFSLogging,
			Value: string(c.peer.Spec.Logging),
		},
	}
}

// Args returns kubo args
func (c *KuboClient) Args() (args []string) {

	peer := c.peer

	args = append(args, GoIPFSDaemonArg)

	args = append(args, GoIPFSRoutingArg, string(peer.Spec.Routing))

	return
}

func (c *KuboClient) HomeDir() string {
	return GoIPFSHomeDir
}
