package shared

import (
	"fmt"
	"github.com/kotalco/kotal/helpers/kerrors"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// validate is the shared validation logic
func validate(r *Resources) (errors []*kerrors.KErrors) {
	cpu := r.CPU
	cpuLimit := r.CPULimit

	if cpu != cpuLimit {
		// validate cpuLimit can't be less than cpu request
		cpuQuantity := resource.MustParse(cpu)
		cpuLimitQuantity := resource.MustParse(cpuLimit)
		if cpuLimitQuantity.Cmp(cpuQuantity) == -1 {
			msg := fmt.Sprintf("must be greater than or equal to cpu %s", string(cpu))
			err := field.Invalid(field.NewPath("spec").Child("resources").Child("cpuLimit"), cpuLimit, msg)
			customErr := kerrors.New(*err)
			customErr.ChildField = "cpuLimit"
			customErr.CustomMsg = msg
			errors = append(errors, customErr)
		}
	}

	memory := r.Memory
	memoryLimit := r.MemoryLimit
	memoryQuantity := resource.MustParse(memory)
	memoryLimitQuantity := resource.MustParse(memoryLimit)

	// validate memory limit must be greater than memory
	if memoryLimitQuantity.Cmp(memoryQuantity) != 1 {
		msg := fmt.Sprintf("must be greater than memory %s", string(memory))
		err := field.Invalid(field.NewPath("spec").Child("resources").Child("memoryLimit"), memoryLimit, msg)
		customErr := kerrors.New(*err)
		customErr.ChildField = "memoryLimit"
		customErr.CustomMsg = msg
		errors = append(errors, customErr)
	}

	return
}

// ValidateCreate validates resources during creation
func ValidateCreate(r *Resources) (errors []*kerrors.KErrors) {
	errors = append(errors, validate(r)...)
	return
}

// ValidateUpdate validates resources during update
func ValidateUpdate(r *Resources, oldResources *Resources) (errors []*kerrors.KErrors) {

	oldStorage := oldResources.Storage
	oldStorageClass := oldResources.StorageClass

	errors = append(errors, validate(r)...)

	// requested storage can't be decreased
	if oldStorage != r.Storage {

		oldStorageQuantity := resource.MustParse(oldStorage)
		newStorageQuantity := resource.MustParse(r.Storage)

		if newStorageQuantity.Cmp(oldStorageQuantity) == -1 {
			msg := fmt.Sprintf("must be greater than or equal to old storage %s", oldStorage)
			err := field.Invalid(field.NewPath("spec").Child("resources").Child("storage"), r.Storage, msg)
			customErr := kerrors.New(*err)
			customErr.ChildField = "storage"
			customErr.CustomMsg = msg
			errors = append(errors, customErr)
		}

	}

	// storage class is immutable
	if oldStorageClass != nil && r.StorageClass != nil && *oldStorageClass != *r.StorageClass {
		msg := "field is immutable"
		err := field.Invalid(field.NewPath("spec").Child("resources").Child("storageClass"), *r.StorageClass, msg)
		customErr := kerrors.New(*err)
		customErr.ChildField = "storage"
		customErr.CustomMsg = msg
		errors = append(errors, customErr)
	}

	return
}
