package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-ipfs-kotal-io-v1alpha1-peer,mutating=false,failurePolicy=fail,groups=ipfs.kotal.io,resources=peers,versions=v1alpha1,name=validate-ipfs-v1alpha1-peer.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &Peer{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Peer) ValidateCreate() error {
	peerlog.Info("validate create", "name", r.Name)

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Peer) ValidateUpdate(old runtime.Object) error {
	peerlog.Info("validate update", "name", r.Name)

	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Peer) ValidateDelete() error {
	peerlog.Info("validate delete", "name", r.Name)

	return nil
}
