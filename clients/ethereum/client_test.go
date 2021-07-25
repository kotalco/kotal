package ethereum

import (
	"testing"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestNewClient(t *testing.T) {
	node := &ethereumv1alpha1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node-with-invalid-client",
		},
		Spec: ethereumv1alpha1.NodeSpec{
			Client:  ethereumv1alpha1.EthereumClient("nokia"),
			Network: ethereumv1alpha1.MainNetwork,
		},
	}

	client, err := NewClient(node)
	if err == nil {
		t.Error("expecting an error")
	}

	if client != nil {
		t.Error("expecting client to be nil")
	}

	expected := "client nokia is not supported"
	got := err.Error()
	if expected != got {
		t.Errorf("expected error message to be: %s", expected)
	}
}
