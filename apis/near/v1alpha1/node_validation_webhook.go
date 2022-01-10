package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-near-kotal-io-v1alpha1-node,mutating=false,failurePolicy=fail,groups=near.kotal.io,resources=nodes,versions=v1alpha1,name=validate-near-v1alpha1-node.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &Node{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateCreate() error {
	nodelog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateUpdate(old runtime.Object) error {
	nodelog.Info("validate update", "name", r.Name)

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateDelete() error {
	nodelog.Info("validate delete", "name", r.Name)

	return nil
}
