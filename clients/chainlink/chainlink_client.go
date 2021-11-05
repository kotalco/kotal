package chainlink

import (
	"fmt"
	"os"

	chainlinkv1alpha1 "github.com/kotalco/kotal/apis/chainlink/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// ChainlinkClient is chainlink official client
// https://github.com/smartcontractkit/chainlink
type ChainlinkClient struct {
	node *chainlinkv1alpha1.Node
}

// Images
const (
	// EnvChainlinkImage is the environment variable used for chainlink client image
	EnvChainlinkImage = "CHAINLINK_IMAGE"
	// DefaultChainlinkImage is the default chainlink client image
	DefaultChainlinkImage = "smartcontract/chainlink:1.0.0"
	// ChainlinkHomeDir is chainlink image home dir
	// TODO: update the home directory
	ChainlinkHomeDir = "/"
)

// Image returns chainlink image
func (c *ChainlinkClient) Image() string {
	if os.Getenv(EnvChainlinkImage) == "" {
		return DefaultChainlinkImage
	}
	return os.Getenv(EnvChainlinkImage)
}

// Command is chainlink entrypoint
func (c *ChainlinkClient) Command() []string {
	return []string{"chainlink"}
}

// Args returns chainlink args
func (c *ChainlinkClient) Args() []string {
	args := []string{"local", "node"}

	args = append(args, ChainlinkPassword, "/secrets/keystore-password")
	args = append(args, ChainlinkAPI, "/.chainlink/.api")

	return args
}

func (c *ChainlinkClient) Env() []corev1.EnvVar {
	node := c.node
	env := []corev1.EnvVar{
		{
			// TODO: update root to data dir
			Name:  EnvRoot,
			Value: "/.chainlink",
		},
		{
			Name:  EnvChainID,
			Value: fmt.Sprintf("%d", node.Spec.EthereumChainId),
		},
		{
			Name:  EnvEthereumURL,
			Value: node.Spec.EthereumWSEndpoint,
		},
		{
			Name:  EnvLinkContractAddress,
			Value: node.Spec.LinkContractAddress,
		},
		{
			Name:  EnvDatabaseURL,
			Value: node.Spec.DatabaseURL,
		},
	}

	return env
}

// HomeDir returns chainlink image home directory
func (c *ChainlinkClient) HomeDir() string {
	return ChainlinkHomeDir
}
