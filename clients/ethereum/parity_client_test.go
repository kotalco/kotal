package ethereum

import (
	"fmt"
	"os"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	"github.com/kotalco/kotal/controllers/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Nethermind Client", func() {

	enode := ethereumv1alpha1.Enode("enode://2281549869465d98e90cebc45e1d6834a01465a990add7bcf07a49287e7e66b50ca27f9c70a46190cef7ad746dd5d5b6b9dfee0c9954104c8e9bd0d42758ec58@10.5.0.2:30300")
	coinbase := "0x5A0b54D5dc17e0AadC383d2db43B0a0D3E029c4c"

	Context("general", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "parity-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Client: ethereumv1alpha1.ParityClient,
				StaticNodes: []ethereumv1alpha1.Enode{
					enode,
				},
			},
		}
		testImage := "kotalco/parity:test"
		client, _ := NewClient(node)

		It("should return correct home directory", func() {
			Expect(client.HomeDir()).To(Equal(ParityHomeDir))
		})

		It("should return correct docker image tag", func() {
			Expect(client.Image()).To(Equal(DefaultParityImage))
			os.Setenv(EnvParityImage, testImage)
			Expect(client.Image()).To(Equal(testImage))
		})

		It("should encode static nodes correctly", func() {
			Expect(client.EncodeStaticNodes()).To(Equal(string(enode)))
		})
	})

	Context("Joining mainnet", func() {
		node := ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "parity-mainnet-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Client:                   ethereumv1alpha1.ParityClient,
				Network:                  ethereumv1alpha1.MainNetwork,
				NodePrivatekeySecretName: "parity-mainnet-nodekey",
				Bootnodes:                []ethereumv1alpha1.Enode{enode},
				StaticNodes:              []ethereumv1alpha1.Enode{enode},
				P2PPort:                  3333,
				SyncMode:                 ethereumv1alpha1.FastSynchronization,
				Logging:                  ethereumv1alpha1.WarnLogs,
				Hosts:                    []string{"whitelisted.host.com"},
				CORSDomains:              []string{"allowed.domain.com"},
				RPC:                      true,
				RPCPort:                  8888,
				RPCAPI: []ethereumv1alpha1.API{
					ethereumv1alpha1.NetworkAPI,
					ethereumv1alpha1.AdminAPI,
					ethereumv1alpha1.DebugAPI,
				},
				WS:     true,
				WSPort: 7777,
				WSAPI: []ethereumv1alpha1.API{
					ethereumv1alpha1.ETHAPI,
					ethereumv1alpha1.TransactionPoolAPI,
				},
			},
		}

		node.Default()

		It("Should generate correct args", func() {
			client, err := NewClient(&node)

			Expect(err).To(BeNil())
			// because rpc and ws are disabled
			Expect(client.Args()).To(Not(ContainElements(ParityDisableRPC, ParityDisableWS)))
			Expect(client.Args()).To(ContainElements(
				ParityDataDir,
				shared.PathData(client.HomeDir()),
				// network parameter is not required in case of mainnet
				ParityLogging,
				client.LoggingArgFromVerbosity(ethereumv1alpha1.WarnLogs),
				ParityNodeKey,
				fmt.Sprintf("%s/nodekey", shared.PathSecrets(client.HomeDir())),
				ParityReservedPeers,
				fmt.Sprintf("%s/static-nodes", shared.PathConfig(client.HomeDir())),
				ParityBootnodes,
				string(enode),
				ParityP2PPort,
				"3333",
				ParitySyncMode,
				string(ethereumv1alpha1.FastSynchronization),
				ParityRPCHTTPHost,
				DefaultHost,
				ParityRPCHTTPPort,
				"8888",
				ParityRPCHTTPAPI,
				"net,admin,debug",
				ParityRPCWSHost,
				DefaultHost,
				ParityRPCWSPort,
				"7777",
				ParityRPCWSAPI,
				"eth,txpool",
				ParityRPCHostWhitelist,
				"whitelisted.host.com",
				ParityRPCWSWhitelist,
				"whitelisted.host.com",
				ParityRPCHTTPCorsOrigins,
				"allowed.domain.com",
				ParityRPCWSCorsOrigins,
				"allowed.domain.com",
			))

		})

	})

	Context("miner in private PoW network", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "parity-pow-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Genesis: &ethereumv1alpha1.Genesis{
					ChainID:   12345,
					NetworkID: 12345,
					Ethash:    &ethereumv1alpha1.Ethash{},
				},
				Client:   ethereumv1alpha1.ParityClient,
				Miner:    true,
				Coinbase: ethereumv1alpha1.EthereumAddress(coinbase),
				Import: &ethereumv1alpha1.ImportedAccount{
					PrivateKeySecretName: "parity-pow-account-key",
					PasswordSecretName:   "parity-pow-account-password",
				},
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).To(ContainElements(
				ParityMinerCoinbase,
				coinbase,
				ParityUnlock,
				coinbase,
				ParityPassword,
				fmt.Sprintf("%s/account.password", shared.PathSecrets(client.HomeDir())),
				ParityNetwork,
				fmt.Sprintf("%s/genesis.json", shared.PathConfig(client.HomeDir())),
				ParityNoDiscovery,
			))
		})

	})

	Context("signer in private PoA network", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "parity-poa-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Genesis: &ethereumv1alpha1.Genesis{
					ChainID:   12345,
					NetworkID: 12345,
					Clique: &ethereumv1alpha1.Clique{
						Signers: []ethereumv1alpha1.EthereumAddress{
							"0xcF2C3fB8F36A863FD1A8c72E2473f81744B4CA6C",
							"0x1990E5760d9f8Ae0ec55dF8B0819C77e59846Ff2",
							"0xB87c1c66b36D98D1A74a9875EbA12c001e0bcEda",
						},
					},
				},
				Client:   ethereumv1alpha1.ParityClient,
				Miner:    true,
				Coinbase: ethereumv1alpha1.EthereumAddress(coinbase),
				Import: &ethereumv1alpha1.ImportedAccount{
					PrivateKeySecretName: "parity-poa-account-key",
					PasswordSecretName:   "parity-poa-account-password",
				},
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).To(ContainElements(
				ParityMinerCoinbase,
				coinbase,
				ParityUnlock,
				coinbase,
				ParityPassword,
				fmt.Sprintf("%s/account.password", shared.PathSecrets(client.HomeDir())),
				ParityNetwork,
				fmt.Sprintf("%s/genesis.json", shared.PathConfig(client.HomeDir())),
				ParityNoDiscovery,
				ParityEngineSigner,
				coinbase,
			))
		})

	})

})
