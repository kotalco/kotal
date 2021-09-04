package v1alpha1

const (
	// DefaultAPIPort is the default port the API server is listening to
	DefaultAPIPort uint = 1234
	// DefaultHost is the default host used by API server and p2p
	DefaultHost = "0.0.0.0"
	// DefaultAPIRequestTimeout is the default API request timeout
	DefaultAPIRequestTimeout uint = 30
	// DefaultNerpaNodeCPURequest is the default nerpa node cpu request
	DefaultNerpaNodeCPURequest = "4"
	// DefaultNerpaNodeCPULimit is the default nerpa node cpu limit
	DefaultNerpaNodeCPULimit = "8"
	// DefaultNerpaNodeMemoryRequest is the default nerpa node memory request
	DefaultNerpaNodeMemoryRequest = "8Gi"
	// DefaultNerpaNodeMemoryLimit is the default nerpa node memory limit
	DefaultNerpaNodeMemoryLimit = "16Gi"
	// DefaultNerpaNodeStorageRequest is the default nerpa node storage
	DefaultNerpaNodeStorageRequest = "100Gi"

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
