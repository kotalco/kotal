package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var swarmlog = logf.Log.WithName("swarm-resource")

// SetupWebhookWithManager registers webhook to be started wth the given manager
func (r *Swarm) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-ipfs-kotal-io-v1alpha1-swarm,mutating=true,failurePolicy=fail,groups=ipfs.kotal.io,resources=swarms,verbs=create;update,versions=v1alpha1,name=mswarm.kb.io

var _ webhook.Defaulter = &Swarm{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Swarm) Default() {
	swarmlog.Info("default", "name", r.Name)

	for i := range r.Spec.Nodes {
		r.DefaultNode(&r.Spec.Nodes[i])
	}
}

// DefaultNode defaults a single ipfs node spec
func (r *Swarm) DefaultNode(node *Node) {
	if node.Resources == nil {
		node.Resources = &NodeResources{}
	}

	if node.Resources.CPU == "" {
		node.Resources.CPU = DefaultNodeCPURequest
	}

	if node.Resources.CPULimit == "" {
		node.Resources.CPULimit = DefaultNodeCPULimit
	}

	if node.Resources.Memory == "" {
		node.Resources.Memory = DefaultNodeMemoryRequest
	}

	if node.Resources.MemoryLimit == "" {
		node.Resources.MemoryLimit = DefaultNodeMemoryLimit
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-ipfs-kotal-io-v1alpha1-swarm,mutating=false,failurePolicy=fail,groups=ipfs.kotal.io,resources=swarms,versions=v1alpha1,name=vswarm.kb.io

var _ webhook.Validator = &Swarm{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Swarm) ValidateCreate() error {
	swarmlog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Swarm) ValidateUpdate(old runtime.Object) error {
	swarmlog.Info("validate update", "name", r.Name)

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Swarm) ValidateDelete() error {
	swarmlog.Info("validate delete", "name", r.Name)

	return nil
}
