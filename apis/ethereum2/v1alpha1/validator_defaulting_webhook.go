package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate-ethereum2-kotal-io-v1alpha1-validator,mutating=true,failurePolicy=fail,groups=ethereum2.kotal.io,resources=validators,verbs=create;update,versions=v1alpha1,name=mvalidator.kb.io

var _ webhook.Defaulter = &Validator{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Validator) Default() {
	validatorlog.Info("default", "name", r.Name)

	if r.Spec.Client == "" {
		r.Spec.Client = DefaultClient
	}
}
