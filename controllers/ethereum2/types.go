package controllers

// Teku client arguments
const (
	// TekuNetwork is the argument used for selecting network
	TekuNetwork = "--network"
	// TekuEth1Endpoint is the argument used for Ethereum 1 JSON RPC endpoint
	TekuEth1Endpoint = "--eth1-endpoint"
)

// Prysm client arguments
const (
	// PrysmWeb3Provider is the argument used for Ethereum 1 JSON RPC endpoint
	PrysmWeb3Provider = "--http-web3provider"
	// PrysmAcceptTermsOfUse is the argument used for accepting terms of use
	PrysmAcceptTermsOfUse = "--accept-terms-of-use"
)

// Lighthouse client arguments
const (
	// LighthouseNetwork is the argument used for selecting network
	LighthouseNetwork = "--network"
	// LighthouseEth1 is the argument used for connecting to Ethereum 1 node
	LighthouseEth1 = "--eth1"
	// LighthouseEth1Endpoints is the argument used for Ethereum 1 JSON RPC endpoints
	LighthouseEth1Endpoints = "--eth1-endpoints"
)

const (
	// NimbusNonInteractive is the argument used for non interactive mode
	NimbusNonInteractive = "--non-interactive"
	// NimbusNetwork is the argument used for selecting network
	NimbusNetwork = "--network"
	// NimbusEth1Endpoint is the argument used for Ethereum 1 JSON RPC endpoint
	NimbusEth1Endpoint = "--web3-url"
)
