package v1alpha1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-chainlink-kotal-io-v1alpha1-node,mutating=false,failurePolicy=fail,groups=chainlink.kotal.io,resources=nodes,versions=v1alpha1,name=validate-chainlink-v1alpha1-node.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &Node{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateCreate() (admission.Warnings, error) {
	var allErrors field.ErrorList

	nodelog.Info("validate create", "name", r.Name)

	allErrors = append(allErrors, r.Spec.Resources.ValidateCreate()...)

	if len(allErrors) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)

}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	var allErrors field.ErrorList
	oldNode := old.(*Node)

	nodelog.Info("validate update", "name", r.Name)

	allErrors = append(allErrors, r.Spec.Resources.ValidateUpdate(&oldNode.Spec.Resources)...)

	if oldNode.Spec.EthereumChainId != r.Spec.EthereumChainId {
		err := field.Invalid(field.NewPath("spec").Child("ethereumChainId"), fmt.Sprintf("%d", r.Spec.EthereumChainId), "field is immutable")
		allErrors = append(allErrors, err)
	}

	if oldNode.Spec.LinkContractAddress != r.Spec.LinkContractAddress {
		err := field.Invalid(field.NewPath("spec").Child("linkContractAddress"), r.Spec.LinkContractAddress, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if len(allErrors) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateDelete() (admission.Warnings, error) {
	nodelog.Info("validate delete", "name", r.Name)
	return nil, nil
}
