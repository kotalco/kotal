package v1alpha1

import (
	"fmt"

	"github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Ethereum 2.0 validator client validation", func() {

	createCases := []struct {
		Title     string
		Validator *Validator
		Errors    field.ErrorList
	}{
		{
			Title: "Validator #1",
			Validator: &Validator{
				Spec: ValidatorSpec{
					Network:  "mainnet",
					Client:   PrysmClient,
					Graffiti: "Kotal is amazing",
					Keystores: []Keystore{
						{
							SecretName: "my-validator",
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.walletPasswordSecret",
					BadValue: "",
					Detail:   "must provide walletPasswordSecret if client is prysm",
				},
			},
		},
	}

	updateCases := []struct {
		Title        string
		OldValidator *Validator
		NewValidator *Validator
		Errors       field.ErrorList
	}{
		{
			Title: "Validator #1",
			OldValidator: &Validator{
				Spec: ValidatorSpec{
					Network:  "mainnet",
					Client:   TekuClient,
					Graffiti: "Kotal is amazing",
					Keystores: []Keystore{
						{
							SecretName: "my-validator",
						},
					},
				},
			},
			NewValidator: &Validator{
				Spec: ValidatorSpec{
					Network:  "mainnet",
					Client:   PrysmClient,
					Graffiti: "Kotal is amazing",
					Keystores: []Keystore{
						{
							SecretName: "my-validator",
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.walletPasswordSecret",
					BadValue: "",
					Detail:   "must provide walletPasswordSecret if client is prysm",
				},
			},
		},
		{
			Title: "Validator #2",
			OldValidator: &Validator{
				Spec: ValidatorSpec{
					Network:  "mainnet",
					Client:   TekuClient,
					Graffiti: "Kotal is amazing",
					Keystores: []Keystore{
						{
							SecretName: "my-validator",
						},
					},
				},
			},
			NewValidator: &Validator{
				Spec: ValidatorSpec{
					Network:  "pyrmont",
					Client:   TekuClient,
					Graffiti: "Kotal is amazing",
					Keystores: []Keystore{
						{
							SecretName: "my-validator",
						},
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.network",
					BadValue: "pyrmont",
					Detail:   "field is immutable",
				},
			},
		},
	}

	Context("While creating validator client", func() {
		for _, c := range createCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.Validator.Default()
					err := cc.Validator.ValidateCreate()

					errStatus := err.(*errors.StatusError)

					causes := shared.ErrorsToCauses(cc.Errors)

					Expect(errStatus.ErrStatus.Details.Causes).To(ContainElements(causes))
				})
			}()
		}
	})

	Context("While updating validator client", func() {
		for _, c := range updateCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					cc.NewValidator.Default()
					err := cc.NewValidator.ValidateUpdate(cc.OldValidator)

					errStatus := err.(*errors.StatusError)

					causes := shared.ErrorsToCauses(cc.Errors)

					Expect(errStatus.ErrStatus.Details.Causes).To(ContainElements(causes))
				})
			}()
		}
	})

})
