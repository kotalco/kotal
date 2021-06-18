package v1alpha1

import (
	"fmt"

	"github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Ethereum node validation", func() {

	var (
		networkID       uint = 77777
		newNetworkID    uint = 8888
		fixedDifficulty uint = 1500
		coinbase             = EthereumAddress("0xd2c21213027cbf4d46c16b55fa98e5252b048706")
	)

	createCases := []struct {
		Title  string
		Node   *Node
		Errors field.ErrorList
	}{
		{
			Title: "node #1",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network:   RinkebyNetwork,
						Consensus: ProofOfWork,
					},
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.consensus",
					BadValue: ProofOfWork,
					Detail:   "must be none while joining a network",
				},
			},
		},
		{
			Title: "node #2",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RinkebyNetwork,
						Genesis: &Genesis{
							ChainID: 444,
						},
					},
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.network",
					BadValue: RinkebyNetwork,
					Detail:   "must be none if spec.genesis is specified",
				},
			},
		},
		{
			Title: "node #3",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis",
					BadValue: "",
					Detail:   "must be specified if spec.network is none",
				},
			},
		},
		{
			Title: "node #4",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Consensus: ProofOfAuthority,
						Genesis: &Genesis{
							ChainID: 1,
						},
					},
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis.chainId",
					BadValue: "1",
					Detail:   "can't use chain id of mainnet network to avoid tx replay",
				},
			},
		},
		{
			Title: "node #5",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Consensus: ProofOfWork,
						Genesis: &Genesis{
							ChainID: 55555,
							Clique:  &Clique{},
						},
					},
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.consensus",
					BadValue: ProofOfWork,
					Detail:   "must be poa if spec.genesis.clique is specified",
				},
			},
		},
		{
			Title: "node #6",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Consensus: ProofOfWork,
						Genesis: &Genesis{
							ChainID: 55555,
							IBFT2:   &IBFT2{},
						},
					},
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.consensus",
					BadValue: ProofOfWork,
					Detail:   "must be ibft2 if spec.genesis.ibft2 is specified",
				},
			},
		},
		{
			Title: "node #7",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Consensus: IstanbulBFT,
						Genesis: &Genesis{
							ChainID: 55555,
							Ethash:  &Ethash{},
						},
					},
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.consensus",
					BadValue: IstanbulBFT,
					Detail:   "must be pow if spec.genesis.ethash is specified",
				},
			},
		},
		{
			Title: "node #10",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Consensus: IstanbulBFT,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Miner:  true,
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.coinbase",
					BadValue: "",
					Detail:   "must provide coinbase if miner is true",
				},
			},
		},
		{
			Title: "node #11",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Consensus: IstanbulBFT,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Coinbase: EthereumAddress("0x676aEda2E67D24eb304cFf75A5190824831E3399"),
					Client:   BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.miner",
					BadValue: false,
					Detail:   "must set miner to true if coinbase is provided",
				},
			},
		},
		{
			Title: "node #12",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.consensus",
					BadValue: "",
					Detail:   "must be specified if spec.genesis is provided",
				},
			},
		},
		{
			Title: "node #13",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Consensus: ProofOfAuthority,
						Genesis: &Genesis{
							ChainID: 55555,
							Forks: &Forks{
								EIP150:    1,
								Homestead: 2,
							},
						},
					},
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis.forks.eip150",
					BadValue: "1",
					Detail:   "Fork eip150 can't be activated (at block 1) before fork homestead (at block 2)",
				},
			},
		},
		{
			Title: "node #14",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						ID:      networkID,
						Network: RinkebyNetwork,
					},
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.id",
					BadValue: fmt.Sprintf("%d", networkID),
					Detail:   "must be none if spec.network is provided",
				},
			},
		},
		{
			Title: "node #15",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Consensus: ProofOfWork,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.id",
					BadValue: "",
					Detail:   "must be specified if spec.network is none",
				},
			},
		},
		{
			Title: "node #16",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						ID:        networkID,
						Consensus: ProofOfWork,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client:   GethClient,
					Miner:    true,
					Coinbase: coinbase,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.import",
					BadValue: "",
					Detail:   "must import coinbase account",
				},
			},
		},
		{
			Title: "node #18",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						ID:        networkID,
						Consensus: ProofOfWork,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client:   BesuClient,
					Miner:    true,
					Coinbase: coinbase,
					Import: &ImportedAccount{
						PrivateKeySecretName: "my-account-privatekey",
						PasswordSecretName:   "my-account-password",
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: "besu",
					Detail:   "must be geth or parity if import is provided",
				},
			},
		},
		{
			Title: "node #19",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						ID:        networkID,
						Consensus: IstanbulBFT,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client: GethClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: "geth",
					Detail:   "client doesn't support ibft2 consensus",
				},
			},
		},
		{
			Title: "node #20",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						ID:        networkID,
						Consensus: ProofOfAuthority,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client:   GethClient,
					RPC:      true,
					Miner:    true,
					Coinbase: coinbase,
					Import: &ImportedAccount{
						PrivateKeySecretName: "my-account-privatekey",
						PasswordSecretName:   "my-account-password",
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.rpc",
					BadValue: true,
					Detail:   "must be false if import is provided",
				},
			},
		},
		{
			Title: "node #21",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						ID:        networkID,
						Consensus: ProofOfAuthority,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client:   GethClient,
					WS:       true,
					Miner:    true,
					Coinbase: coinbase,
					Import: &ImportedAccount{
						PrivateKeySecretName: "my-account-privatekey",
						PasswordSecretName:   "my-account-password",
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.ws",
					BadValue: true,
					Detail:   "must be false if import is provided",
				},
			},
		},
		{
			Title: "node #22",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						ID:        networkID,
						Consensus: ProofOfAuthority,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client:   GethClient,
					GraphQL:  true,
					Miner:    true,
					Coinbase: coinbase,
					Import: &ImportedAccount{
						PrivateKeySecretName: "my-account-privatekey",
						PasswordSecretName:   "my-account-password",
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.graphql",
					BadValue: true,
					Detail:   "must be false if import is provided",
				},
			},
		},
		{
			Title: "node #23",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						ID:        networkID,
						Consensus: ProofOfWork,
						Genesis: &Genesis{
							ChainID: 55555,
							Ethash: &Ethash{
								FixedDifficulty: &fixedDifficulty,
							},
						},
					},
					Client: GethClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: "geth",
					Detail:   "client doesn't support fixed difficulty pow networks",
				},
			},
		},
		{
			Title: "node #24",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RinkebyNetwork,
					},
					Client:   BesuClient,
					SyncMode: LightSynchronization,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: "besu",
					Detail:   "must be geth if syncMode is light",
				},
			},
		},
		{
			Title: "node #25",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RinkebyNetwork,
					},
					Client: BesuClient,
					Resources: shared.Resources{
						CPU:      "2",
						CPULimit: "1",
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.resources.cpuLimit",
					BadValue: "1",
					Detail:   "must be greater than or equal to cpu 2",
				},
			},
		},
		{
			Title: "node #26",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RinkebyNetwork,
					},
					Client: BesuClient,
					Resources: shared.Resources{
						CPU:         "1",
						CPULimit:    "2",
						Memory:      "2Gi",
						MemoryLimit: "1Gi",
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.resources.memoryLimit",
					BadValue: "1Gi",
					Detail:   "must be greater than memory 2Gi",
				},
			},
		},
		{
			Title: "node #28",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RinkebyNetwork,
					},
					Client:  GethClient,
					Logging: FatalLogs,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.logging",
					BadValue: FatalLogs,
					Detail:   "not supported by client geth",
				},
			},
		},
		{
			Title: "node #29",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RinkebyNetwork,
					},
					Client:  GethClient,
					Logging: TraceLogs,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.logging",
					BadValue: TraceLogs,
					Detail:   "not supported by client geth",
				},
			},
		},
		{
			Title: "node #30",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RinkebyNetwork,
					},
					Client:  ParityClient,
					Logging: NoLogs,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.logging",
					BadValue: NoLogs,
					Detail:   "not supported by client parity",
				},
			},
		},
		{
			Title: "node #31",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RinkebyNetwork,
					},
					Client:  ParityClient,
					Logging: FatalLogs,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.logging",
					BadValue: FatalLogs,
					Detail:   "not supported by client parity",
				},
			},
		},
		{
			Title: "node #32",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RinkebyNetwork,
					},
					Client:  ParityClient,
					Logging: AllLogs,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.logging",
					BadValue: AllLogs,
					Detail:   "not supported by client parity",
				},
			},
		},
		{
			Title: "node #33",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RinkebyNetwork,
					},
					Client:  ParityClient,
					GraphQL: true,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: ParityClient,
					Detail:   "client doesn't support graphQL",
				},
			},
		},
		{
			Title: "node #34",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						ID:        networkID,
						Consensus: IstanbulBFT,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client: ParityClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: "parity",
					Detail:   "client doesn't support ibft2 consensus",
				},
			},
		},
		{
			Title: "node #35",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RinkebyNetwork,
					},
					Client:   ParityClient,
					SyncMode: LightSynchronization,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: "parity",
					Detail:   "must be geth if syncMode is light",
				},
			},
		},
		{
			Title: "node #36",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						ID:        networkID,
						Consensus: ProofOfWork,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client:   ParityClient,
					Miner:    true,
					Coinbase: coinbase,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: ParityClient,
					Detail:   "client doesn't support mining",
				},
			},
		},
		{
			Title: "node #37",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RinkebyNetwork,
					},
					Client:  GethClient,
					GraphQL: true,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.rpc",
					BadValue: false,
					Detail:   "must enable rpc if client is geth and graphql is enabled",
				},
			},
		},
	}

	updateCases := []struct {
		Title   string
		OldNode *Node
		NewNode *Node
		Errors  field.ErrorList
	}{
		{
			Title: "node #1",
			OldNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RinkebyNetwork,
					},
					Client: BesuClient,
				},
			},
			NewNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Network: RopstenNetwork,
					},
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.network",
					BadValue: RopstenNetwork,
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "node #2",
			OldNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Consensus: ProofOfAuthority,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client: BesuClient,
				},
			},
			NewNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Consensus: ProofOfWork,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.consensus",
					BadValue: ProofOfWork,
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "node #3",
			OldNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Consensus: ProofOfAuthority,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client: GethClient,
				},
			},
			NewNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						Consensus: ProofOfAuthority,
						Genesis: &Genesis{
							ChainID: 4444,
						},
					},
					Client: GethClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis",
					BadValue: "",
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "node #5",
			OldNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						ID:        networkID,
						Consensus: ProofOfAuthority,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client: BesuClient,
				},
			},
			NewNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					NetworkConfig: NetworkConfig{
						ID:        newNetworkID,
						Consensus: ProofOfAuthority,
						Genesis: &Genesis{
							ChainID: 55555,
						},
					},
					Client: BesuClient,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.id",
					BadValue: fmt.Sprintf("%d", newNetworkID),
					Detail:   "field is immutable",
				},
			},
		},
	}

	Context("While creating node", func() {
		for _, c := range createCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.Node.Default()
					err := cc.Node.ValidateCreate()

					errStatus := err.(*errors.StatusError)

					causes := shared.ErrorsToCauses(cc.Errors)

					Expect(errStatus.ErrStatus.Details.Causes).To(ContainElements(causes))
				})
			}()
		}
	})

	Context("While updating node", func() {
		for _, c := range updateCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.OldNode.Default()
					cc.NewNode.Default()
					err := cc.NewNode.ValidateUpdate(cc.OldNode)

					errStatus := err.(*errors.StatusError)

					causes := shared.ErrorsToCauses(cc.Errors)

					Expect(errStatus.ErrStatus.Details.Causes).To(ContainElements(causes))
				})
			}()
		}
	})

})
