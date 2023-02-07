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

var _ = Describe("Besu Client", func() {

	enode := ethereumv1alpha1.Enode("enode://2281549869465d98e90cebc45e1d6834a01465a990add7bcf07a49287e7e66b50ca27f9c70a46190cef7ad746dd5d5b6b9dfee0c9954104c8e9bd0d42758ec58@10.5.0.2:30300")
	coinbase := "0x5A0b54D5dc17e0AadC383d2db43B0a0D3E029c4c"

	Context("general", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "gneral",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Client: ethereumv1alpha1.BesuClient,
				StaticNodes: []ethereumv1alpha1.Enode{
					enode,
				},
			},
		}
		client, _ := NewClient(node)

		It("should return correct home directory", func() {
			Expect(client.HomeDir()).To(Equal(BesuHomeDir))
		})

		It("should encode static nodes correctly", func() {
			Expect(client.EncodeStaticNodes()).To(Equal(fmt.Sprintf("[\"%s\"]", enode)))
		})

	})

	Context("Joining mainnet", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "besu-mainnet-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Client:                   ethereumv1alpha1.BesuClient,
				Network:                  ethereumv1alpha1.MainNetwork,
				Bootnodes:                []ethereumv1alpha1.Enode{enode},
				NodePrivateKeySecretName: "besu-mainnet-nodekey",
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
				BesuDataPath,
				shared.PathData(client.HomeDir()),
				BesuNatMethod,
				"KUBERNETES",
				BesuNetwork,
				ethereumv1alpha1.MainNetwork,
				BesuLogging,
				"WARN",
				BesuNodePrivateKey,
				fmt.Sprintf("%s/nodekey", shared.PathSecrets(client.HomeDir())),
				BesuStaticNodesFile,
				fmt.Sprintf("%s/static-nodes.json", shared.PathConfig(client.HomeDir())),
				BesuBootnodes,
				string(enode),
				BesuP2PPort,
				"3333",
				BesuSyncMode,
				string(ethereumv1alpha1.LightSynchronization),
				BesuRPCHTTPEnabled,
				BesuRPCHTTPHost,
				"0.0.0.0",
				BesuRPCHTTPPort,
				"8888",
				BesuRPCHTTPAPI,
				"NET,ADMIN,DEBUG",
				BesuEngineRpcEnabled,
				BesuEngineHostAllowList,
				"whitelisted.host.com",
				BesuEngineRpcPort,
				"8552",
				BesuEngineJwtSecret,
				fmt.Sprintf("%s/jwt.secret", shared.PathSecrets(client.HomeDir())),
				BesuRPCWSEnabled,
				BesuRPCWSHost,
				"0.0.0.0",
				BesuRPCWSPort,
				"7777",
				BesuRPCWSAPI,
				"ETH,TXPOOL",
				BesuGraphQLHTTPEnabled,
				BesuGraphQLHTTPHost,
				"0.0.0.0",
				BesuGraphQLHTTPPort,
				"9999",
				BesuHostAllowlist,
				"whitelisted.host.com",
				BesuRPCHTTPCorsOrigins,
				"allowed.domain.com",
				BesuGraphQLHTTPCorsOrigins,
				"allowed.domain.com",
			))
		})

	})

	Context("miner in private PoW network", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "besu-pow-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Genesis: &ethereumv1alpha1.Genesis{
					ChainID:   12345,
					NetworkID: 12345,
					Ethash:    &ethereumv1alpha1.Ethash{},
				},
				Client:                   ethereumv1alpha1.BesuClient,
				Miner:                    true,
				NodePrivateKeySecretName: "besu-pow-nodekey",
				Coinbase:                 sharedAPI.EthereumAddress(coinbase),
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).NotTo(ContainElements(BesuNetwork))
			Expect(client.Args()).To(ContainElements(
				BesuGenesisFile,
				fmt.Sprintf("%s/genesis.json", shared.PathConfig(client.HomeDir())),
				BesuMinerEnabled,
				BesuMinerCoinbase,
				coinbase,
				BesuNetworkID,
				"12345",
				BesuDiscoveryEnabled,
				"false",
			))
		})

	})

	Context("signer in private PoA network", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "besu-poa-node",
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
				Client:                   ethereumv1alpha1.BesuClient,
				Miner:                    true,
				NodePrivateKeySecretName: "besu-poa-nodekey",
				Coinbase:                 sharedAPI.EthereumAddress(coinbase),
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).NotTo(ContainElements(BesuNetwork))
			Expect(client.Args()).To(ContainElements(
				BesuGenesisFile,
				fmt.Sprintf("%s/genesis.json", shared.PathConfig(client.HomeDir())),
				BesuMinerEnabled,
				BesuMinerCoinbase,
				coinbase,
				BesuNetworkID,
				"12345",
				BesuDiscoveryEnabled,
				"false",
			))
		})

	})

	Context("validator in private IBFT2 network", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "besu-ibft2-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				Genesis: &ethereumv1alpha1.Genesis{
					ChainID:   12345,
					NetworkID: 12345,
					IBFT2: &ethereumv1alpha1.IBFT2{
						Validators: []sharedAPI.EthereumAddress{
							"0xcF2C3fB8F36A863FD1A8c72E2473f81744B4CA6C",
							"0x1990E5760d9f8Ae0ec55dF8B0819C77e59846Ff2",
							"0xB87c1c66b36D98D1A74a9875EbA12c001e0bcEda",
						},
					},
				},
				Client:                   ethereumv1alpha1.BesuClient,
				Miner:                    true,
				NodePrivateKeySecretName: "besu-ibft2-nodekey",
				Coinbase:                 sharedAPI.EthereumAddress(coinbase),
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).NotTo(ContainElements(BesuNetwork))
			Expect(client.Args()).To(ContainElements(
				BesuGenesisFile,
				fmt.Sprintf("%s/genesis.json", shared.PathConfig(client.HomeDir())),
				BesuMinerEnabled,
				BesuMinerCoinbase,
				coinbase,
				BesuNetworkID,
				"12345",
				BesuDiscoveryEnabled,
				"false",
			))
		})

	})

})
