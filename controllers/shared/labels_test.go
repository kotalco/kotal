package shared

import (
	"testing"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
)

func TestUpdateLabels(t *testing.T) {

	if err := ethereumv1alpha1.AddToScheme(scheme.Scheme); err != nil {
		t.Error(err)
	}

	ethereumNode := ethereumv1alpha1.Node{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Node",
			APIVersion: "ethereum/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-node",
			Namespace: "default",
		},
		Spec: ethereumv1alpha1.NodeSpec{
			Client:  ethereumv1alpha1.BesuClient,
			Network: ethereumv1alpha1.GoerliNetwork,
		},
	}

	UpdateLabels(&ethereumNode, string(ethereumNode.Spec.Client), ethereumv1alpha1.GoerliNetwork)

	labels := map[string]string{
		"app.kubernetes.io/name":       string(ethereumNode.Spec.Client),
		"app.kubernetes.io/instance":   ethereumNode.Name,
		"app.kubernetes.io/component":  "ethereum-node",
		"app.kubernetes.io/managed-by": "kotal-operator",
		"app.kubernetes.io/created-by": "ethereum-node-controller",
		"kotal.io/protocol":            "ethereum",
		"kotal.io/network":             ethereumv1alpha1.GoerliNetwork,
	}

	for k, v := range labels {
		if ethereumNode.Labels[k] != v {
			t.Errorf("Expecting label with key %s to have value %s, but got %s", k, v, ethereumNode.Labels[k])
		}
	}

}
