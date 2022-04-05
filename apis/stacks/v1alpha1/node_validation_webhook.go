package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-stacks-kotal-io-v1alpha1-node,mutating=false,failurePolicy=fail,groups=stacks.kotal.io,resources=nodes,versions=v1alpha1,name=validate-stacks-v1alpha1-node.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &Node{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateCreate() error {
	var allErrors field.ErrorList

	nodelog.Info("validate create", "name", r.Name)

	allErrors = append(allErrors, r.Spec.Resources.ValidateCreate()...)

	if r.Spec.Miner && r.Spec.SeedPrivateKeySecretName == "" {
		err := field.Invalid(field.NewPath("spec").Child("seedPrivateKeySecretName"), r.Spec.SeedPrivateKeySecretName, "seedPrivateKeySecretName is required if node is miner")
		allErrors = append(allErrors, err)
	}

	if r.Spec.SeedPrivateKeySecretName != "" && !r.Spec.Miner {
		err := field.Invalid(field.NewPath("spec").Child("miner"), r.Spec.Miner, "node must be a miner if seedPrivateKeySecretName is given")
		allErrors = append(allErrors, err)
	}

	if r.Spec.MineMicroblocks && !r.Spec.Miner {
		err := field.Invalid(field.NewPath("spec").Child("miner"), r.Spec.Miner, "node must be a miner if mineMicroblocks is true")
		allErrors = append(allErrors, err)
	}

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
