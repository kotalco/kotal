package controllers

import (
	"encoding/json"
	"fmt"
	"strings"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
)

// EncodeStaticNodes returns the static nodes encoding for client implementation
// TODO: change impelemntation to staticNodesFor[CLIENT_NAME]
// TODO: change to client.generateStaticNodesConfig ...
func EncodeStaticNodes(node *ethereumv1alpha1.Node) string {
	var result string
	noStaticNodes := len(node.Spec.StaticNodes) == 0

	// Parity (Open Ethereum) client static nodes are encoded as one enodeURL per line
	// enodeURL1
	// enodeURL2
	if node.Spec.Client == ethereumv1alpha1.ParityClient {
		nodes := []string{}
		for _, s := range node.Spec.StaticNodes {
			nodes = append(nodes, string(s))
		}
		result = strings.Join(nodes, "\n")
	}

	encoded, _ := json.Marshal(node.Spec.StaticNodes)

	// geth (Go Ethereum) client static nodes are encoded as
	// [Node.P2P]
	// StaticNodes = [enodeURL1, enodeURL2 ...]
	if node.Spec.Client == ethereumv1alpha1.GethClient {
		if noStaticNodes {
			encoded = []byte("[]")
		}
		result = fmt.Sprintf("[Node.P2P]\nStaticNodes = %s", string(encoded))
	}

	// Hyperledger Besu client static nodes are encoded as enodeURLs array
	// [enodeURL1, enodeURL2 ...]
	if node.Spec.Client == ethereumv1alpha1.BesuClient {
		if noStaticNodes {
			result = "[]"
		} else {
			result = string(encoded)
		}
	}

	return result
}
