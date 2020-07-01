package v1alpha1

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Ethereum network validation", func() {

	var (
		networkID    uint = 77777
		newNetworkID uint = 8888
	)

	createCases := []struct {
		Title   string
		Network *Network
		Errors  field.ErrorList
	}{
		{
			Title: "network #1",
			Network: &Network{
				Spec: NetworkSpec{
					Join:      RinkebyNetwork,
					Consensus: ProofOfWork,
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
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
			Title: "network #2",
			Network: &Network{
				Spec: NetworkSpec{
					Join: RinkebyNetwork,
					Genesis: &Genesis{
						ChainID: 444,
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.join",
					BadValue: RinkebyNetwork,
					Detail:   "must be none if spec.genesis is specified",
				},
			},
		},
		{
			Title: "network #3",
			Network: &Network{
				Spec: NetworkSpec{
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis",
					BadValue: "",
					Detail:   "must be specified if spec.join is none",
				},
			},
		},
		{
			Title: "network #4",
			Network: &Network{
				Spec: NetworkSpec{
					Consensus: ProofOfAuthority,
					Genesis: &Genesis{
						ChainID: 1,
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
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
			Title: "network #5",
			Network: &Network{
				Spec: NetworkSpec{
					Consensus: ProofOfWork,
					Genesis: &Genesis{
						ChainID: 55555,
						Clique:  &Clique{},
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
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
			Title: "network #6",
			Network: &Network{
				Spec: NetworkSpec{
					Consensus: ProofOfWork,
					Genesis: &Genesis{
						ChainID: 55555,
						IBFT2:   &IBFT2{},
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
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
			Title: "network #7",
			Network: &Network{
				Spec: NetworkSpec{
					Consensus: IstanbulBFT,
					Genesis: &Genesis{
						ChainID: 55555,
						Ethash:  &Ethash{},
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
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
			Title: "network #8",
			Network: &Network{
				Spec: NetworkSpec{
					Consensus: IstanbulBFT,
					Genesis: &Genesis{
						ChainID: 55555,
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
						{
							Name: "node-2",
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.nodes[0].bootnode",
					BadValue: false,
					Detail:   "first node must be a bootnode if network has multiple nodes",
				},
			},
		},
		{
			Title: "network #9",
			Network: &Network{
				Spec: NetworkSpec{
					Consensus: IstanbulBFT,
					Genesis: &Genesis{
						ChainID: 55555,
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
						{
							Name: "node-1",
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.nodes[1].name",
					BadValue: "node-1",
					Detail:   "already used by spec.nodes[0].name",
				},
			},
		},
		{
			Title: "network #10",
			Network: &Network{
				Spec: NetworkSpec{
					Consensus: IstanbulBFT,
					Genesis: &Genesis{
						ChainID: 55555,
					},
					Nodes: []Node{
						{
							Name:     "node-1",
							Bootnode: true,
						},
						{
							Name: "node-2",
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.nodes[0].nodekey",
					BadValue: "",
					Detail:   "must provide nodekey if bootnode is true",
				},
			},
		},
		{
			Title: "network #11",
			Network: &Network{
				Spec: NetworkSpec{
					Consensus: IstanbulBFT,
					Genesis: &Genesis{
						ChainID: 55555,
					},
					Nodes: []Node{
						{
							Name:  "node-1",
							Miner: true,
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.nodes[0].coinbase",
					BadValue: "",
					Detail:   "must provide coinbase if miner is true",
				},
			},
		},
		{
			Title: "network #12",
			Network: &Network{
				Spec: NetworkSpec{
					Consensus: IstanbulBFT,
					Genesis: &Genesis{
						ChainID: 55555,
					},
					Nodes: []Node{
						{
							Name:     "node-1",
							Coinbase: EthereumAddress("0x676aEda2E67D24eb304cFf75A5190824831E3399"),
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.nodes[0].miner",
					BadValue: false,
					Detail:   "must set miner to true if coinbase is provided",
				},
			},
		},
		{
			Title: "network #13",
			Network: &Network{
				Spec: NetworkSpec{
					Genesis: &Genesis{
						ChainID: 55555,
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
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
			Title: "network #14",
			Network: &Network{
				Spec: NetworkSpec{
					Consensus: ProofOfAuthority,
					Genesis: &Genesis{
						ChainID: 55555,
						Forks: &Forks{
							DAO:       1,
							Homestead: 2,
						},
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis.forks.dao",
					BadValue: "1",
					Detail:   "Fork dao can't be activated (at block 1) before fork homestead (at block 2)",
				},
			},
		},
		{
			Title: "network #15",
			Network: &Network{
				Spec: NetworkSpec{
					ID:   networkID,
					Join: RinkebyNetwork,
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.id",
					BadValue: fmt.Sprintf("%d", networkID),
					Detail:   "must be none if spec.join is provided",
				},
			},
		},
		{
			Title: "network #16",
			Network: &Network{
				Spec: NetworkSpec{
					Consensus: ProofOfWork,
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.id",
					BadValue: "",
					Detail:   "must be specified if spec.join is none",
				},
			},
		},
	}

	// errorsToCauses converts field error list into array of status cause
	errorsToCauses := func(errs field.ErrorList) []metav1.StatusCause {
		causes := make([]metav1.StatusCause, 0, len(errs))

		for i := range errs {
			err := errs[i]
			causes = append(causes, metav1.StatusCause{
				Type:    metav1.CauseType(err.Type),
				Message: err.ErrorBody(),
				Field:   err.Field,
			})
		}

		return causes
	}

	updateCases := []struct {
		Title      string
		OldNetwork *Network
		NewNetwork *Network
		Errors     field.ErrorList
	}{
		{
			Title: "network #1",
			OldNetwork: &Network{
				Spec: NetworkSpec{
					Join: RinkebyNetwork,
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
				},
			},
			NewNetwork: &Network{
				Spec: NetworkSpec{
					Join: RopstenNetwork,
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.join",
					BadValue: RopstenNetwork,
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "network #2",
			OldNetwork: &Network{
				Spec: NetworkSpec{
					Consensus: ProofOfAuthority,
					Genesis: &Genesis{
						ChainID: 55555,
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
				},
			},
			NewNetwork: &Network{
				Spec: NetworkSpec{
					Consensus: ProofOfWork,
					Genesis: &Genesis{
						ChainID: 55555,
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
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
			Title: "network #3",
			OldNetwork: &Network{
				Spec: NetworkSpec{
					Consensus: ProofOfAuthority,
					Genesis: &Genesis{
						ChainID: 55555,
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
				},
			},
			NewNetwork: &Network{
				Spec: NetworkSpec{
					Consensus: ProofOfAuthority,
					Genesis: &Genesis{
						ChainID: 4444,
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
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
			Title: "network #4",
			OldNetwork: &Network{
				Spec: NetworkSpec{
					Consensus: ProofOfAuthority,
					Genesis: &Genesis{
						ChainID: 55555,
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
				},
			},
			NewNetwork: &Network{
				Spec: NetworkSpec{
					Consensus: ProofOfAuthority,
					Genesis: &Genesis{
						ChainID: 55555,
					},
					Nodes: []Node{
						{
							Name: "node-2",
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.nodes[0].name",
					BadValue: "node-2",
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "network #5",
			OldNetwork: &Network{
				Spec: NetworkSpec{
					ID:        networkID,
					Consensus: ProofOfAuthority,
					Genesis: &Genesis{
						ChainID: 55555,
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
				},
			},
			NewNetwork: &Network{
				Spec: NetworkSpec{
					ID:        newNetworkID,
					Consensus: ProofOfAuthority,
					Genesis: &Genesis{
						ChainID: 55555,
					},
					Nodes: []Node{
						{
							Name: "node-1",
						},
					},
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

	Context("While creating network", func() {
		for _, c := range createCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.Network.Default()
					err := cc.Network.ValidateCreate()

					errStatus := err.(*errors.StatusError)

					causes := errorsToCauses(cc.Errors)

					Expect(errStatus.ErrStatus.Details.Causes).To(ContainElements(causes))
				})
			}()
		}
	})

	Context("While updating network", func() {
		for _, c := range updateCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.NewNetwork.Default()
					err := cc.NewNetwork.ValidateUpdate(cc.OldNetwork)

					errStatus := err.(*errors.StatusError)

					causes := errorsToCauses(cc.Errors)

					Expect(errStatus.ErrStatus.Details.Causes).To(ContainElements(causes))
				})
			}()
		}
	})

})
