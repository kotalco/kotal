package v1alpha1

import (
	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PeerSpec defines the desired state of Peer
type PeerSpec struct {
	// Image is ipfs peer client image
	Image string `json:"image,omitempty"`
	// InitProfiles is the intial profiles to apply during
	// +listType=set
	InitProfiles []Profile `json:"initProfiles,omitempty"`
	// Profiles is the configuration profiles to apply after peer initialization
	// +listType=set
	Profiles []Profile `json:"profiles,omitempty"`
	// API enables API server
	API bool `json:"api,omitempty"`
	// APIPort is api server port
	APIPort uint `json:"apiPort,omitempty"`
	// Gateway enables IPFS gateway server
	Gateway bool `json:"gateway,omitempty"`
	// GatewayPort is local gateway port
	GatewayPort uint `json:"gatewayPort,omitempty"`
	// Routing is the content routing mechanism
	Routing RoutingMechanism `json:"routing,omitempty"`
	// SwarmKeySecretName is the k8s secret holding swarm key
	SwarmKeySecretName string `json:"swarmKeySecretName,omitempty"`
	// Logging is logging verboisty level
	// +kubebuilder:validation:Enum=error;warn;info;debug;notice
	Logging shared.VerbosityLevel `json:"logging,omitempty"`
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// Profile is ipfs configuration
// +kubebuilder:validation:Enum=server;randomports;default-datastore;local-discovery;test;default-networking;flatfs;badgerds;lowpower
type Profile string

const (
	// ServerProfile is the server profile
	ServerProfile Profile = "server"
	// RandomPortsProfile is the random ports profile
	RandomPortsProfile Profile = "randomports"
	// DefaultDatastoreProfile is the default data store profile
	DefaultDatastoreProfile Profile = "default-datastore"
	// LocalDiscoveryProfile is the local discovery profile
	LocalDiscoveryProfile Profile = "local-discovery"
	// TestProfile is the test profile
	TestProfile Profile = "test"
	// DefaultNetworkingProfile is the default networking profile
	DefaultNetworkingProfile Profile = "default-networking"
	// FlatFSProfile is the flat file system profile
	FlatFSProfile Profile = "flatfs"
	// BadgerDSProfile is badger data store profile
	BadgerDSProfile Profile = "badgerds"
	// LowPowerProfile is the low power profile
	LowPowerProfile Profile = "lowpower"
)

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
