package controllers

const (
	// EnvIPFSPath is the environment variable used for go-ipfs path
	EnvIPFSPath = "IPFS_PATH"
	// EnvSecretsPath is the environment variable used for secrets path
	EnvSecretsPath = "SECRETS_PATH"
)

const (
	// GoIPFSDaemonArg is the argument used to run go ipfs daemon
	GoIPFSDaemonArg = "daemon"
	// GoIPFSRoutingArg is the argument used to set content routing mechanism
	GoIPFSRoutingArg = "--routing"
)
