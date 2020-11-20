package shared

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

// ErrorsToCauses converts error list into array of status cause
func ErrorsToCauses(errs field.ErrorList) []metav1.StatusCause {
	causes := make([]metav1.StatusCause, 0, len(errs))

	for i := range errs {
		err := errs[i]
		causes = append(causes, metav1.StatusCause{
			Type:    metav1.CauseType(err.Type),
			Message: err.ErrorBody(),
			Field:   err.Field,
		})
	}

	return causes
}
