package shared

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// Resources is node compute and storage resources
// +k8s:deepcopy-gen=true
type Resources struct {
	// CPU is cpu cores the node requires
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*m?$"
	CPU string `json:"cpu,omitempty"`
	// CPULimit is cpu cores the node is limited to
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*m?$"
	CPULimit string `json:"cpuLimit,omitempty"`
	// Memory is memmory requirements
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*[KMGTPE]i$"
	Memory string `json:"memory,omitempty"`
	// MemoryLimit is cpu cores the node is limited to
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*[KMGTPE]i$"
	MemoryLimit string `json:"memoryLimit,omitempty"`
	// Storage is disk space storage requirements
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*[KMGTPE]i$"
	Storage string `json:"storage,omitempty"`
	// StorageClass is the volume storage class
	StorageClass *string `json:"storageClass,omitempty"`
}

// validate is the shared validation logic
func (r *Resources) validate() (errors field.ErrorList) {
	cpu := r.CPU
	cpuLimit := r.CPULimit

	if cpu != cpuLimit {
		// validate cpuLimit can't be less than cpu request
		cpuQuantity := resource.MustParse(cpu)
		cpuLimitQuantity := resource.MustParse(cpuLimit)
		if cpuLimitQuantity.Cmp(cpuQuantity) == -1 {
			msg := fmt.Sprintf("must be greater than or equal to cpu %s", string(cpu))
			err := field.Invalid(field.NewPath("spec").Child("resources").Child("cpuLimit"), cpuLimit, msg)
			errors = append(errors, err)
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
		errors = append(errors, err)
	}

	return
}

// ValidateCreate validates resources during creation
func (r *Resources) ValidateCreate() (errors field.ErrorList) {
	errors = append(errors, r.validate()...)
	return
}

// ValidateUpdate validates resources during update
func (r *Resources) ValidateUpdate(oldResources *Resources) (errors field.ErrorList) {

	oldStorage := oldResources.Storage
	oldStorageClass := oldResources.StorageClass

	errors = append(errors, r.validate()...)

	// requested storage can't be decreased
	if oldStorage != r.Storage {

		oldStorageQuantity := resource.MustParse(oldStorage)
		newStorageQuantity := resource.MustParse(r.Storage)

		if newStorageQuantity.Cmp(oldStorageQuantity) == -1 {
			msg := fmt.Sprintf("must be greater than or equal to old storage %s", oldStorage)
			err := field.Invalid(field.NewPath("spec").Child("resources").Child("storage"), r.Storage, msg)
			errors = append(errors, err)
		}

	}

	// storage class is immutable
	if oldStorageClass != nil && r.StorageClass != nil && *oldStorageClass != *r.StorageClass {
		msg := "field is immutable"
		err := field.Invalid(field.NewPath("spec").Child("resources").Child("storageClass"), *r.StorageClass, msg)
		errors = append(errors, err)
	}

	return
}
