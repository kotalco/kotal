package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-near-kotal-io-v1alpha1-node,mutating=false,failurePolicy=fail,groups=near.kotal.io,resources=nodes,versions=v1alpha1,name=validate-near-v1alpha1-node.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &Node{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (n *Node) ValidateCreate() error {
	var allErrors field.ErrorList

	nodelog.Info("validate create", "name", n.Name)

	allErrors = append(allErrors, n.Spec.Resources.ValidateCreate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, n.Name, allErrors)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (n *Node) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList
	oldNode := old.(*Node)

	nodelog.Info("validate update", "name", n.Name)

	allErrors = append(allErrors, n.Spec.Resources.ValidateUpdate(&oldNode.Spec.Resources)...)

	if n.Spec.Network != oldNode.Spec.Network {
		err := field.Invalid(field.NewPath("spec").Child("network"), n.Spec.Network, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, n.Name, allErrors)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (n *Node) ValidateDelete() error {
	nodelog.Info("validate delete", "name", n.Name)

	return nil
}
