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
	GetLabels() map[string]string
}

// UpdateLabels adds missing labels to the resource
func UpdateLabels(cr CustomResource, client, network string) {

	gvk := cr.GroupVersionKind()
	group := strings.Replace(gvk.Group, ".kotal.io", "", 1)
	kind := strings.ToLower(gvk.Kind)

	labels := cr.GetLabels()
	if labels == nil {
		labels = map[string]string{}
	}

	labels["app.kubernetes.io/name"] = client
	labels["app.kubernetes.io/instance"] = cr.GetName()
	labels["app.kubernetes.io/component"] = fmt.Sprintf("%s-%s", group, kind)
	labels["app.kubernetes.io/managed-by"] = "kotal-operator"
	labels["app.kubernetes.io/created-by"] = fmt.Sprintf("%s-%s-controller", group, kind)
	labels["kotal.io/protocol"] = group
	if network != "" {
		labels["kotal.io/network"] = network
	}

	cr.SetLabels(labels)

}
