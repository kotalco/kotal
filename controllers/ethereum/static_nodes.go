package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
)

// [Node.P2P]
// StaticNodes = []

func staticNodesFromBootnodes(bootnodes []string, client ethereumv1alpha1.EthereumClient) string {
	var result string

	if client == ethereumv1alpha1.ParityClient {
		result = strings.Join(bootnodes, "\n")
		return result
	}

	encoded, _ := json.Marshal(bootnodes)
	result = string(encoded)

	if client == ethereumv1alpha1.GethClient {
		result = fmt.Sprintf("[Node.P2P]\nStaticNodes = %s", result)
	}

	return result
}
