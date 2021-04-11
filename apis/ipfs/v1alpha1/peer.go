package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PeerSpec defines the desired state of Peer
type PeerSpec struct {
	// APIPort is api server port
	APIPort uint `json:"apiPort,omitempty"`
	// APIHost is api server host
	APIHost string `json:"apiHost,omitempty"`
	// GatewayPort is local gateway port
	GatewayPort uint `json:"gatewayPort,omitempty"`
	// Routing is the content routing mechanism
	Routing RoutingMechanism `json:"routing,omitempty"`
	// SwarmKeySecret is the k8s secret holding swarm key
	SwarmKeySecret string `json:"swarmKeySecret,omitempty"`
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// RoutingMechanism is the content routing mechanism
// +kubebuilder:validation:Enum=none;dht;dhtclient;dhtserver
type RoutingMechanism string

const (
	// NoneRouting is no routing mechanism
	NoneRouting RoutingMechanism = "none"
	// DHTRouting is automatic dht routing mechanism
	DHTRouting RoutingMechanism = "dht"
	// DHTClientRouting is the dht client routing mechanism
	DHTClientRouting RoutingMechanism = "dhtclient"
	// DHTServerRouting is the dht server routing mechanism
	DHTServerRouting RoutingMechanism = "dhtserver"
)

// PeerStatus defines the observed state of Peer
type PeerStatus struct {
	Client string `json:"client,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Peer is the Schema for the peers API
// +kubebuilder:printcolumn:name="Client",type=string,JSONPath=".status.client"
type Peer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PeerSpec   `json:"spec,omitempty"`
	Status PeerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PeerList contains a list of Peer
type PeerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Peer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Peer{}, &PeerList{})
}
