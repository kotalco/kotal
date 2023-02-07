package ethereum

import (
	"fmt"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	sharedAPI "github.com/kotalco/kotal/apis/shared"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Geth Client", func() {

	enode := ethereumv1alpha1.Enode("enode://2281549869465d98e90cebc45e1d6834a01465a990add7bcf07a49287e7e66b50ca27f9c70a46190cef7ad746dd5d5b6b9dfee0c9954104c8e9bd0d42758ec58@10.5.0.2:30300")
	coinbase := "0x5A0b54D5dc17e0AadC383d2db43B0a0D3E029c4c"

	Context("general", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gneral",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Client: ethereumv1alpha1.GethClient,
				StaticNodes: []ethereumv1alpha1.Enode{
					enode,
				},
			},
		}
		client, _ := NewClient(node)

		It("should return correct home directory", func() {
			Expect(client.HomeDir()).To(Equal(GethHomeDir))
		})

		It("should encode static nodes correctly", func() {

			Expect(client.EncodeStaticNodes()).To(Equal(fmt.Sprintf("[Node.P2P]\nStaticNodes = [\"%s\"]", string(enode))))
		})
	})

	Context("Joining mainnet", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "geth-mainnet-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Network:                  ethereumv1alpha1.MainNetwork,
				Client:                   ethereumv1alpha1.GethClient,
				Bootnodes:                []ethereumv1alpha1.Enode{enode},
				NodePrivateKeySecretName: "geth-mainnet-nodekey",
				StaticNodes:              []ethereumv1alpha1.Enode{enode},
				P2PPort:                  3333,
				SyncMode:                 ethereumv1alpha1.LightSynchronization,
				Logging:                  sharedAPI.WarnLogs,
				Hosts:                    []string{"whitelisted.host.com"},
				CORSDomains:              []string{"allowed.domain.com"},
				RPC:                      true,
				RPCPort:                  8888,
				RPCAPI: []ethereumv1alpha1.API{
					ethereumv1alpha1.NetworkAPI,
					ethereumv1alpha1.AdminAPI,
					ethereumv1alpha1.DebugAPI,
				},
				Engine:        true,
				EnginePort:    8552,
				JWTSecretName: "jwt-secret",
				WS:            true,
				WSPort:        7777,
				WSAPI: []ethereumv1alpha1.API{
					ethereumv1alpha1.ETHAPI,
					ethereumv1alpha1.TransactionPoolAPI,
				},
				GraphQL:     true,
				GraphQLPort: 9999,
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).To(ContainElements(
				GethDataDir,
				shared.PathData(client.HomeDir()),
				GethDisableIPC,
				fmt.Sprintf("--%s", ethereumv1alpha1.MainNetwork),
				GethLogging,
				"2", // warn logs
				GethNodeKey,
				fmt.Sprintf("%s/nodekey", shared.PathSecrets(client.HomeDir())),
				GethConfig,
				fmt.Sprintf("%s/config.toml", shared.PathConfig(client.HomeDir())),
				GethBootnodes,
				string(enode),
				GethP2PPort,
				"3333",
				GethSyncMode,
				string(ethereumv1alpha1.LightSynchronization),
				GethRPCHTTPEnabled,
				GethRPCHTTPHost,
				"0.0.0.0",
				GethRPCHTTPPort,
				"8888",
				GethRPCHTTPAPI,
				"net,admin,debug",
				GethRPCWSEnabled,
				GethRPCWSHost,
				"0.0.0.0",
				GethAuthRPCAddress,
				"0.0.0.0",
				GethAuthRPCPort,
				"8552",
				GethAuthRPCHosts,
				"whitelisted.host.com",
				GethAuthRPCJwtSecret,
				fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir())),
				GethRPCWSPort,
				"7777",
				GethRPCWSAPI,
				"eth,txpool",
				GethGraphQLHTTPEnabled,
				GethRPCHostWhitelist,
				"whitelisted.host.com",
				GethGraphQLHostWhitelist,
				"whitelisted.host.com",
				GethRPCHTTPCorsOrigins,
				"allowed.domain.com",
				GethGraphQLHTTPCorsOrigins,
				"allowed.domain.com",
				GethWSOrigins,
				"allowed.domain.com",
			))
		})
	})

	Context("miner in private PoW network", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "geth-pow-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Genesis: &ethereumv1alpha1.Genesis{
					ChainID:   12345,
					NetworkID: 12345,
					Ethash:    &ethereumv1alpha1.Ethash{},
				},
				Client:   ethereumv1alpha1.GethClient,
				Miner:    true,
				Coinbase: sharedAPI.EthereumAddress(coinbase),
				Import: &ethereumv1alpha1.ImportedAccount{
					PrivateKeySecretName: "geth-pow-account-key",
					PasswordSecretName:   "geth-pow-account-password",
				},
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).To(ContainElements(
				GethMinerEnabled,
				GethMinerCoinbase,
				coinbase,
				GethUnlock,
				coinbase,
				GethPassword,
				fmt.Sprintf("%s/account.password", shared.PathSecrets(client.HomeDir())),
				GethNetworkID,
				"12345",
				GethNoDiscovery,
			))
		})

	})

	Context("signer in private PoA network", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "geth-poa-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Genesis: &ethereumv1alpha1.Genesis{
					ChainID:   12345,
					NetworkID: 12345,
					Clique: &ethereumv1alpha1.Clique{
						Signers: []sharedAPI.EthereumAddress{
							"0xcF2C3fB8F36A863FD1A8c72E2473f81744B4CA6C",
							"0x1990E5760d9f8Ae0ec55dF8B0819C77e59846Ff2",
							"0xB87c1c66b36D98D1A74a9875EbA12c001e0bcEda",
						},
					},
				},
				Client:   ethereumv1alpha1.GethClient,
				SyncMode: ethereumv1alpha1.FullSynchronization,
				Miner:    true,
				Coinbase: sharedAPI.EthereumAddress(coinbase),
				Import: &ethereumv1alpha1.ImportedAccount{
					PrivateKeySecretName: "geth-poa-account-key",
					PasswordSecretName:   "geth-poa-account-password",
				},
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).To(ContainElements(
				GethSyncMode,
				string(node.Spec.SyncMode),
				GethCachePreImages,
				GethTxLookupLimit,
				"0",
				GethMinerEnabled,
				GethMinerCoinbase,
				coinbase,
				GethUnlock,
				coinbase,
				GethPassword,
				fmt.Sprintf("%s/account.password", shared.PathSecrets(client.HomeDir())),
				GethNetworkID,
				"12345",
				GethNoDiscovery,
			))
		})

	})

})
