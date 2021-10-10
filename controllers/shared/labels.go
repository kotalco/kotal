package shared

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CustomResource interface {
	GroupVersionKind() schema.GroupVersionKind
	GetName() string
	SetLabels(map[string]string)
}

// UpdateLabels adds missing labels to the resource
func UpdateLabels(cr CustomResource, client string) {

	gvk := cr.GroupVersionKind()
	group := strings.Replace(gvk.Group, ".kotal.io", "", 1)
	kind := strings.ToLower(gvk.Kind)

	labels := map[string]string{
		"app.kubernetes.io/name":       client,
		"app.kubernetes.io/instance":   cr.GetName(),
		"app.kubernetes.io/component":  fmt.Sprintf("%s-%s", group, kind),
		"app.kubernetes.io/managed-by": "kotal",
		"app.kubernetes.io/created-by": fmt.Sprintf("%s-%s-controller", group, kind),
	}

	cr.SetLabels(labels)

}
