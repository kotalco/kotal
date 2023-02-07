package v1alpha1

import (
	"fmt"

	"github.com/kotalco/kotal/apis/shared"
	. "github.com/onsi/ginkgo/v2"
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
		{
			Title: "Validator #2",
			Validator: &Validator{
				Spec: ValidatorSpec{
					Network:  "mainnet",
					Client:   TekuClient,
					Graffiti: "Kotal is amazing",
					BeaconEndpoints: []string{
						"http://10.96.130.88:9999",
						"http://10.96.130.88:9988",
					},
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.beaconEndpoints",
					BadValue: "http://10.96.130.88:9999,http://10.96.130.88:9988",
					Detail:   "multiple beacon node endpoints not supported by teku client",
				},
			},
		},
		{
			Title: "Validator #3",
			Validator: &Validator{
				Spec: ValidatorSpec{
					Network:  "mainnet",
					Client:   LighthouseClient,
					Graffiti: "Kotal is amazing",
					BeaconEndpoints: []string{
						"http://10.96.130.88:9999",
					},
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
					Field:    "spec.keystores[0].publicKey",
					BadValue: "",
					Detail:   "keystore public key is required if client is lighthouse",
				},
			},
		},
		{
			Title: "Validator #4",
			Validator: &Validator{
				Spec: ValidatorSpec{
					Network:        "mainnet",
					Client:         LighthouseClient,
					CertSecretName: "my-cert",
				},
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.certSecretName",
					BadValue: "my-cert",
					Detail:   "not supported by lighthouse client",
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
					Network:  "goerli",
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
					BadValue: "goerli",
					Detail:   "field is immutable",
				},
			},
		},
		{
			Title: "Validator #3",
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
					Field:    "spec.client",
					BadValue: "prysm",
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
					cc.OldValidator.Default()
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
