package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate-near-kotal-io-v1alpha1-node,mutating=true,failurePolicy=fail,groups=near.kotal.io,resources=nodes,verbs=create;update,versions=v1alpha1,name=mutate-near-v1alpha1-node.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Defaulter = &Node{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (n *Node) Default() {
	nodelog.Info("default", "name", n.Name)

	if n.Spec.Image == "" {
		n.Spec.Image = DefaultNearImage
	}

	if n.Spec.MinPeers == 0 {
		n.Spec.MinPeers = DefaultMinPeers
	}

	if n.Spec.RPCPort == 0 {
		n.Spec.RPCPort = DefaultRPCPort
	}

	if n.Spec.PrometheusPort == 0 {
		n.Spec.PrometheusPort = DefaultPrometheusPort
	}

	if n.Spec.P2PPort == 0 {
		n.Spec.P2PPort = DefaultP2PPort
	}

	if n.Spec.CPU == "" {
		n.Spec.CPU = DefaultNodeCPURequest
	}
	if n.Spec.CPULimit == "" {
		n.Spec.CPULimit = DefaultNodeCPULimit
	}

	if n.Spec.Memory == "" {
		n.Spec.Memory = DefaultNodeMemoryRequest
	}
	if n.Spec.MemoryLimit == "" {
		n.Spec.MemoryLimit = DefaultNodeMemoryLimit
	}

	if n.Spec.Storage == "" {
		storage := DefaultNodeStorageRequest
		if n.Spec.Archive {
			storage = DefaultArchivalNodeStorageRequest
		}
		n.Spec.Storage = storage
	}

}
