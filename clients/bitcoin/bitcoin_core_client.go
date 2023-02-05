package bitcoin

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	bitcoinv1alpha1 "github.com/kotalco/kotal/apis/bitcoin/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BitcoinCoreClient is Bitcoin core client
// https://github.com/bitcoin/bitcoin
type BitcoinCoreClient struct {
	node   *bitcoinv1alpha1.Node
	client client.Client
}

var hashCash map[string]string = map[string]string{}

// Images
const (
	// BitcoinCoreHomeDir is Bitcoin core image home dir
	BitcoinCoreHomeDir = "/home/bitcoin"
)

// Command returns environment variables for the client
func (c *BitcoinCoreClient) Env() (env []corev1.EnvVar) {
	env = append(env, corev1.EnvVar{
		Name:  EnvBitcoinData,
		Value: shared.PathData(c.HomeDir()),
	})

	return
}

// Command is Bitcoin core client entrypoint
func (c *BitcoinCoreClient) Command() (command []string) {
	command = append(command, "bitcoind")
	return
}

// Args returns Bitcoin core client args
func (c *BitcoinCoreClient) Args() (args []string) {
	node := c.node

	networks := map[string]string{
		"mainnet": "main",
		"testnet": "test",
	}

	args = append(args, fmt.Sprintf("%s=%s", BitcoinArgDataDir, shared.PathData(c.HomeDir())))
	args = append(args, fmt.Sprintf("%s=%s", BitcoinArgChain, networks[string(node.Spec.Network)]))
	args = append(args, fmt.Sprintf("%s=%s:%d", BitcoinArgBind, shared.Host(true), node.Spec.P2PPort))

	if c.node.Spec.RPC {
		args = append(args, fmt.Sprintf("%s=1", BitcoinArgServer))
		args = append(args, fmt.Sprintf("%s=%d", BitcoinArgRPCPort, node.Spec.RPCPort))
		args = append(args, fmt.Sprintf("%s=%s", BitcoinArgRPCBind, shared.Host(node.Spec.RPC)))
		args = append(args, fmt.Sprintf("%s=%s/0", BitcoinArgRPCAllowIp, shared.Host(node.Spec.RPC)))

		for _, rpcUser := range node.Spec.RPCUsers {
			name := types.NamespacedName{Name: rpcUser.PasswordSecretName, Namespace: node.Namespace}
			password, _ := shared.GetSecret(context.TODO(), c.client, name, "password")
			saltedHash, found := hashCash[password]
			if !found {
				salt, hash := HmacSha256(password)
				saltedHash = fmt.Sprintf("%s$%s", salt, hash)
				hashCash[password] = saltedHash
			}
			args = append(args, fmt.Sprintf("%s=%s:%s", BitcoinArgRPCAuth, rpcUser.Username, saltedHash))
		}
	} else {
		args = append(args, fmt.Sprintf("%s=0", BitcoinArgServer))
	}

	var txIndex uint
	if node.Spec.TransactionIndex {
		txIndex = 1
	}
	args = append(args, fmt.Sprintf("-txindex=%d", txIndex))

	if !node.Spec.Wallet {
		args = append(args, BitcoinArgDisableWallet)
	}

	return
}

// HomeDir is the home directory of Bitcoin core client image
func (c *BitcoinCoreClient) HomeDir() string {
	return BitcoinCoreHomeDir
}

// HmacSha256 creates new hmac sha256 hash
// reference implementation:
// https://github.com/bitcoin/bitcoin/blob/master/share/rpcauth/rpcauth.py
func HmacSha256(password string) (salt, hash string) {
	random := make([]byte, 16)
	rand.Read(random)
	salt = hex.EncodeToString(random)

	h := hmac.New(sha256.New, []byte(salt))
	h.Write([]byte(password))

	hash = fmt.Sprintf("%x", h.Sum(nil))

	return
}
