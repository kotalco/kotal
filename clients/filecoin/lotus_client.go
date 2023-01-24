package filecoin

import (
	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
)

// LotusClient is lotus filecoin client
// https://github.com/filecoin-project/lotus
type LotusClient struct {
	node *filecoinv1alpha1.Node
}

// Images
const (
	//  LotusHomeDir is lotus client image home dir
	LotusHomeDir = "/home/filecoin"
)

// Command is lotus image command
func (c *LotusClient) Command() (command []string) {
	command = append(command, "lotus", "daemon")
	return
}

// Command returns environment variables for the client
func (c *LotusClient) Env() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name:  EnvLotusPath,
			Value: shared.PathData(c.HomeDir()),
		},
		{
			Name:  EnvLogLevel,
			Value: string(c.node.Spec.Logging),
		},
	}
}

// Args returns lotus client args from node spec
func (c *LotusClient) Args() []string {
	return nil
}

// HomeDir returns lotus image home directory
func (c *LotusClient) HomeDir() string {
	return LotusHomeDir
}
