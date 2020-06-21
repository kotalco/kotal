package v1alpha1

import "testing"

func TestIsBootnode(t *testing.T) {
	n := &Network{
		Spec: NetworkSpec{
			Join: RinkebyNetwork,
			Nodes: []Node{
				{
					Name: "node-1",
				},
			},
		},
	}
	expected := false
	got := n.Spec.Nodes[0].IsBootnode()

	if got != expected {
		t.Errorf("Expecting bootnode to be %t got %t", expected, got)
	}

}

func TestWithNodekey(t *testing.T) {
	n := &Network{
		Spec: NetworkSpec{
			Join: RinkebyNetwork,
			Nodes: []Node{
				{
					Name:    "node-1",
					Nodekey: PrivateKey("0xd42baddc4e6072f670da327e2ebc835d695bb9b911642dd70b81d522f0afe90c"),
				},
			},
		},
	}
	expected := true
	got := n.Spec.Nodes[0].WithNodekey()

	if got != expected {
		t.Errorf("Expecting bootnode to be %t got %t", expected, got)
	}
}
