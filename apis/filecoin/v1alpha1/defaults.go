package v1alpha1

import "github.com/kotalco/kotal/apis/shared"

const (
	// DefaultAPIPort is the default port the API server is listening to
	DefaultAPIPort uint = 1234
	// DefaultP2PPort is the default p2p port
	DefaultP2PPort uint = 4444
	// DefaultAPIRequestTimeout is the default API request timeout
	DefaultAPIRequestTimeout uint = 30
	// DefaultLogging is the default logging verbosity
	DefaultLogging = shared.InfoLogs

	// DefaultMainnetNodeCPURequest is the default mainnet node cpu request
	DefaultMainnetNodeCPURequest = "8"
	// DefaultMainnetNodeCPULimit is the default mainnet node cpu limit
	DefaultMainnetNodeCPULimit = "16"
	// DefaultMainnetNodeMemoryRequest is the default mainnet node memory request
	DefaultMainnetNodeMemoryRequest = "16Gi"
	// DefaultMainnetNodeMemoryLimit is the default mainnet node memory limit
	DefaultMainnetNodeMemoryLimit = "32Gi"
	// DefaultMainnetNodeStorageRequest is the default mainnet node storage
	DefaultMainnetNodeStorageRequest = "200Gi"

	// DefaultCalibrationNodeCPURequest is the default calibration node cpu request
	DefaultCalibrationNodeCPURequest = "8"
	// DefaultCalibrationNodeCPULimit is the default calibration node cpu limit
	DefaultCalibrationNodeCPULimit = "16"
	// DefaultCalibrationNodeMemoryRequest is the default calibration node memory request
	DefaultCalibrationNodeMemoryRequest = "16Gi"
	// DefaultCalibrationNodeMemoryLimit is the default calibration node memory limit
	DefaultCalibrationNodeMemoryLimit = "32Gi"
	// DefaultCalibrationNodeStorageRequest is the default calibration node storage
	DefaultCalibrationNodeStorageRequest = "200Gi"
)

const (
	// DefaultLotusImage is the default lotus client image
	DefaultLotusImage = "kotalco/lotus:v1.18.0"
	// DefaultLotusCalibrationImage is the default lotus client image for calibration network
	DefaultLotusCalibrationImage = "kotalco/lotus:v1.18.0-calibration"
)
