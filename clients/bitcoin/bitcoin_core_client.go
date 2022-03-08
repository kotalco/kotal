package bitcoin

import (
	"fmt"
	"os"

	bitcoinv1alpha1 "github.com/kotalco/kotal/apis/bitcoin/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// BitcoinCoreClient is Bitcoin core client
// https://github.com/bitcoin/bitcoin
type BitcoinCoreClient struct {
	node *bitcoinv1alpha1.Node
}

// Images
const (
	// EnvBitcoinCoreImage is the environment variable used for Bitcoin core client image
	EnvBitcoinCoreImage = "BITCOIN_CORE_IMAGE"
	// DefaultBitcoinCoreImage is the default Bitcoin core client image
	DefaultBitcoinCoreImage = "ruimarinho/bitcoin-core:22.0"
	// BitcoinCoreHomeDir is Bitcoin core image home dir
	BitcoinCoreHomeDir = "/home/bitcoin"
)

// Image returns Bitcoin core client image
func (c *BitcoinCoreClient) Image() string {
	if os.Getenv(EnvBitcoinCoreImage) == "" {
		return DefaultBitcoinCoreImage
	}
	return os.Getenv(EnvBitcoinCoreImage)
}

// Command returns environment variables for the client
func (c *BitcoinCoreClient) Env() (env []corev1.EnvVar) {
	env = append(env, corev1.EnvVar{
		Name:  EnvBitcoinData,
		Value: shared.PathData(c.HomeDir()),
	})

	return
}

// Command is Bitcoin core client entrypoint
func (c *BitcoinCoreClient) Command() []string {
	return nil
}

// Args returns Bitcoin core client args
func (c *BitcoinCoreClient) Args() (args []string) {
	node := c.node

	networks := map[string]string{
		"mainnet": "main",
		"testnet": "test",
	}

	args = append(args, fmt.Sprintf("%s=%s", BitcoinArgDataDir, shared.PathData(c.HomeDir())))

	args = append(args, fmt.Sprintf("%s=%s", BitcoinArgChain, networks[string(node.Spec.Network)]))

	args = append(args, fmt.Sprintf("%s=%d", BitcoinArgRPCPort, node.Spec.RPCPort))

	return
}

// HomeDir is the home directory of Bitcoin core client image
func (c *BitcoinCoreClient) HomeDir() string {
	return BitcoinCoreHomeDir
}
