package v1alpha1

import "sigs.k8s.io/controller-runtime/pkg/webhook"

// +kubebuilder:webhook:path=/mutate-ethereum2-kotal-io-v1alpha1-beaconnode,mutating=true,failurePolicy=fail,groups=ethereum2.kotal.io,resources=beaconnodes,verbs=create;update,versions=v1alpha1,name=mutate-ethereum2-v1alpha1-beaconnode.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Defaulter = &BeaconNode{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *BeaconNode) Default() {
	nodelog.Info("default", "name", r.Name)

	if r.Spec.Image == "" {
		var image string

		switch r.Spec.Client {
		case TekuClient:
			image = DefaultTekuBeaconNodeImage
		case PrysmClient:
			image = DefaultPrysmBeaconNodeImage
		case NimbusClient:
			image = DefaultNimbusBeaconNodeImage
		case LighthouseClient:
			image = DefaultLighthouseBeaconNodeImage
		}

		r.Spec.Image = image
	}

	if r.Spec.P2PPort == 0 {
		r.Spec.P2PPort = DefaultP2PPort
	}

	if r.Spec.Logging == "" {
		r.Spec.Logging = DefaultLogging
	}

	if r.Spec.RESTPort == 0 {
		r.Spec.RESTPort = DefaultRestPort
	}

	if r.Spec.RPCPort == 0 {
		r.Spec.RPCPort = DefaultRPCPort
	}

	if r.Spec.GRPCPort == 0 {
		r.Spec.GRPCPort = DefaultGRPCPort
	}

	if len(r.Spec.CORSDomains) == 0 {
		r.Spec.CORSDomains = DefaultOrigins
	}
	if len(r.Spec.Hosts) == 0 {
		r.Spec.Hosts = DefaultOrigins
	}

	if r.Spec.FeeRecipient == "" {
		r.Spec.FeeRecipient = ZeroAddress
	}

	r.DefaultNodeResources()

}

// DefaultNodeResources defaults Ethereum 2.0 node cpu, memory and storage resources
func (r *BeaconNode) DefaultNodeResources() {
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
