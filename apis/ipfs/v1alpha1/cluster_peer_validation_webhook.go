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

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterPeer) ValidateCreate() error {
	var allErrors field.ErrorList

	clusterpeerlog.Info("validate create", "name", r.Name)

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
