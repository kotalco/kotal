package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-ipfs-kotal-io-v1alpha1-clusterpeer,mutating=false,failurePolicy=fail,groups=ipfs.kotal.io,resources=clusterpeers,versions=v1alpha1,name=validate-ipfs-v1alpha1-clusterpeer.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &ClusterPeer{}

// validate validates a node with a given path
func (r *ClusterPeer) validate() field.ErrorList {
	var nodeErrors field.ErrorList

	// privateKeySecretName is required if id is given
	if r.Spec.ID != "" && r.Spec.PrivateKeySecretName == "" {
		err := field.Invalid(field.NewPath("spec").Child("privateKeySecretName"), "", "must provide privateKeySecretName if id is provided")
		nodeErrors = append(nodeErrors, err)
	}

	// id is required if privateKeySecretName is given
	if r.Spec.PrivateKeySecretName != "" && r.Spec.ID == "" {
		err := field.Invalid(field.NewPath("spec").Child("id"), "", "must provide id if privateKeySecretName is provided")
		nodeErrors = append(nodeErrors, err)
	}

	return nodeErrors

}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterPeer) ValidateCreate() error {
	var allErrors field.ErrorList

	clusterpeerlog.Info("validate create", "name", r.Name)

	allErrors = append(allErrors, r.validate()...)
	allErrors = append(allErrors, r.Spec.Resources.ValidateCreate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterPeer) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList
	oldClusterPeer := old.(*ClusterPeer)

	clusterpeerlog.Info("validate update", "name", r.Name)

	if oldClusterPeer.Spec.Consensus != r.Spec.Consensus {
		err := field.Invalid(field.NewPath("spec").Child("consensus"), r.Spec.Consensus, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if oldClusterPeer.Spec.ID != r.Spec.ID {
		err := field.Invalid(field.NewPath("spec").Child("id"), r.Spec.ID, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if oldClusterPeer.Spec.PrivateKeySecretName != r.Spec.PrivateKeySecretName {
		err := field.Invalid(field.NewPath("spec").Child("privateKeySecretName"), r.Spec.PrivateKeySecretName, "field is immutable")
		allErrors = append(allErrors, err)
	}

	allErrors = append(allErrors, r.validate()...)
	allErrors = append(allErrors, r.Spec.Resources.ValidateUpdate(&oldClusterPeer.Spec.Resources)...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterPeer) ValidateDelete() error {
	clusterpeerlog.Info("validate delete", "name", r.Name)

	return nil
}
