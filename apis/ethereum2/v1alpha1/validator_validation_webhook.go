package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-ethereum2-kotal-io-v1alpha1-validator,mutating=false,failurePolicy=fail,groups=ethereum2.kotal.io,resources=validators,versions=v1alpha1,name=validate-ethereum2-v1alpha1-validator.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &Validator{}

// Validate validates an Ethereum 2.0 validator client
func (r *Validator) Validate() field.ErrorList {
	var validatorErrors field.ErrorList

	if r.Spec.Client == PrysmClient && r.Spec.WalletPasswordSecret == "" {
		err := field.Invalid(field.NewPath("spec").Child("walletPasswordSecret"), r.Spec.WalletPasswordSecret, "must provide walletPasswordSecret if client is prysm")
		validatorErrors = append(validatorErrors, err)
	}

	return validatorErrors
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Validator) ValidateCreate() error {
	var allErrors field.ErrorList

	validatorlog.Info("validate create", "name", r.Name)

	allErrors = append(allErrors, r.Validate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Validator) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList

	validatorlog.Info("validate update", "name", r.Name)

	allErrors = append(allErrors, r.Validate()...)

	oldValidator := old.(*Validator)

	if oldValidator.Spec.Network != r.Spec.Network {
		err := field.Invalid(field.NewPath("spec").Child("network"), r.Spec.Network, "field is immutable")
		allErrors = append(allErrors, err)
	}

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
