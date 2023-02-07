package shared

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

var _ = Describe("Resource validation", func() {
	createCases := []struct {
		Title     string
		Resources *Resources
		Errors    field.ErrorList
	}{
		{
			Title: "invalid cpu limit value",
			Resources: &Resources{
				CPU:         "2",
				CPULimit:    "1",
				Memory:      "1Gi",
				MemoryLimit: "2Gi",
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.resources.cpuLimit",
					BadValue: "1",
					Detail:   "must be greater than or equal to cpu 2",
				},
			},
		},
		{
			Title: "invalid memory limit value",
			Resources: &Resources{
				CPU:         "1",
				CPULimit:    "2",
				Memory:      "2Gi",
				MemoryLimit: "1Gi",
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.resources.memoryLimit",
					BadValue: "1Gi",
					Detail:   "must be greater than memory 2Gi",
				},
			},
		},
	}

	storageClass := "standard"
	newStorageClass := "custom"

	updateCases := []struct {
		Title        string
		OldResources *Resources
		NewResources *Resources
		Errors       field.ErrorList
	}{
		{
			Title: "invalid new cpu limit value",
			OldResources: &Resources{
				CPU:         "1",
				CPULimit:    "2",
				Memory:      "1Gi",
				MemoryLimit: "2Gi",
			},
			NewResources: &Resources{
				CPU:         "2",
				CPULimit:    "1",
				Memory:      "1Gi",
				MemoryLimit: "2Gi",
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.resources.cpuLimit",
					BadValue: "1",
					Detail:   "must be greater than or equal to cpu 2",
				},
			},
		},
		{
			Title: "invalid new memory limit value",
			OldResources: &Resources{
				CPU:         "1",
				CPULimit:    "2",
				Memory:      "1Gi",
				MemoryLimit: "2Gi",
			},
			NewResources: &Resources{
				CPU:         "1",
				CPULimit:    "2",
				Memory:      "2Gi",
				MemoryLimit: "1Gi",
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.resources.memoryLimit",
					BadValue: "1Gi",
					Detail:   "must be greater than memory 2Gi",
				},
			},
		},
		{
			Title: "invalid new storage class value",
			OldResources: &Resources{
				CPU:          "1",
				CPULimit:     "2",
				Memory:       "1Gi",
				MemoryLimit:  "2Gi",
				StorageClass: &storageClass,
			},
			NewResources: &Resources{
				CPU:          "1",
				CPULimit:     "2",
				Memory:       "1Gi",
				MemoryLimit:  "2Gi",
				StorageClass: &newStorageClass,
			},
			Errors: field.ErrorList{
				{
					Type:     field.ErrorTypeInvalid,
					Field:    "spec.resources.storageClass",
					BadValue: "custom",
					Detail:   "field is immutable",
				},
			},
		},
	}

	Context("While creating node", func() {
		for _, c := range createCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					errorList := cc.Resources.ValidateCreate()
					Expect(errorList).To(ContainElements(cc.Errors))
				})
			}()
		}
	})

	Context("While updating node", func() {
		for _, c := range updateCases {
			func() {
				cc := c
				It(fmt.Sprintf("Should validate %s", cc.Title), func() {
					errorList := cc.NewResources.ValidateUpdate(cc.OldResources)
					Expect(errorList).To(ContainElements(cc.Errors))
				})
			}()
		}
	})

})
