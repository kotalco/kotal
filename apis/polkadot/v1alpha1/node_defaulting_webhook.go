package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate-polkadot-kotal-io-v1alpha1-node,mutating=true,failurePolicy=fail,groups=polkadot.kotal.io,resources=nodes,verbs=create;update,versions=v1alpha1,name=mutate-polkadot-v1alpha1-node.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Defaulter = &Node{}

func (r *Node) DefaultNodeResources() {
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
func (r *Node) Default() {
	nodelog.Info("default", "name", r.Name)

	r.DefaultNodeResources()

	if r.Spec.Image == "" {
		r.Spec.Image = DefaultPolkadotImage
	}

	if r.Spec.SyncMode == "" {
		r.Spec.SyncMode = DefaultSyncMode
	}

	if r.Spec.P2PPort == 0 {
		r.Spec.P2PPort = DefaultP2PPort
	}

	if r.Spec.Pruning == nil {
		t := true
		r.Spec.Pruning = &t
	}

	if r.Spec.RetainedBlocks == 0 {
		r.Spec.RetainedBlocks = DefaultRetainedBlocks
	}

	if r.Spec.Logging == "" {
		r.Spec.Logging = DefaultLoggingVerbosity
	}

	if r.Spec.RPCPort == 0 {
		r.Spec.RPCPort = DefaultRPCPort
	}

	if r.Spec.WSPort == 0 {
		r.Spec.WSPort = DefaultWSPort
	}

	if len(r.Spec.CORSDomains) == 0 {
		r.Spec.CORSDomains = []string{DefaultCORSDomain}
	}

	if r.Spec.TelemetryURL == "" {
		r.Spec.TelemetryURL = DefaultTelemetryURL
	}

	if r.Spec.PrometheusPort == 0 {
		r.Spec.PrometheusPort = DefaultPrometheusPort
	}

}
