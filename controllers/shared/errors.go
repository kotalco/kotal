package shared

import "k8s.io/apimachinery/pkg/api/errors"

// IgnoreConflicts ignore conflict errors
func IgnoreConflicts(err *error) {
	if errors.IsConflict(*err) {
		*err = nil
	}
}
