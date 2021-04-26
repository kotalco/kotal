package controllers

const (
	// EnvIPFSPath is the environment variable used for go-ipfs path
	EnvIPFSPath = "IPFS_PATH"
	// EnvSecretsPath is the environment variable used for secrets path
	EnvSecretsPath = "SECRETS_PATH"
	// EnvIPFSAPIPort is the environment variable used for api port
	EnvIPFSAPIPort = "IPFS_API_PORT"
	// EnvIPFSAPIHost is the environment variable used for api host
	EnvIPFSAPIHost = "IPFS_API_HOST"
	// EnvIPFSGatewayPort is the environment variable used for local gateway port
	EnvIPFSGatewayPort = "IPFS_GATEWAY_PORT"
	// EnvIPFSGatewayHost is the environment variable used for local gateway host
	EnvIPFSGatewayHost = "IPFS_GATEWAY_HOST"
	// EnvIPFSInitProfiles is the environment variables used for initial profiles
	EnvIPFSInitProfiles = "IPFS_INIT_PROFILES"

	// EnvIPFSClusterPath is the environment variables used for ipfs-cluster-service path
	EnvIPFSClusterPath = "IPFS_CLUSTER_PATH"
	// EnvIPFSClusterConsensus is the environment variables used for ipfs cluster consnsus
	EnvIPFSClusterConsensus = "IPFS_CLUSTER_CONSENSUS"
	// EnvIPFSClusterPeerEndpoint is the environment variables used for ipfs cluster peer API endpoint
	EnvIPFSClusterPeerEndpoint = "CLUSTER_IPFSHTTP_NODEMULTIADDRESS"
	// EnvIPFSClusterPeerName is the environment variables used for ipfs cluster peer name
	EnvIPFSClusterPeerName = "CLUSTER_PEERNAME"
	// EnvIPFSClusterSecret is the environment variables used for ipfs cluster secret
	EnvIPFSClusterSecret = "CLUSTER_SECRET"
)

const (
	// GoIPFSDaemonArg is the argument used to run go ipfs daemon
	GoIPFSDaemonArg = "daemon"
	// GoIPFSRoutingArg is the argument used to set content routing mechanism
	GoIPFSRoutingArg = "--routing"

	// GoIPFSClusterDaemonArg is the argument used to run go ipfs cluster daemon
	GoIPFSClusterDaemonArg = "daemon"
	// GoIPFSClusterBootstrapArg is the argument used for go ipfs cluster bootstrap peers
	GoIPFSClusterBootstrapArg = "--bootstrap"
)
