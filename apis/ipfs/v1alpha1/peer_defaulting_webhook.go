package v1alpha1

import "sigs.k8s.io/controller-runtime/pkg/webhook"

// +kubebuilder:webhook:path=/mutate-ipfs-kotal-io-v1alpha1-peer,mutating=true,failurePolicy=fail,groups=ipfs.kotal.io,resources=peers,verbs=create;update,versions=v1alpha1,name=mutate-ipfs-v1alpha1-peer.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Defaulter = &Peer{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Peer) Default() {
	peerlog.Info("default", "name", r.Name)

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
