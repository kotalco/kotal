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
		fixedDifficulty uint = 1500
		coinbase             = EthereumAddress("0xd2c21213027cbf4d46c16b55fa98e5252b048706")
	)

	createCases := []struct {
		Title  string
		Node   *Node
		Errors field.ErrorList
	}{
		{
			Title: "node #2",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					Genesis: &Genesis{
						ChainID: 444,
					},
					Client:  BesuClient,
					Network: RinkebyNetwork,
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
			Title: "node #10",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					Genesis: &Genesis{
						ChainID: 55555,
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
					Genesis: &Genesis{
						ChainID: 55555,
						IBFT2:   &IBFT2{},
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
			Title: "node #16",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					Genesis: &Genesis{
						ChainID:   55555,
						NetworkID: networkID,
						Ethash:    &Ethash{},
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
					Genesis: &Genesis{
						ChainID:   55555,
						NetworkID: networkID,
						Ethash:    &Ethash{},
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
					Detail:   "client doesn't support importing accounts",
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
					Genesis: &Genesis{
						NetworkID: networkID,
						ChainID:   55555,
						IBFT2:     &IBFT2{},
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
					Genesis: &Genesis{
						ChainID:   55555,
						NetworkID: networkID,
						Clique:    &Clique{},
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
					Genesis: &Genesis{
						ChainID:   55555,
						NetworkID: networkID,
						Clique:    &Clique{},
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
					Genesis: &Genesis{
						ChainID:   55555,
						NetworkID: networkID,
						Clique:    &Clique{},
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
					Genesis: &Genesis{
						ChainID:   55555,
						NetworkID: networkID,
						Ethash: &Ethash{
							FixedDifficulty: &fixedDifficulty,
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
					Client:   BesuClient,
					Network:  RinkebyNetwork,
					SyncMode: LightSynchronization,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: "besu",
					Detail:   "must be geth or nethermind if syncMode is light",
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
					Client:  BesuClient,
					Network: RinkebyNetwork,
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
					Client:  BesuClient,
					Network: RinkebyNetwork,
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
					Client:  GethClient,
					Network: RinkebyNetwork,
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
					Client:  GethClient,
					Network: RinkebyNetwork,
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
					Client:  ParityClient,
					Network: RinkebyNetwork,
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
					Client:  ParityClient,
					Network: RinkebyNetwork,
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
					Client:  ParityClient,
					Network: RinkebyNetwork,
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
					Client:  ParityClient,
					Network: RinkebyNetwork,
					GraphQL: true,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: ParityClient,
					Detail:   "client doesn't support GraphQL",
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
					Genesis: &Genesis{
						ChainID:   55555,
						NetworkID: networkID,
						IBFT2:     &IBFT2{},
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
					Client:   ParityClient,
					Network:  RinkebyNetwork,
					SyncMode: LightSynchronization,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: "parity",
					Detail:   "must be geth or nethermind if syncMode is light",
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
					Genesis: &Genesis{
						ChainID:   55555,
						NetworkID: networkID,
						Ethash:    &Ethash{},
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
					Client:  GethClient,
					Network: RinkebyNetwork,
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
		{
			Title: "node #38",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					Client:  NethermindClient,
					Network: RinkebyNetwork,
					Hosts:   []string{"kotal.com"},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: NethermindClient,
					Detail:   "client doesn't support hosts whitelisting",
				},
			},
		},
		{
			Title: "node #39",
			Node: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					Client:      NethermindClient,
					Network:     RinkebyNetwork,
					CORSDomains: []string{"kotal.com"},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: NethermindClient,
					Detail:   "client doesn't support CORS domains",
				},
			},
		},
	}

	// TODO: move .resources validation to shared resources package
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
					Client:  BesuClient,
					Network: RinkebyNetwork,
				},
			},
			NewNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-1",
				},
				Spec: NodeSpec{
					Client:  BesuClient,
					Network: RopstenNetwork,
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
					Name: "node-2",
				},
				Spec: NodeSpec{
					Client:  BesuClient,
					Network: RopstenNetwork,
					Resources: shared.Resources{
						Storage: "20Gi",
					},
				},
			},
			NewNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-2",
				},
				Spec: NodeSpec{
					Client:  BesuClient,
					Network: RopstenNetwork,
					Resources: shared.Resources{
						Storage: "10Gi",
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.resources.storage",
					BadValue: "10Gi",
					Detail:   "must be greater than or equal to old storage 20Gi",
				},
			},
		},
		{
			Title: "node #3",
			OldNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-3",
				},
				Spec: NodeSpec{
					Client:  BesuClient,
					Network: RopstenNetwork,
					Resources: shared.Resources{
						CPU:      "1",
						CPULimit: "2",
					},
				},
			},
			NewNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-3",
				},
				Spec: NodeSpec{
					Client:  BesuClient,
					Network: RopstenNetwork,
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
			Title: "node #4",
			OldNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-4",
				},
				Spec: NodeSpec{
					Client:  BesuClient,
					Network: RopstenNetwork,
					Resources: shared.Resources{
						Memory:      "1Gi",
						MemoryLimit: "2Gi",
					},
				},
			},
			NewNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-4",
				},
				Spec: NodeSpec{
					Client:  BesuClient,
					Network: RopstenNetwork,
					Resources: shared.Resources{
						Memory:      "1Gi",
						MemoryLimit: "1Gi",
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.resources.memoryLimit",
					BadValue: "1Gi",
					Detail:   "must be greater than memory 1Gi",
				},
			},
		},
		{
			Title: "node #5",
			OldNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-5",
				},
				Spec: NodeSpec{
					Client:  BesuClient,
					Network: RopstenNetwork,
				},
			},
			NewNode: &Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "node-5",
				},
				Spec: NodeSpec{
					Client:  GethClient,
					Network: RopstenNetwork,
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.client",
					BadValue: GethClient,
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
