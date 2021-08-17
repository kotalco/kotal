package controllers

import (
	"errors"

	filecoinv1alpha1 "github.com/kotalco/kotal/apis/filecoin/v1alpha1"
)

const (
	// ErrLotusImageNotAvailable is the error used when lotus image not available for provided network
	ErrLotusImageNotAvailable = "lotus image is not available for the provided network"

	// EnvLotusImage is the environment variable used for lotus image
	EnvLotusImage = "LOTUS_IMAGE"

	// DefaultLotusMainnetImage is the lotus image used for mainnet
	DefaultLotusMainnetImage = "kotalco/lotus:mainnet-v1.1.2"
	// DefaultLotusNerpaImage is the lotus image used for nerpa network
	DefaultLotusNerpaImage = "kotalco/lotus:nerpa-11.04.0"
	// DefaultLotusCalibrationImage is the lotus image used for calibration network
	DefaultLotusCalibrationImage = "kotalco/lotus:calibration-10.22.0"
	// DefaultLotusButterflyImage is the lotus image used for butterfly network
	DefaultLotusButterflyImage = "kotalco/lotus:butterfly-10.22.0"
)

// LotusImage returns the Filecoin lotus image to be used by the node
func LotusImage(network filecoinv1alpha1.FilecoinNetwork) (string, error) {
	switch network {
	case filecoinv1alpha1.MainNetwork:
		return DefaultLotusMainnetImage, nil
	case filecoinv1alpha1.NerpaNetwork:
		return DefaultLotusNerpaImage, nil
	case filecoinv1alpha1.CalibrationNetwork:
		return DefaultLotusCalibrationImage, nil
	default:
		return "", errors.New(ErrLotusImageNotAvailable)
	}
}
