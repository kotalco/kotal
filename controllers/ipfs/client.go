package controllers

// IPFSClient is IPFS client
type IPFSClient interface {
	Image() string
	Command() []string
	Args() []string
	HomeDir() string
}

// NewIPFSClient creates new ipfs client
func NewIPFSClient() IPFSClient {
	// TODO: update after multi-client support
	return &GoIPFSClient{}
}
