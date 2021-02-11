package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-ethereum2-kotal-io-v1alpha1-validator,mutating=false,failurePolicy=fail,groups=ethereum2.kotal.io,resources=validators,versions=v1alpha1,name=vvalidator.kb.io

var _ webhook.Validator = &Validator{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Validator) ValidateCreate() error {
	validatorlog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Validator) ValidateUpdate(old runtime.Object) error {
	validatorlog.Info("validate update", "name", r.Name)

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Validator) ValidateDelete() error {
	validatorlog.Info("validate delete", "name", r.Name)

	return nil
}
