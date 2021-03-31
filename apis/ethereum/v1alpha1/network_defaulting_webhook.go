package v1alpha1

import (
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:path=/mutate-ethereum-kotal-io-v1alpha1-network,mutating=true,failurePolicy=fail,groups=ethereum.kotal.io,resources=networks,verbs=create;update,versions=v1alpha1,name=mutate-ethereum-v1alpha1-network.kb.io

var _ webhook.Defaulter = &Network{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (n *Network) Default() {
	networklog.Info("default", "name", n.Name)

	if n.Spec.HighlyAvailable {
		if n.Spec.TopologyKey == "" {
			n.Spec.TopologyKey = DefaultTopologyKey
		}
	}

	// default genesis block
	if n.Spec.Genesis != nil {
		n.Spec.Genesis.Default(n.Spec.Consensus)
	}

	// default network nodes
	for i := range n.Spec.Nodes {

		node := &Node{
			Spec: n.Spec.Nodes[i].NodeSpec,
		}
		node.Spec.NetworkConfig = n.Spec.NetworkConfig
		node.Spec.AvailabilityConfig = n.Spec.AvailabilityConfig
		node.Default()

		n.Spec.Nodes[i].NodeSpec = node.Spec
	}

}
