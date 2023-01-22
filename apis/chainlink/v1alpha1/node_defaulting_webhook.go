package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate-chainlink-kotal-io-v1alpha1-node,mutating=true,failurePolicy=fail,groups=chainlink.kotal.io,resources=nodes,verbs=create;update,versions=v1alpha1,name=mutate-chainlink-v1alpha1-node.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Defaulter = &Node{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Node) Default() {
	nodelog.Info("default", "name", r.Name)

	if r.Spec.Image == "" {
		r.Spec.Image = DefaultChainlinkImage
	}

	if r.Spec.P2PPort == 0 {
		r.Spec.P2PPort = DefaultP2PPort
	}

	if r.Spec.APIPort == 0 {
		r.Spec.APIPort = DefaultAPIPort
	}

	if r.Spec.CPU == "" {
		r.Spec.CPU = DefaultNodeCPURequest
	}

	if r.Spec.CPULimit == "" {
		r.Spec.CPULimit = DefaultNodeCPULimit
	}

	if r.Spec.Memory == "" {
		r.Spec.Memory = DefaultNodeMemoryRequest
	}

	if r.Spec.MemoryLimit == "" {
		r.Spec.MemoryLimit = DefaultNodeMemoryLimit
	}

	if r.Spec.Storage == "" {
		r.Spec.Storage = DefaultNodeStorageRequest
	}

	if r.Spec.TLSPort == 0 {
		r.Spec.TLSPort = DefaultTLSPort
	}

	if r.Spec.Logging == "" {
		r.Spec.Logging = shared.InfoLogs
	}

	if len(r.Spec.CORSDomains) == 0 {
		r.Spec.CORSDomains = DefaultCorsDomains
	}

}
