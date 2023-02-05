package v1alpha1

import "sigs.k8s.io/controller-runtime/pkg/webhook"

// +kubebuilder:webhook:path=/mutate-ipfs-kotal-io-v1alpha1-peer,mutating=true,failurePolicy=fail,groups=ipfs.kotal.io,resources=peers,verbs=create;update,versions=v1alpha1,name=mutate-ipfs-v1alpha1-peer.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Defaulter = &Peer{}

// DefaultPeerResources defaults peer resources
func (r *Peer) DefaultPeerResources() {
	if r.Spec.Resources.CPU == "" {
		r.Spec.Resources.CPU = DefaultNodeCPURequest
	}

	if r.Spec.Resources.CPULimit == "" {
		r.Spec.Resources.CPULimit = DefaultNodeCPULimit
	}

	if r.Spec.Resources.Memory == "" {
		r.Spec.Resources.Memory = DefaultNodeMemoryRequest
	}

	if r.Spec.Resources.MemoryLimit == "" {
		r.Spec.Resources.MemoryLimit = DefaultNodeMemoryLimit
	}

	if r.Spec.Resources.Storage == "" {
		r.Spec.Resources.Storage = DefaultNodeStorageRequest
	}
}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Peer) Default() {
	peerlog.Info("default", "name", r.Name)

	if r.Spec.Image == "" {
		r.Spec.Image = DefaultGoIPFSImage
	}

	if r.Spec.Routing == "" {
		r.Spec.Routing = DefaultRoutingMode
	}

	if r.Spec.APIPort == 0 {
		r.Spec.APIPort = DefaultAPIPort
	}

	if r.Spec.GatewayPort == 0 {
		r.Spec.GatewayPort = DefaultGatewayPort
	}

	if len(r.Spec.InitProfiles) == 0 {
		r.Spec.InitProfiles = []Profile{DefaultDatastoreProfile}
	}

	if r.Spec.Logging == "" {
		r.Spec.Logging = DefaultLogging
	}

	r.DefaultPeerResources()

}
