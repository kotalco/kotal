package ethereum2

import (
	"fmt"
	"strings"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// LighthouseBeaconNode is SigmaPrime Ethereum 2.0 client
// https://github.com/sigp/lighthouse
type LighthouseBeaconNode struct {
	node *ethereum2v1alpha1.BeaconNode
}

// HomeDir returns container home directory
func (t *LighthouseBeaconNode) HomeDir() string {
	return LighthouseHomeDir
}

// Command returns environment variables for running the client
func (t *LighthouseBeaconNode) Env() []corev1.EnvVar {
	return nil
}

// Args returns command line arguments required for client
func (t *LighthouseBeaconNode) Args() (args []string) {

	node := t.node

	args = append(args, LighthouseDataDir, shared.PathData(t.HomeDir()))

	args = append(args, LighthouseDebugLevel, string(t.node.Spec.Logging))

	args = append(args, LighthouseNetwork, node.Spec.Network)

	args = append(args, LighthouseExecutionEngineEndpoint, node.Spec.ExecutionEngineEndpoint)

	args = append(args, LighthouseFeeRecipient, string(node.Spec.FeeRecipient))

	jwtSecretPath := fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(t.HomeDir()))
	args = append(args, LighthouseJwtSecretFile, jwtSecretPath)

	if node.Spec.REST {
		args = append(args, LighthouseHTTP)
		args = append(args, LighthouseAllowOrigins, strings.Join(node.Spec.CORSDomains, ","))
		args = append(args, LighthouseHTTPPort, fmt.Sprintf("%d", node.Spec.RESTPort))
		args = append(args, LighthouseHTTPAddress, shared.Host(node.Spec.REST))
	}

	if node.Spec.CheckpointSyncURL != "" {
		args = append(args, LighthouseCheckpointSyncUrl, node.Spec.CheckpointSyncURL)
	}

	args = append(args, LighthousePort, fmt.Sprintf("%d", node.Spec.P2PPort))
	args = append(args, LighthouseDiscoveryPort, fmt.Sprintf("%d", node.Spec.P2PPort))

	return
}

// Command returns command for running the client
func (t *LighthouseBeaconNode) Command() (command []string) {
	command = []string{"lighthouse", "bn"}
	return
}
