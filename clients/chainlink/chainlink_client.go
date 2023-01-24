package chainlink

import (
	"fmt"
	"strings"

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
	args := []string{"local", "node"}

	args = append(args, ChainlinkPassword, fmt.Sprintf("%s/keystore-password", shared.PathSecrets(c.HomeDir())))
	args = append(args, ChainlinkAPI, fmt.Sprintf("%s/.api", shared.PathData(c.HomeDir())))

	return args
}

func (c *ChainlinkClient) Env() []corev1.EnvVar {
	node := c.node
	env := []corev1.EnvVar{
		{
			// TODO: update root to data dir
			Name:  EnvRoot,
			Value: shared.PathData(c.HomeDir()),
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
		{
			Name:  EnvLogLevel,
			Value: string(c.node.Spec.Logging),
		},
		{
			Name:  EnvAllowOrigins,
			Value: strings.Join(c.node.Spec.CORSDomains, ","),
		},
		{
			Name:  EnvSecureCookies,
			Value: fmt.Sprintf("%t", c.node.Spec.SecureCookies),
		},
		// TODO: update with P2P_ANNOUNCE_PORT
		{
			Name:  EnvP2PListenPort,
			Value: fmt.Sprintf("%d", node.Spec.P2PPort),
		},
		{
			Name:  EnvPort,
			Value: fmt.Sprintf("%d", node.Spec.APIPort),
		},
	}

	if c.node.Spec.CertSecretName != "" {
		env = append(env,
			corev1.EnvVar{
				Name:  EnvTLSCertPath,
				Value: fmt.Sprintf("%s/tls.crt", shared.PathSecrets(c.HomeDir())),
			},
			corev1.EnvVar{
				Name:  EnvTLSKeyPath,
				Value: fmt.Sprintf("%s/tls.key", shared.PathSecrets(c.HomeDir())),
			},
			corev1.EnvVar{
				Name:  EnvTLSPort,
				Value: fmt.Sprintf("%d", node.Spec.TLSPort),
			},
		)
	}

	extraEndpoints := []string{}
	for i, endpoint := range c.node.Spec.EthereumHTTPEndpoints {
		if i == 0 {
			env = append(env, corev1.EnvVar{
				Name:  EnvHTTPURL,
				Value: endpoint,
			})
		} else {
			extraEndpoints = append(extraEndpoints, endpoint)
		}
	}
	if len(extraEndpoints) != 0 {
		env = append(env, corev1.EnvVar{
			Name:  EnvSecondaryURLs,
			Value: strings.Join(extraEndpoints, ","),
		})
	}

	return env
}

// HomeDir returns chainlink image home directory
func (c *ChainlinkClient) HomeDir() string {
	return ChainlinkHomeDir
}
