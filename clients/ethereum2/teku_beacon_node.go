package ethereum2

import (
	"fmt"
	"strings"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// TekuBeaconNode is ConsenSys Pegasys Ethereum 2.0 client
// https://github.com/Consensys/teku/
type TekuBeaconNode struct {
	node *ethereum2v1alpha1.BeaconNode
}

// HomeDir returns container home directory
func (t *TekuBeaconNode) HomeDir() string {
	return TekuHomeDir
}

// Args returns command line arguments required for client
func (t *TekuBeaconNode) Args() (args []string) {

	node := t.node

	args = append(args, TekuDataPath, shared.PathData(t.HomeDir()))

	args = append(args, TekuNetwork, node.Spec.Network)

	args = append(args, TekuLogging, strings.ToUpper(string(node.Spec.Logging)))

	args = append(args, TekuExecutionEngineEndpoint, node.Spec.ExecutionEngineEndpoint)

	args = append(args, TekuFeeRecipient, string(node.Spec.FeeRecipient))

	jwtSecretPath := fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(t.HomeDir()))
	args = append(args, TekuJwtSecretFile, jwtSecretPath)

	if node.Spec.REST {
		args = append(args, TekuRestEnabled)
		args = append(args, TekuRESTAPICorsOrigins, strings.Join(node.Spec.CORSDomains, ","))
		args = append(args, TekuRESTAPIHostAllowlist, strings.Join(node.Spec.Hosts, ","))
		args = append(args, TekuRestPort, fmt.Sprintf("%d", node.Spec.RESTPort))
		args = append(args, TekuRestHost, shared.Host(node.Spec.REST))
	}

	if node.Spec.CheckpointSyncURL != "" {
		args = append(args, TekuInitialState, node.Spec.CheckpointSyncURL)
	}

	args = append(args, TekuP2PPort, fmt.Sprintf("%d", node.Spec.P2PPort))

	return
}

// Command returns command for running the client
func (t *TekuBeaconNode) Command() (command []string) {
	return
}

// Command returns environment variables for running the client
func (t *TekuBeaconNode) Env() []corev1.EnvVar {
	return nil
}
