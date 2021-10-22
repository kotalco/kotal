package filecoin

import (
	"os"

	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
)

// LotusClient is lotus filecoin client
// https://github.com/filecoin-project/lotus
type LotusClient struct {
	node *filecoinv1alpha1.Node
}

// Images
const (
	// EnvLotusImage is the environment variable used for lotus filecoin client image
	EnvLotusImage = "LOTUS_IMAGE"
	// DefaultLotusImage is the default lotus client image
	DefaultLotusImage = "kotalco/lotus:v1.12.0"
	// DefaultLotusCalibrationImage is the default lotus client image for calibration network
	DefaultLotusCalibrationImage = "kotalco/lotus:v1.12.0-calibration"
	//  LotusHomeDir is lotus client image home dir
	LotusHomeDir = "/home/filecoin"
)

// Image returns lotus image for node's network
func (c *LotusClient) Image() string {
	if os.Getenv(EnvLotusImage) == "" {
		switch c.node.Spec.Network {
		case filecoinv1alpha1.MainNetwork:
			return DefaultLotusImage
		case filecoinv1alpha1.CalibrationNetwork:
			return DefaultLotusCalibrationImage
		}
	}
	return os.Getenv(EnvLotusImage)
}

// Command is lotus image command
func (c *LotusClient) Command() []string {
	return nil
}

// Args returns lotus client args from node spec
func (c *LotusClient) Args() (args []string) {
	args = append(args, "lotus", "daemon")
	return
}

// HomeDir returns lotus image home directory
func (c *LotusClient) HomeDir() string {
	return LotusHomeDir
}
