package controllers

const (
	// EnvIPFSPath is the environment variable used for go-ipfs path
	EnvIPFSPath = "IPFS_PATH"
	// EnvSecretsPath is the environment variable used for secrets path
	EnvSecretsPath = "SECRETS_PATH"
	// EnvIPFSAPIPort is the environment variable used for api port
	EnvIPFSAPIPort = "IPFS_API_PORT"
	// EnvIPFSGatewayPort is the environment variable used for local gateway port
	EnvIPFSGatewayPort = "IPFS_GATEWAY_PORT"
	// EnvIPFSAPIHost is the environment variable used for api host
	EnvIPFSAPIHost = "IPFS_API_HOST"
)

const (
	// GoIPFSDaemonArg is the argument used to run go ipfs daemon
	GoIPFSDaemonArg = "daemon"
	// GoIPFSRoutingArg is the argument used to set content routing mechanism
	GoIPFSRoutingArg = "--routing"
)
