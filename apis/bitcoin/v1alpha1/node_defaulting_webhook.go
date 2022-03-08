package v1alpha1

import "sigs.k8s.io/controller-runtime/pkg/webhook"

// +kubebuilder:webhook:path=/mutate-bitcoin-kotal-io-v1alpha1-node,mutating=true,failurePolicy=fail,groups=bitcoin.kotal.io,resources=nodes,verbs=create;update,versions=v1alpha1,name=mutate-bitcoin-v1alpha1-node.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Defaulter = &Node{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Node) Default() {
	nodelog.Info("default", "name", r.Name)

	if r.Spec.RPCPort == 0 {
		if r.Spec.Network == Mainnet {
			r.Spec.RPCPort = DefaultMainnetRPCPort
		}
		if r.Spec.Network == Testnet {
			r.Spec.RPCPort = DefaultTestnetRPCPort
		}
	}

}
