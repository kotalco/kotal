package v1alpha1

import (
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-ethereum2-kotal-io-v1alpha1-validator,mutating=false,failurePolicy=fail,groups=ethereum2.kotal.io,resources=validators,versions=v1alpha1,name=validate-ethereum2-v1alpha1-validator.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &Validator{}

// validate validates an Ethereum 2.0 validator client
func (r *Validator) validate() field.ErrorList {
	var validatorErrors field.ErrorList

	// prysm requires wallet password
	if r.Spec.Client == PrysmClient && r.Spec.WalletPasswordSecret == "" {
		msg := "must provide walletPasswordSecret if client is prysm"
		err := field.Invalid(field.NewPath("spec").Child("walletPasswordSecret"), r.Spec.WalletPasswordSecret, msg)
		validatorErrors = append(validatorErrors, err)
	}

	if r.Spec.CertSecretName != "" && r.Spec.Client != PrysmClient {
		err := field.Invalid(field.NewPath("spec").Child("certSecretName"), r.Spec.CertSecretName, fmt.Sprintf("not supported by %s client", r.Spec.Client))
		validatorErrors = append(validatorErrors, err)
	}

	if !r.Spec.Client.SupportsVerbosityLevel(r.Spec.Logging, false) {
		err := field.Invalid(field.NewPath("spec").Child("logging"), r.Spec.Logging, fmt.Sprintf("not supported by %s client", r.Spec.Client))
		validatorErrors = append(validatorErrors, err)
	}

	// lighthouse is the only client supporting multiple beacon endpoints
	if r.Spec.Client != LighthouseClient && len(r.Spec.BeaconEndpoints) > 1 {
		msg := fmt.Sprintf("multiple beacon node endpoints not supported by %s client", r.Spec.Client)
		err := field.Invalid(field.NewPath("spec").Child("beaconEndpoints"), strings.Join(r.Spec.BeaconEndpoints, ","), msg)
		validatorErrors = append(validatorErrors, err)
	}

	if r.Spec.Client == LighthouseClient {
		for i, keystore := range r.Spec.Keystores {
			if keystore.PublicKey == "" {
				msg := "keystore public key is required if client is lighthouse"
				err := field.Invalid(field.NewPath("spec").Child("keystores").Index(i).Child("publicKey"), "", msg)
				validatorErrors = append(validatorErrors, err)
			}
		}
	}

	return validatorErrors
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Validator) ValidateCreate() error {
	var allErrors field.ErrorList

	validatorlog.Info("validate create", "name", r.Name)

	allErrors = append(allErrors, r.validate()...)
	allErrors = append(allErrors, r.Spec.Resources.ValidateCreate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Validator) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList
	oldValidator := old.(*Validator)

	validatorlog.Info("validate update", "name", r.Name)

	allErrors = append(allErrors, r.validate()...)
	allErrors = append(allErrors, r.Spec.Resources.ValidateUpdate(&oldValidator.Spec.Resources)...)

	if oldValidator.Spec.Client != r.Spec.Client {
		err := field.Invalid(field.NewPath("spec").Child("client"), r.Spec.Client, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if oldValidator.Spec.Network != r.Spec.Network {
		err := field.Invalid(field.NewPath("spec").Child("network"), r.Spec.Network, "field is immutable")
		allErrors = append(allErrors, err)
	}

	allErrors = append(allErrors, r.Spec.Resources.ValidateCreate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Validator) ValidateDelete() error {
	validatorlog.Info("validate delete", "name", r.Name)

	return nil
}
