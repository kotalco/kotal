package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BeaconNodeSpec defines the desired state of BeaconNode
type BeaconNodeSpec struct {
	// Network is the network to join
	Network string `json:"network"`
	// Client is the Ethereum 2.0 client to use
	Client Ethereum2Client `json:"client"`
	// Eth1Endpoints is Ethereum 1 endpoints
	Eth1Endpoints []string `json:"eth1Endpoints,omitempty"`

	// REST enables Beacon REST API
	REST bool `json:"rest,omitempty"`
	// RESTHost is Beacon REST API server host
	RESTHost string `json:"restHost,omitempty"`
	// RESTPort is Beacon REST API server port
	RESTPort uint `json:"restPort,omitempty"`

	// RPC enables RPC server
	RPC bool `json:"rpc,omitempty"`
	// RPCHost is host on which RPC server should listen
	RPCHost string `json:"rpcHost,omitempty"`
	// RPCPort is RPC server port
	RPCPort uint `json:"rpcPort,omitempty"`

	// GRPC enables GRPC gateway server
	GRPC bool `json:"grpc,omitempty"`
	// GRPCHost is GRPC gateway server host
	GRPCHost string `json:"grpcHost,omitempty"`
	// GRPCPort is GRPC gateway server port
	GRPCPort uint `json:"grpcPort,omitempty"`

	// P2PPort is p2p and discovery port
	P2PPort uint `json:"p2pPort,omitempty"`

	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// BeaconNodeStatus defines the observed state of BeaconNode
type BeaconNodeStatus struct {
}

// +kubebuilder:object:root=true

// BeaconNode is the Schema for the beaconnodes API
// +kubebuilder:printcolumn:name="Client",type=string,JSONPath=".spec.client"
// +kubebuilder:printcolumn:name="Network",type=string,JSONPath=".spec.network"
type BeaconNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BeaconNodeSpec   `json:"spec,omitempty"`
	Status BeaconNodeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BeaconNodeList contains a list of BeaconNodes
type BeaconNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BeaconNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BeaconNode{}, &BeaconNodeList{})
}
