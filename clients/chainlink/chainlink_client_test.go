package chainlink

import (
	"fmt"

	chainlinkv1alpha1 "github.com/kotalco/kotal/apis/chainlink/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Chainlink Client", func() {
	node := &chainlinkv1alpha1.Node{
		Spec: chainlinkv1alpha1.NodeSpec{
			EthereumChainId:    1,
			EthereumWSEndpoint: "ws://my-eth-node:8546",
			EthereumHTTPEndpoints: []string{
				"http://my-eth-node:8545",
				"http://my-eth-node2:8545",
				"http://my-eth-node3:8545",
			},
			LinkContractAddress: "0x01BE23585060835E02B77ef475b0Cc51aA1e0709",
			DatabaseURL:         "postgresql://postgres:secret@postgres:5432/postgres",
			CertSecretName:      "my-certificate",
			TLSPort:             9999,
			P2PPort:             4444,
			APIPort:             7777,
			Logging:             sharedAPI.PanicLogs,
			CORSDomains:         []string{"*"},
			SecureCookies:       true,
		},
	}

	client := NewClient(node)

	It("Should get correct command", func() {
		Expect(client.Command()).To(ConsistOf("chainlink"))
	})

	It("Should get correct environment variables", func() {
		Expect(client.Env()).To(ContainElements(
			corev1.EnvVar{
				Name:  EnvRoot,
				Value: shared.PathData(client.HomeDir()),
			},
			corev1.EnvVar{
				Name:  EnvChainID,
				Value: "1",
			},
			corev1.EnvVar{
				Name:  EnvEthereumURL,
				Value: "ws://my-eth-node:8546",
			},
			corev1.EnvVar{
				Name:  EnvLinkContractAddress,
				Value: "0x01BE23585060835E02B77ef475b0Cc51aA1e0709",
			},
			corev1.EnvVar{
				Name:  EnvDatabaseURL,
				Value: "postgresql://postgres:secret@postgres:5432/postgres",
			},
			corev1.EnvVar{
				Name:  EnvTLSCertPath,
				Value: fmt.Sprintf("%s/tls.crt", shared.PathSecrets(client.HomeDir())),
			},
			corev1.EnvVar{
				Name:  EnvTLSKeyPath,
				Value: fmt.Sprintf("%s/tls.key", shared.PathSecrets(client.HomeDir())),
			},
			corev1.EnvVar{
				Name:  EnvTLSPort,
				Value: "9999",
			},
			corev1.EnvVar{
				Name:  EnvP2PListenPort,
				Value: "4444",
			},
			corev1.EnvVar{
				Name:  EnvPort,
				Value: "7777",
			},
			corev1.EnvVar{
				Name:  EnvHTTPURL,
				Value: "http://my-eth-node:8545",
			},
			corev1.EnvVar{
				Name:  EnvSecondaryURLs,
				Value: "http://my-eth-node2:8545,http://my-eth-node3:8545",
			},
			corev1.EnvVar{
				Name:  EnvLogLevel,
				Value: string(sharedAPI.PanicLogs),
			},
			corev1.EnvVar{
				Name:  EnvAllowOrigins,
				Value: "*",
			},
			corev1.EnvVar{
				Name:  EnvSecureCookies,
				Value: "true",
			},
		))
	})

	It("Should get correct home dir", func() {
		Expect(client.HomeDir()).To(Equal(ChainlinkHomeDir))
	})

	It("Should get correct args", func() {
		Expect(client.Args()).To(ContainElements(
			"local",
			"node",
			ChainlinkPassword,
			fmt.Sprintf("%s/keystore-password", shared.PathSecrets(client.HomeDir())),
			ChainlinkAPI,
			fmt.Sprintf("%s/.api", shared.PathData(client.HomeDir())),
		))
	})

})
