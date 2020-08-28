package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var network = &Network{
	ObjectMeta: metav1.ObjectMeta{
		Name: "test-network",
	},
	Spec: NetworkSpec{
		Join: RinkebyNetwork,
		Nodes: []Node{
			{
				Name:    "node-1",
				Client:  BesuClient,
				Nodekey: PrivateKey("0xd42baddc4e6072f670da327e2ebc835d695bb9b911642dd70b81d522f0afe90c"),
			},
		},
	},
}

func TestIsBootnode(t *testing.T) {
	node := network.Spec.Nodes[0]
	expected := false
	got := node.IsBootnode()

	if got != expected {
		t.Errorf("Expecting bootnode to be %t got %t", expected, got)
	}

}

func TestWithNodekey(t *testing.T) {
	node := network.Spec.Nodes[0]
	expected := true
	got := node.WithNodekey()

	if got != expected {
		t.Errorf("Expecting bootnode to be %t got %t", expected, got)
	}
}

func TestDeploymentName(t *testing.T) {
	node := network.Spec.Nodes[0]
	expected := "test-network-node-1"
	got := node.DeploymentName(network.Name)

	if got != expected {
		t.Errorf("Expecting bootnode to be %s got %s", expected, got)
	}
}

func TestPVCName(t *testing.T) {
	node := network.Spec.Nodes[0]
	expected := "test-network-node-1"
	got := node.PVCName(network.Name)

	if got != expected {
		t.Errorf("Expecting bootnode to be %s got %s", expected, got)
	}
}

func TestSecretName(t *testing.T) {
	node := network.Spec.Nodes[0]
	expected := "test-network-node-1"
	got := node.SecretName(network.Name)

	if got != expected {
		t.Errorf("Expecting bootnode to be %s got %s", expected, got)
	}
}

func TestServiceName(t *testing.T) {
	node := network.Spec.Nodes[0]
	expected := "test-network-node-1"
	got := node.ServiceName(network.Name)

	if got != expected {
		t.Errorf("Expecting bootnode to be %s got %s", expected, got)
	}
}

func TestConfigmapName(t *testing.T) {
	node := network.Spec.Nodes[0]
	expected := "test-network-besu"
	got := node.ConfigmapName(network.Name, node.Client)

	if got != expected {
		t.Errorf("Expecting bootnode to be %s got %s", expected, got)
	}
}

func TestLabels(t *testing.T) {
	node := network.Spec.Nodes[0]
	expected := map[string]string{
		"name":     "node",
		"instance": "node-1",
		"network":  "test-network",
	}
	got := node.Labels(network.Name)

	if !reflect.DeepEqual(got, expected) {
		t.Errorf("Expecting node labels to be %s got %s", expected, got)
	}
}
