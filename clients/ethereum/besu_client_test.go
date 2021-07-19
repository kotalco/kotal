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
		testImage := "kotalco/besu:test"
		client, _ := NewClient(node)

		It("should return correct home directory", func() {
			Expect(client.HomeDir()).To(Equal(BesuHomeDir))
		})

		It("should return correct docker image tag", func() {
			Expect(client.Image()).To(Equal(DefaultBesuImage))
			os.Setenv(EnvBesuImage, testImage)
			Expect(client.Image()).To(Equal(testImage))
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
				NetworkConfig:     ethereumv1alpha1.NetworkConfig{Network: ethereumv1alpha1.MainNetwork},
				Client:            ethereumv1alpha1.BesuClient,
				Bootnodes:         []ethereumv1alpha1.Enode{enode},
				NodekeySecretName: "besu-mainnet-nodekey",
				StaticNodes:       []ethereumv1alpha1.Enode{enode},
				P2PPort:           3333,
				SyncMode:          ethereumv1alpha1.LightSynchronization,
				Logging:           ethereumv1alpha1.WarnLogs,
				Hosts:             []string{"whitelisted.host.com"},
				CORSDomains:       []string{"allowed.domain.com"},
				RPC:               true,
				RPCPort:           8888,
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
				GraphQL:     true,
				GraphQLPort: 9999,
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).To(ContainElements(
				[]string{
					BesuDataPath,
					shared.PathData(client.HomeDir()),
					BesuNatMethod,
					"KUBERNETES",
					BesuNetwork,
					ethereumv1alpha1.MainNetwork,
					BesuLogging,
					client.LoggingArgFromVerbosity(ethereumv1alpha1.WarnLogs),
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
					DefaultHost,
					BesuRPCHTTPPort,
					"8888",
					BesuRPCHTTPAPI,
					"net,admin,debug",
					BesuRPCWSEnabled,
					BesuRPCWSHost,
					DefaultHost,
					BesuRPCWSPort,
					"7777",
					BesuRPCWSAPI,
					"eth,txpool",
					BesuGraphQLHTTPEnabled,
					BesuGraphQLHTTPHost,
					DefaultHost,
					BesuGraphQLHTTPPort,
					"9999",
					BesuHostAllowlist,
					"whitelisted.host.com",
					BesuRPCHTTPCorsOrigins,
					"allowed.domain.com",
					BesuGraphQLHTTPCorsOrigins,
					"allowed.domain.com",
				},
			))
		})

	})

	Context("miner in private PoW network", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "besu-pow-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				NetworkConfig: ethereumv1alpha1.NetworkConfig{
					Consensus: ethereumv1alpha1.ProofOfWork,
					ID:        12345,
					Genesis: &ethereumv1alpha1.Genesis{
						ChainID: 12345,
						Ethash:  &ethereumv1alpha1.Ethash{},
					},
				},
				Client:            ethereumv1alpha1.BesuClient,
				Miner:             true,
				NodekeySecretName: "besu-pow-nodekey",
				Coinbase:          ethereumv1alpha1.EthereumAddress(coinbase),
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).NotTo(ContainElements(
				[]string{
					BesuNetwork,
				},
			))
			Expect(client.Args()).To(ContainElements(
				[]string{
					BesuGenesisFile,
					fmt.Sprintf("%s/genesis.json", shared.PathConfig(client.HomeDir())),
					BesuMinerEnabled,
					BesuMinerCoinbase,
					coinbase,
					BesuNetworkID,
					"12345",
					BesuDiscoveryEnabled,
					"false",
				},
			))
		})

	})

	Context("signer in private PoA network", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "besu-poa-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				NetworkConfig: ethereumv1alpha1.NetworkConfig{
					Consensus: ethereumv1alpha1.ProofOfWork,
					ID:        12345,
					Genesis: &ethereumv1alpha1.Genesis{
						ChainID: 12345,
						Clique: &ethereumv1alpha1.Clique{
							Signers: []ethereumv1alpha1.EthereumAddress{
								"0xcF2C3fB8F36A863FD1A8c72E2473f81744B4CA6C",
								"0x1990E5760d9f8Ae0ec55dF8B0819C77e59846Ff2",
								"0xB87c1c66b36D98D1A74a9875EbA12c001e0bcEda",
							},
						},
					},
				},
				Client:            ethereumv1alpha1.BesuClient,
				Miner:             true,
				NodekeySecretName: "besu-poa-nodekey",
				Coinbase:          ethereumv1alpha1.EthereumAddress(coinbase),
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).NotTo(ContainElements(
				[]string{
					BesuNetwork,
				},
			))
			Expect(client.Args()).To(ContainElements(
				[]string{
					BesuGenesisFile,
					fmt.Sprintf("%s/genesis.json", shared.PathConfig(client.HomeDir())),
					BesuMinerEnabled,
					BesuMinerCoinbase,
					coinbase,
					BesuNetworkID,
					"12345",
					BesuDiscoveryEnabled,
					"false",
				},
			))
		})

	})

	Context("validator in private IBFT2 network", func() {
		node := &ethereumv1alpha1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "besu-ibft2-node",
			},
			Spec: ethereumv1alpha1.NodeSpec{
				NetworkConfig: ethereumv1alpha1.NetworkConfig{
					Consensus: ethereumv1alpha1.ProofOfWork,
					ID:        12345,
					Genesis: &ethereumv1alpha1.Genesis{
						ChainID: 12345,
						IBFT2: &ethereumv1alpha1.IBFT2{
							Validators: []ethereumv1alpha1.EthereumAddress{
								"0xcF2C3fB8F36A863FD1A8c72E2473f81744B4CA6C",
								"0x1990E5760d9f8Ae0ec55dF8B0819C77e59846Ff2",
								"0xB87c1c66b36D98D1A74a9875EbA12c001e0bcEda",
							},
						},
					},
				},
				Client:            ethereumv1alpha1.BesuClient,
				Miner:             true,
				NodekeySecretName: "besu-ibft2-nodekey",
				Coinbase:          ethereumv1alpha1.EthereumAddress(coinbase),
			},
		}
		node.Default()

		It("should generate correct arguments", func() {

			client, err := NewClient(node)

			Expect(err).To(BeNil())
			Expect(client.Args()).NotTo(ContainElements(
				[]string{
					BesuNetwork,
				},
			))
			Expect(client.Args()).To(ContainElements(
				[]string{
					BesuGenesisFile,
					fmt.Sprintf("%s/genesis.json", shared.PathConfig(client.HomeDir())),
					BesuMinerEnabled,
					BesuMinerCoinbase,
					coinbase,
					BesuNetworkID,
					"12345",
					BesuDiscoveryEnabled,
					"false",
				},
			))
		})

	})

})
