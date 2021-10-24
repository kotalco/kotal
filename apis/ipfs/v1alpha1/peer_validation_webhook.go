package v1alpha1

import (
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-ipfs-kotal-io-v1alpha1-peer,mutating=false,failurePolicy=fail,groups=ipfs.kotal.io,resources=peers,versions=v1alpha1,name=validate-ipfs-v1alpha1-peer.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &Peer{}

// ValidateCreate valdates ipfs peers during their creation
func (p *Peer) ValidateCreate() error {
	var allErrors field.ErrorList

	peerlog.Info("validate create", "name", p.Name)

	allErrors = append(allErrors, p.Spec.Resources.ValidateCreate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, p.Name, allErrors)
}

// initProfilesChanged returns true if initial profiles changed
func initProfilesChanged(old, peer *Peer) bool {
	for i, profile := range old.Spec.InitProfiles {
		if peer.Spec.InitProfiles[i] != profile {
			return true
		}
	}
	return false
}

// ValidateUpdate validates ipfs peers while being updated
func (p *Peer) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList
	oldPeer := old.(*Peer)

	peerlog.Info("validate update", "name", p.Name)

	if oldPeer.Spec.SwarmKeySecretName != p.Spec.SwarmKeySecretName {
		err := field.Invalid(field.NewPath("spec").Child("swarmKeySecretName"), p.Spec.SwarmKeySecretName, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if len(oldPeer.Spec.InitProfiles) != len(p.Spec.InitProfiles) || initProfilesChanged(oldPeer, p) {
		profiles := []string{}
		for _, profile := range p.Spec.InitProfiles {
			profiles = append(profiles, string(profile))
		}
		err := field.Invalid(field.NewPath("spec").Child("initProfiles"), strings.Join(profiles, ","), "field is immutable")
		allErrors = append(allErrors, err)
	}

	allErrors = append(allErrors, p.Spec.Resources.ValidateUpdate(&oldPeer.Spec.Resources)...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, p.Name, allErrors)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (p *Peer) ValidateDelete() error {
	peerlog.Info("validate delete", "name", p.Name)

	return nil
}
