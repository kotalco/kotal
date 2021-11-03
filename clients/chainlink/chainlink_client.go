package chainlink

import (
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
func (c *ChainlinkClient) Args() (args []string) {
	return []string{"local", "node"}
}

func (c *ChainlinkClient) Env() []corev1.EnvVar {
	return nil
}

// HomeDir returns chainlink image home directory
func (c *ChainlinkClient) HomeDir() string {
	return ChainlinkHomeDir
}
