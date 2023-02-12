package ethereum2

import (
	"fmt"
	"strings"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// PrysmBeaconNode is Prysmatic Labs Ethereum 2.0 client
// https://github.com/prysmaticlabs/prysm
type PrysmBeaconNode struct {
	node *ethereum2v1alpha1.BeaconNode
}

// HomeDir returns container home directory
func (t *PrysmBeaconNode) HomeDir() string {
	return PrysmHomeDir
}

// Command returns environment variables for running the client
func (t *PrysmBeaconNode) Env() []corev1.EnvVar {
	return nil
}

// Args returns command line arguments required for client
func (t *PrysmBeaconNode) Args() (args []string) {

	node := t.node

	args = append(args, PrysmAcceptTermsOfUse)

	args = append(args, PrysmDataDir, shared.PathData(t.HomeDir()))

	args = append(args, PrysmLogging, string(t.node.Spec.Logging))

	args = append(args, PrysmExecutionEngineEndpoint, node.Spec.ExecutionEngineEndpoint)

	args = append(args, PrysmFeeRecipient, string(node.Spec.FeeRecipient))

	jwtSecretPath := fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(t.HomeDir()))
	args = append(args, PrysmJwtSecretFile, jwtSecretPath)

	args = append(args, fmt.Sprintf("--%s", node.Spec.Network))

	if node.Spec.CheckpointSyncURL != "" {
		args = append(args, PrysmCheckpointSyncUrl, node.Spec.CheckpointSyncURL)
		args = append(args, PrysmGenesisBeaconApiUrl, node.Spec.CheckpointSyncURL)
	}

	if node.Spec.RPCPort != 0 {
		args = append(args, PrysmRPCPort, fmt.Sprintf("%d", node.Spec.RPCPort))
	}

	// RPC is always on in prysm
	args = append(args, PrysmRPCHost, shared.Host(true))

	if node.Spec.GRPC {
		args = append(args, PrysmGRPCGatewayCorsDomains, strings.Join(node.Spec.CORSDomains, ","))
		args = append(args, PrysmGRPCPort, fmt.Sprintf("%d", node.Spec.GRPCPort))
		args = append(args, PrysmGRPCHost, shared.Host(node.Spec.GRPC))
	} else {
		args = append(args, PrysmDisableGRPC)
	}

	if node.Spec.CertSecretName != "" {
		args = append(args, PrysmTLSCert, fmt.Sprintf("%s/tls.crt", shared.PathSecrets(t.HomeDir())))
		args = append(args, PrysmTLSKey, fmt.Sprintf("%s/tls.key", shared.PathSecrets(t.HomeDir())))
	}

	args = append(args, PrysmP2PTCPPort, fmt.Sprintf("%d", node.Spec.P2PPort))
	args = append(args, PrysmP2PUDPPort, fmt.Sprintf("%d", node.Spec.P2PPort))

	return
}

// Command returns command for running the client
func (t *PrysmBeaconNode) Command() (command []string) {
	command = []string{"beacon-chain"}
	return
}
