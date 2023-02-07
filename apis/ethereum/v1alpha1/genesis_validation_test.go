package v1alpha1

import (
	"fmt"

	"github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Genesis Block validation", func() {

	createCases := []struct {
		Title   string
		Genesis *Genesis
		Errors  field.ErrorList
	}{
		{
			Title: "using mainnet chain id",
			Genesis: &Genesis{
				ChainID:   1,
				NetworkID: 55555,
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
			Title: "bad fork activation order",
			Genesis: &Genesis{
				ChainID:   55555,
				NetworkID: 55555,
				Ethash:    &Ethash{},
				Forks: &Forks{
					EIP150:    1,
					Homestead: 2,
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
			Title: "consensus configuration is missing",
			Genesis: &Genesis{
				ChainID:   4444,
				NetworkID: 4444,
			},
			Errors: []*field.Error{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis",
					BadValue: "",
					Detail:   "consensus configuration (ethash, clique, or ibft2) is missing",
				},
			},
		},
		{
			Title: "multiple consensus configurations are used",
			Genesis: &Genesis{
				ChainID:   4444,
				NetworkID: 4444,
				Ethash:    &Ethash{},
				Clique:    &Clique{},
			},
			Errors: []*field.Error{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis",
					BadValue: "",
					Detail:   "multiple consensus configurations (clique, ethash) are enabled",
				},
			},
		},
		{
			Title: "reserved account is used",
			Genesis: &Genesis{
				ChainID:   4444,
				NetworkID: 4444,
				Ethash:    &Ethash{},
				Accounts: []Account{
					{
						Address: shared.EthereumAddress("0x0000000000000000000000000000000000000015"),
						Balance: HexString("0xffffff"),
					},
				},
			},
			Errors: []*field.Error{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis.accounts",
					BadValue: "0x0000000000000000000000000000000000000015",
					Detail:   "reserved account is used",
				},
			},
		},
	}

	updateCases := []struct {
		Title      string
		OldGenesis *Genesis
		NewGenesis *Genesis
		Errors     field.ErrorList
	}{
		{
			Title: "updating coinbase",
			OldGenesis: &Genesis{
				NetworkID: 55555,
				ChainID:   55555,
				Ethash:    &Ethash{},
				Forks:     &Forks{},
			},
			NewGenesis: &Genesis{
				NetworkID: 55555,
				ChainID:   55555,
				Coinbase:  "0x0000000000000000000000000000000000000001",
				Ethash:    &Ethash{},
				Forks:     &Forks{},
			},
			Errors: []*field.Error{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis.coinbase",
					BadValue: shared.EthereumAddress("0x0000000000000000000000000000000000000001"),
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "updating difficulty",
			OldGenesis: &Genesis{
				NetworkID: 55555,
				ChainID:   55555,
				Ethash:    &Ethash{},
				Forks:     &Forks{},
			},
			NewGenesis: &Genesis{
				NetworkID:  55555,
				ChainID:    55555,
				Difficulty: "0xffff",
				Ethash:     &Ethash{},
				Forks:      &Forks{},
			},
			Errors: []*field.Error{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis.difficulty",
					BadValue: HexString("0xffff"),
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "updating mixHash",
			OldGenesis: &Genesis{
				NetworkID: 55555,
				ChainID:   55555,
				Ethash:    &Ethash{},
				Forks:     &Forks{},
			},
			NewGenesis: &Genesis{
				NetworkID: 55555,
				ChainID:   55555,
				MixHash:   Hash("0x00000000000000000000000000000000000000000000000000000000000000ff"),
				Ethash:    &Ethash{},
				Forks:     &Forks{},
			},
			Errors: []*field.Error{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis.mixHash",
					BadValue: Hash("0x00000000000000000000000000000000000000000000000000000000000000ff"),
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "updating gasLimit",
			OldGenesis: &Genesis{
				NetworkID: 55555,
				ChainID:   55555,
				Ethash:    &Ethash{},
				Forks:     &Forks{},
			},
			NewGenesis: &Genesis{
				NetworkID: 55555,
				ChainID:   55555,
				GasLimit:  HexString("0x47bfff"),
				Ethash:    &Ethash{},
				Forks:     &Forks{},
			},
			Errors: []*field.Error{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis.gasLimit",
					BadValue: HexString("0x47bfff"),
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "updating nonce",
			OldGenesis: &Genesis{
				NetworkID: 55555,
				ChainID:   55555,
				Ethash:    &Ethash{},
				Forks:     &Forks{},
			},
			NewGenesis: &Genesis{
				NetworkID: 55555,
				ChainID:   55555,
				Nonce:     HexString("0x1"),
				Ethash:    &Ethash{},
				Forks:     &Forks{},
			},
			Errors: []*field.Error{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis.nonce",
					BadValue: HexString("0x1"),
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "updating timestamp",
			OldGenesis: &Genesis{
				NetworkID: 55555,
				ChainID:   55555,
				Ethash:    &Ethash{},
				Forks:     &Forks{},
			},
			NewGenesis: &Genesis{
				NetworkID: 55555,
				ChainID:   55555,
				Timestamp: HexString("0x1"),
				Ethash:    &Ethash{},
				Forks:     &Forks{},
			},
			Errors: []*field.Error{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis.timestamp",
					BadValue: HexString("0x1"),
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "updating accounts",
			OldGenesis: &Genesis{
				NetworkID: 55555,
				ChainID:   55555,
				Ethash:    &Ethash{},
				Forks:     &Forks{},
				Accounts: []Account{
					{
						Address: shared.EthereumAddress("0xB1368D309179D8E7f25B34398e4cF9D9dEFdC75C"),
						Balance: HexString("0xffffff"),
					},
				},
			},
			NewGenesis: &Genesis{
				NetworkID: 55555,
				ChainID:   55555,
				Ethash:    &Ethash{},
				Forks:     &Forks{},
				Accounts: []Account{
					{
						Address: shared.EthereumAddress("0xB1368D309179D8E7f25B34398e4cF9D9dEFdC75C"),
						Balance: HexString("0x111111"), // change account balance
					},
				},
			},
			Errors: []*field.Error{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.genesis.accounts",
					BadValue: "",
					Detail:   "field is immutable",
				},
			},
		},
	}

	Context("While creating genesis", func() {
		for _, c := range createCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.Genesis.Default()

					err := cc.Genesis.ValidateCreate()

					Expect(err).To(ContainElements(cc.Errors))
				})
			}()
		}
	})

	Context("While updating genesis", func() {
		for _, c := range updateCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.OldGenesis.Default()
					cc.NewGenesis.Default()

					err := cc.NewGenesis.ValidateUpdate(cc.OldGenesis)

					Expect(err).To(ContainElements(cc.Errors))
				})
			}()
		}
	})

})
