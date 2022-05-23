package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-aptos-kotal-io-v1alpha1-node,mutating=false,failurePolicy=fail,groups=aptos.kotal.io,resources=nodes,versions=v1alpha1,name=validate-aptos-v1alpha1-node.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &Node{}

// validate shared validation logic for create and update resources
func (r *Node) validate() field.ErrorList {
	var nodeErrors field.ErrorList

	if r.Spec.NodePrivateKeySecretName != "" && r.Spec.PeerId == "" {
		err := field.Invalid(field.NewPath("spec").Child("peerId"), r.Spec.PeerId, "must provide peerId if nodePrivateKeySecretName is provided")
		nodeErrors = append(nodeErrors, err)
	}

	if r.Spec.PeerId != "" && r.Spec.NodePrivateKeySecretName == "" {
		err := field.Invalid(field.NewPath("spec").Child("nodePrivateKeySecretName"), r.Spec.NodePrivateKeySecretName, "must provide nodePrivateKeySecretName if peerId is provided")
		nodeErrors = append(nodeErrors, err)
	}

	return nodeErrors
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateCreate() error {
	var allErrors field.ErrorList

	nodelog.Info("validate create", "name", r.Name)

	allErrors = append(allErrors, r.validate()...)
	allErrors = append(allErrors, r.Spec.Resources.ValidateCreate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList
	oldNode := old.(*Node)

	nodelog.Info("validate update", "name", r.Name)

	allErrors = append(allErrors, r.validate()...)
	allErrors = append(allErrors, r.Spec.Resources.ValidateUpdate(&oldNode.Spec.Resources)...)

	if r.Spec.Network != oldNode.Spec.Network {
		err := field.Invalid(field.NewPath("spec").Child("network"), r.Spec.Network, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateDelete() error {
	nodelog.Info("validate delete", "name", r.Name)

	return nil
}
