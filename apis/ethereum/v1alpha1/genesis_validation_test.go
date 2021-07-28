package v1alpha1

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Genesis Block validation", func() {

	createCases := []struct {
		Title   string
		Genesis *Genesis
		Errors  field.ErrorList
	}{}

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
					BadValue: EthereumAddress("0x0000000000000000000000000000000000000001"),
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
		}, {
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