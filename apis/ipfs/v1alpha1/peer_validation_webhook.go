package v1alpha1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-ipfs-kotal-io-v1alpha1-peer,mutating=false,failurePolicy=fail,groups=ipfs.kotal.io,resources=peers,versions=v1alpha1,name=validate-ipfs-v1alpha1-peer.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &Peer{}

// Validate is the common validation for create and update
func (p *Peer) Validate() (peerErrors field.ErrorList) {

	cpu := p.Spec.Resources.CPU
	cpuLimit := p.Spec.Resources.CPULimit

	if cpu != cpuLimit {
		// validate cpuLimit can't be less than cpu request
		cpuQuantity := resource.MustParse(cpu)
		cpuLimitQuantity := resource.MustParse(cpuLimit)
		if cpuLimitQuantity.Cmp(cpuQuantity) == -1 {
			msg := fmt.Sprintf("must be greater than or equal to cpu %s", string(cpu))
			err := field.Invalid(field.NewPath("spec").Child("resources").Child("cpuLimit"), cpuLimit, msg)
			peerErrors = append(peerErrors, err)
		}
	}

	memory := p.Spec.Resources.Memory
	memoryLimit := p.Spec.Resources.MemoryLimit

	// validate memory and memory limit can't be equal
	if memory == memoryLimit {
		msg := fmt.Sprintf("must be greater than memory %s", string(memory))
		err := field.Invalid(field.NewPath("spec").Child("resources").Child("memoryLimit"), memoryLimit, msg)
		peerErrors = append(peerErrors, err)
	} else {
		// validate memoryLimit can't be less than memory request
		memoryQuantity := resource.MustParse(memory)
		memoryLimitQuantity := resource.MustParse(memoryLimit)

		if memoryLimitQuantity.Cmp(memoryQuantity) == -1 {
			msg := fmt.Sprintf("must be greater than memory %s", string(memory))
			err := field.Invalid(field.NewPath("spec").Child("resources").Child("memoryLimit"), memoryLimit, msg)
			peerErrors = append(peerErrors, err)
		}
	}

	return peerErrors
}

// ValidateCreate valdates ipfs peers during their creation
func (p *Peer) ValidateCreate() error {
	var allErrors field.ErrorList

	peerlog.Info("validate create", "name", p.Name)

	allErrors = append(allErrors, p.Validate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, p.Name, allErrors)
}

// ValidateUpdate validates ipfs peers while being updated
func (p *Peer) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList

	peerlog.Info("validate update", "name", p.Name)

	allErrors = append(allErrors, p.Validate()...)

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
