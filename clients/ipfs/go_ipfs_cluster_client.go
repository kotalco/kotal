package ipfs

import (
	"strings"

	ipfsv1alpha1 "github.com/kotalco/kotal/apis/ipfs/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// GoIPFSClusterClient is ipfs cluster service client
// https://github.com/ipfs/ipfs-cluster
type GoIPFSClusterClient struct {
	peer *ipfsv1alpha1.ClusterPeer
}

const (
	//  GoIPFSClusterHomeDir is go ipfs cluster image home dir
	GoIPFSClusterHomeDir = "/home/ipfs-cluster"
)

// Command returns go ipfs cluster entrypoint
func (c *GoIPFSClusterClient) Command() []string {
	return []string{"ipfs-cluster-service"}
}

// Command returns environment variables for the client
func (c *GoIPFSClusterClient) Env() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  EnvIPFSClusterPath,
			Value: shared.PathData(c.HomeDir()),
		},
		{
			Name:  EnvIPFSClusterPeerName,
			Value: c.peer.Name,
		},
		{
			Name:  EnvIPFSLogging,
			Value: string(c.peer.Spec.Logging),
		},
	}
}

// Arg returns go ipfs cluster arguments
func (c *GoIPFSClusterClient) Args() (args []string) {
	args = append(args, GoIPFSClusterDaemonArg)

	if len(c.peer.Spec.BootstrapPeers) != 0 {
		args = append(args, GoIPFSClusterBootstrapArg, strings.Join(c.peer.Spec.BootstrapPeers, ","))
	}

	return
}

// HomeDir returns go ipfs cluster image home directory
func (c *GoIPFSClusterClient) HomeDir() string {
	return GoIPFSClusterHomeDir
}
