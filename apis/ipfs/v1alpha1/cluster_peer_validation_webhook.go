package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-ipfs-kotal-io-v1alpha1-clusterpeer,mutating=false,failurePolicy=fail,groups=ipfs.kotal.io,resources=clusterpeers,versions=v1alpha1,name=validate-ipfs-v1alpha1-clusterpeer.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &ClusterPeer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterPeer) ValidateCreate() error {
	clusterpeerlog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterPeer) ValidateUpdate(old runtime.Object) error {
	clusterpeerlog.Info("validate update", "name", r.Name)

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterPeer) ValidateDelete() error {
	clusterpeerlog.Info("validate delete", "name", r.Name)

	return nil
}
