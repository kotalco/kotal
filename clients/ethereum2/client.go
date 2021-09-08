package ethereum2

import (
	"fmt"

	ethereum2v1alpha1 "github.com/kotalco/kotal/apis/ethereum2/v1alpha1"
	"github.com/kotalco/kotal/clients"
	"k8s.io/apimachinery/pkg/runtime"
)

// Ethereum2Client is Ethereum 2.0 beacon node or validator client
type Ethereum2Client interface {
	clients.Interface
}

// NewClient creates new ethereum 2.0 beacon node or validator client
func NewClient(obj runtime.Object) (Ethereum2Client, error) {

	switch component := obj.(type) {

	// create beacon nodes
	case *ethereum2v1alpha1.BeaconNode:
		switch component.Spec.Client {
		case ethereum2v1alpha1.TekuClient:
			return &TekuBeaconNode{component}, nil
		case ethereum2v1alpha1.PrysmClient:
			return &PrysmBeaconNode{component}, nil
		case ethereum2v1alpha1.LighthouseClient:
			return &LighthouseBeaconNode{component}, nil
		case ethereum2v1alpha1.NimbusClient:
			return &NimbusBeaconNode{component}, nil
		default:
			return nil, fmt.Errorf("client %s is not supported", component.Spec.Client)
		}

	// create validator clients
	case *ethereum2v1alpha1.Validator:
		switch component.Spec.Client {
		case ethereum2v1alpha1.TekuClient:
			return &TekuValidatorClient{component}, nil
		case ethereum2v1alpha1.PrysmClient:
			return &PrysmValidatorClient{component}, nil
		case ethereum2v1alpha1.LighthouseClient:
			return &LighthouseValidatorClient{component}, nil
		case ethereum2v1alpha1.NimbusClient:
			return &NimbusValidatorClient{component}, nil
		default:
			return nil, fmt.Errorf("client %s is not supported", component.Spec.Client)
		}
	default:
		return nil, fmt.Errorf("no client support for %s", obj)
	}
}
