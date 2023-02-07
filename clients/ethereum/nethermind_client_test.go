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

var _ = Describe("Nethermind Client", func() {

	enode := ethereumv1alpha1.Enode("enode://2281549869465d98e90cebc45e1d6834a01465a990add7bcf07a49287e7e66b50ca27f9c70a46190cef7ad746dd5d5b6b9dfee0c9954104c8e9bd0d42758ec58@10.5.0.2:30300")
	coinbase := "0x5A0b54D5dc17e0AadC383d2db43B0a0D3E029c4c"

	Context("general", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gneral",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Client: ethereumv1alpha1.NethermindClient,
				StaticNodes: []ethereumv1alpha1.Enode{
					enode,
				},
			},
		}
		client, _ := NewClient(node)

		It("should return correct home directory", func() {
			Expect(client.HomeDir()).To(Equal(NethermindHomeDir))
		})

		It("should encode static nodes correctly", func() {
			Expect(client.EncodeStaticNodes()).To(Equal(fmt.Sprintf("[\"%s\"]", string(enode))))
		})
	})

	Context("Joining mainnet", func() {
		node := ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nethermind-mainnet-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Client:                   ethereumv1alpha1.NethermindClient,
				Network:                  ethereumv1alpha1.MainNetwork,
				NodePrivateKeySecretName: "mainnet-nethermind-nodekey",
				Logging:                  sharedAPI.WarnLogs,
				RPC:                      true,
				RPCPort:                  8799,
				RPCAPI: []ethereumv1alpha1.API{
					ethereumv1alpha1.AdminAPI,
					ethereumv1alpha1.NetworkAPI,
				},
				Engine:        true,
				EnginePort:    8552,
				JWTSecretName: "jwt-secret",
				P2PPort:       30306,
				WS:            true,
				WSPort:        30307,
				SyncMode:      ethereumv1alpha1.FastSynchronization,
				StaticNodes: []ethereumv1alpha1.Enode{
					enode,
				},
			},
		}

		node.Default()

		It("Should generate correct args", func() {
			client, err := NewClient(&node)

			Expect(err).To(BeNil())
			Expect(client.Args()).To(ContainElements(
				NethermindNodePrivateKey,
				fmt.Sprintf("%s/kotal_nodekey", shared.PathData(client.HomeDir())),
				NethermindStaticNodesFile,
				fmt.Sprintf("%s/static-nodes.json", shared.PathConfig(client.HomeDir())),
				NethermindDataPath,
				shared.PathData(client.HomeDir()),
				NethermindNetwork,
				node.Spec.Network,
				NethermindP2PPort,
				"30306",
				NethermindFastSync,
				"true",
				NethermindFastBlocks,
				"true",
				NethermindDownloadBodiesInFastSync,
				"true",
				NethermindDownloadReceiptsInFastSync,
				"true",
				NethermindRPCHTTPEnabled,
				"true",
				NethermindRPCHTTPHost,
				"0.0.0.0",
				NethermindRPCHTTPPort,
				"8799",
				NethermindRPCHTTPAPI,
				"admin,net",
				NethermindRPCEnginePort,
				"8552",
				NethermindRPCEngineHost,
				"0.0.0.0",
				NethermindRPCJwtSecretFile,
				fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir())),
				NethermindRPCWSEnabled,
				"true",
				NethermindRPCWSPort,
				"30307",
			))

		})

	})

	Context("miner in private PoW network", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nethermind-pow-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Genesis: &ethereumv1alpha1.Genesis{
					ChainID:   12345,
					NetworkID: 12345,
					Ethash:    &ethereumv1alpha1.Ethash{},
				},
				Client:   ethereumv1alpha1.NethermindClient,
				Miner:    true,
				Coinbase: sharedAPI.EthereumAddress(coinbase),
				Import: &ethereumv1alpha1.ImportedAccount{
					PrivateKeySecretName: "nethermind-pow-account-key",
					PasswordSecretName:   "nethermind-pow-account-password",
				},
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).To(ContainElements(
				NethermindMiningEnabled,
				"true",
				NethermindMinerCoinbase,
				coinbase,
				NethermindUnlockAccounts,
				fmt.Sprintf("[%s]", coinbase),
				NethermindPasswordFiles,
				fmt.Sprintf("[%s/account.password]", shared.PathSecrets(client.HomeDir())),
				NethermindDiscoveryEnabled,
				"false",
				NethermindNetwork,
				fmt.Sprintf("%s/empty.cfg", shared.PathConfig(client.HomeDir())),
			))
		})

	})

	Context("signer in private PoA network", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "nethermind-poa-node",
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
				Client:   ethereumv1alpha1.NethermindClient,
				Miner:    true,
				Coinbase: sharedAPI.EthereumAddress(coinbase),
				Import: &ethereumv1alpha1.ImportedAccount{
					PrivateKeySecretName: "nethermind-poa-account-key",
					PasswordSecretName:   "nethermind-poa-account-password",
				},
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).To(ContainElements(
				NethermindMiningEnabled,
				"true",
				NethermindMinerCoinbase,
				coinbase,
				NethermindUnlockAccounts,
				fmt.Sprintf("[%s]", coinbase),
				NethermindPasswordFiles,
				fmt.Sprintf("[%s/account.password]", shared.PathSecrets(client.HomeDir())),
				NethermindDiscoveryEnabled,
				"false",
				NethermindNetwork,
				fmt.Sprintf("%s/empty.cfg", shared.PathConfig(client.HomeDir())),
			))
		})

	})

})
