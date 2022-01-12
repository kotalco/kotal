package near

import (
	nearv1alpha1 "github.com/kotalco/kotal/apis/near/v1alpha1"
	clients "github.com/kotalco/kotal/clients"
)

func NewClient(node *nearv1alpha1.Node) clients.Interface {
	return &NearClient{node}
}
