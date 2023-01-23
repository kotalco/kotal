package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate-ethereum2-kotal-io-v1alpha1-validator,mutating=true,failurePolicy=fail,groups=ethereum2.kotal.io,resources=validators,verbs=create;update,versions=v1alpha1,name=mutate-ethereum2-v1alpha1-validator.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Defaulter = &Validator{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Validator) Default() {
	validatorlog.Info("default", "name", r.Name)

	if r.Spec.Graffiti == "" {
		r.Spec.Graffiti = DefaultGraffiti
	}

	if r.Spec.FeeRecipient == "" {
		r.Spec.FeeRecipient = ZeroAddress
	}

	if r.Spec.Image == "" {
		var image string

		switch r.Spec.Client {
		case TekuClient:
			image = DefaultTekuValidatorImage
		case LighthouseClient:
			image = DefaultLighthouseValidatorImage
		case NimbusClient:
			image = DefaultNimbusValidatorImage
		case PrysmClient:
			image = DefaultPrysmValidatorImage
		}

		r.Spec.Image = image
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

	if r.Spec.Logging == "" {
		r.Spec.Logging = DefaultLogging
	}
}
