package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-near-kotal-io-v1alpha1-node,mutating=false,failurePolicy=fail,groups=near.kotal.io,resources=nodes,versions=v1alpha1,name=validate-near-v1alpha1-node.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &Node{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (n *Node) ValidateCreate() error {
	nodelog.Info("validate create", "name", n.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (n *Node) ValidateUpdate(old runtime.Object) error {
	nodelog.Info("validate update", "name", n.Name)

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (n *Node) ValidateDelete() error {
	nodelog.Info("validate delete", "name", n.Name)

	return nil
}
