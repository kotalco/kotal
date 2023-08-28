package chainlink

import (
	"fmt"

	chainlinkv1alpha1 "github.com/kotalco/kotal/apis/chainlink/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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
		Expect(client.Env()).To(BeNil())
	})

	It("Should get correct home dir", func() {
		Expect(client.HomeDir()).To(Equal(ChainlinkHomeDir))
	})

	It("Should get correct args", func() {
		Expect(client.Args()).To(ContainElements(
			"local",
			"node",
			ChainlinkAPI,
			fmt.Sprintf("%s/.api", shared.PathData(client.HomeDir())),
		))
	})

})
