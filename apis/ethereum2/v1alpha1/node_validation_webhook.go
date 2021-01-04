package v1alpha1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-ethereum2-kotal-io-v1alpha1-node,mutating=false,failurePolicy=fail,groups=ethereum2.kotal.io,resources=nodes,versions=v1alpha1,name=vnode.kb.io

var _ webhook.Validator = &Node{}

// Validate is the shared validate create and update logic
func (r *Node) Validate() field.ErrorList {
	var nodeErrors field.ErrorList

	path := field.NewPath("spec")

	if r.Spec.REST && r.Spec.Client != TekuClient && r.Spec.Client != LighthouseClient {
		err := field.Invalid(path.Child("rest"), r.Spec.REST, fmt.Sprintf("not supported by %s client", r.Spec.Client))
		nodeErrors = append(nodeErrors, err)
	}

	return nodeErrors
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateCreate() error {
	var allErrors field.ErrorList

	nodelog.Info("validate create", "name", r.Name)

	allErrors = append(allErrors, r.Validate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)

}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Node) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList

	nodelog.Info("validate update", "name", r.Name)

	allErrors = append(allErrors, r.Validate()...)

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
