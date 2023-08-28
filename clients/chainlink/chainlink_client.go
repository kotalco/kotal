package chainlink

import (
	"fmt"

	chainlinkv1alpha1 "github.com/kotalco/kotal/apis/chainlink/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// ChainlinkClient is chainlink official client
// https://github.com/smartcontractkit/chainlink
type ChainlinkClient struct {
	node *chainlinkv1alpha1.Node
}

// Images
const (
	// ChainlinkHomeDir is chainlink image home dir
	// TODO: update the home directory
	ChainlinkHomeDir = "/home/chainlink"
)

// Command is chainlink entrypoint
func (c *ChainlinkClient) Command() []string {
	return []string{"chainlink"}
}

// Args returns chainlink args
func (c *ChainlinkClient) Args() []string {
	args := []string{
		"local",
		"--config",
		fmt.Sprintf("%s/config.toml", shared.PathConfig(c.HomeDir())),
		"--secrets",
		fmt.Sprintf("%s/secrets.toml", shared.PathConfig(c.HomeDir())),
		"node",
	}

	args = append(args, ChainlinkAPI, fmt.Sprintf("%s/.api", shared.PathData(c.HomeDir())))

	return args
}

func (c *ChainlinkClient) Env() []corev1.EnvVar {
	return nil
}

// HomeDir returns chainlink image home directory
func (c *ChainlinkClient) HomeDir() string {
	return ChainlinkHomeDir
}
