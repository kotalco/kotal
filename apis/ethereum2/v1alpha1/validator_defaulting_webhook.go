package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate-ethereum2-kotal-io-v1alpha1-validator,mutating=true,failurePolicy=fail,groups=ethereum2.kotal.io,resources=validators,verbs=create;update,versions=v1alpha1,name=mutate-ethereum2-v1alpha1-validator.kb.io

var _ webhook.Defaulter = &Validator{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Validator) Default() {
	validatorlog.Info("default", "name", r.Name)

	if r.Spec.Client == "" {
		r.Spec.Client = DefaultClient
	}

	if r.Spec.Graffiti == "" {
		r.Spec.Graffiti = DefaultGraffiti
	}

	r.DefaultNodeResources()

}

// DefaultNodeResources defaults Ethereum 2.0 validator client cpu, memory and storage resources
func (r *Validator) DefaultNodeResources() {
	if r.Spec.Resources.CPU == "" {
		r.Spec.Resources.CPU = DefaultCPURequest
	}

	if r.Spec.Resources.CPULimit == "" {
		r.Spec.Resources.CPULimit = DefaultCPULimit
	}

	if r.Spec.Resources.Memory == "" {
		r.Spec.Resources.Memory = DefaultMemoryRequest
	}

	if r.Spec.Resources.MemoryLimit == "" {
		r.Spec.Resources.MemoryLimit = DefaultMemoryLimit
	}

	if r.Spec.Resources.Storage == "" {
		r.Spec.Resources.Storage = DefaultStorage
	}
}
