package v1alpha1

import "sigs.k8s.io/controller-runtime/pkg/webhook"

// +kubebuilder:webhook:path=/mutate-ethereum2-kotal-io-v1alpha1-node,mutating=true,failurePolicy=fail,groups=ethereum2.kotal.io,resources=nodes,verbs=create;update,versions=v1alpha1,name=mnode.kb.io

var _ webhook.Defaulter = &Node{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Node) Default() {
	nodelog.Info("default", "name", r.Name)

	if r.Spec.Client == "" {
		r.Spec.Client = DefaultClient
	}

	if r.Spec.REST {
		if r.Spec.RESTPort == 0 {
			r.Spec.RESTPort = DefaultRestPort
		}
	}

	if r.Spec.RPC {
		if r.Spec.RPCPort == 0 {
			r.Spec.RPCPort = DefaultRPCPort
		}
	}

	if r.Spec.GRPC {
		if r.Spec.GRPCPort == 0 {
			r.Spec.GRPCPort = DefaultGRPCPort
		}
	}

}
