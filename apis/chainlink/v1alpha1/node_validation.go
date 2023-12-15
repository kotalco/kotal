package v1alpha1

import (
	"fmt"
	"github.com/kotalco/kotal/apis/shared"
	"github.com/kotalco/kotal/helpers/kerrors"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func ValidateCreate(n *Node) []*kerrors.KErrors {
	var allErrors []*kerrors.KErrors

	nodelog.Info("validate create", "name", n.Name)

	allErrors = append(allErrors, shared.ValidateCreate(&n.Spec.Resources)...)

	if len(allErrors) == 0 {
		return nil
	}

	return allErrors
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func ValidateUpdate(n *Node, old runtime.Object) (errors []*kerrors.KErrors) {
	oldNode := old.(*Node)

	nodelog.Info("validate update", "name", n.Name)

	errors = append(errors, shared.ValidateUpdate(&n.Spec.Resources, &oldNode.Spec.Resources)...)

	if oldNode.Spec.EthereumChainId != n.Spec.EthereumChainId {
		err := field.Invalid(field.NewPath("spec").Child("ethereumChainId"), fmt.Sprintf("%d", n.Spec.EthereumChainId), "field is immutable")
		customErr := kerrors.New(*err)
		customErr.ChildField = "ethereumChainId"
		customErr.CustomMsg = err.Detail
		errors = append(errors, customErr)
	}

	if oldNode.Spec.LinkContractAddress != n.Spec.LinkContractAddress {
		err := field.Invalid(field.NewPath("spec").Child("linkContractAddress"), n.Spec.LinkContractAddress, "field is immutable")
		customErr := kerrors.New(*err)
		customErr.ChildField = "linkContractAddress"
		customErr.CustomMsg = err.Detail
		errors = append(errors, customErr)
	}

	return
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func ValidateDelete(n *Node) []*kerrors.KErrors {
	nodelog.Info("validate delete", "name", n.Name)
	return nil
}
