package v1alpha1

import (
	"fmt"

	"github.com/kotalco/kotal/apis/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SwarmSpec defines the desired state of Swarm
type SwarmSpec struct {
	// Nodes is swarm nodes
	// +kubebuilder:validation:MinItems=1
	Nodes []Node `json:"nodes"`
}

// Node is ipfs node
type Node struct {
	// Name is node name
	Name string `json:"name"`
	// ID is node peer ID
	ID string `json:"id"`
	// PrivateKey is node private key
	PrivateKey string `json:"privateKey"`
	// Profiles is a list of profiles to apply
	Profiles []Profile `json:"profiles,omitempty"`
	// Resources is node compute and storage resources
	shared.Resources `json:"resources,omitempty"`
}

// SwarmAddress returns node swarm address
func (n *Node) SwarmAddress(ip string) string {
	// TODO: replace hardcoded 4001 port with node swarm port
	return fmt.Sprintf("/ip4/%s/tcp/4001/p2p/%s", ip, n.ID)
}

// StatefulSetName returns name to be used by node stateful
func (n *Node) StatefulSetName(swarm string) string {
	return fmt.Sprintf("%s-%s", swarm, n.Name)
}

// PVCName returns name to be used by node pvc
func (n *Node) PVCName(swarm string) string {
	return n.StatefulSetName(swarm) // same as stateful name
}

// ConfigName returns name to be used by node config map
func (n *Node) ConfigName(swarm string) string {
	return n.StatefulSetName(swarm) // same as stateful name
}

// ServiceName returns name to be used by node service
func (n *Node) ServiceName(swarm string) string {
	return n.StatefulSetName(swarm) // same as stateful name
}

// Labels to be used by node resources
func (n *Node) Labels(swarm string) map[string]string {
	return map[string]string{
		"name":     "node",
		"instance": n.Name,
		"swarm":    swarm,
	}
}

// Profile is ipfs configuration
// +kubebuilder:validation:Enum=server;randomports;default-datastore;local-discovery;test;default-networking;flatfs;badgerds;lowpower
type Profile string

// SwarmStatus defines the observed state of Swarm
type SwarmStatus struct {
	// NodesCount is number of nodes in this swarm
	NodesCount int `json:"nodesCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Swarm is the Schema for the swarms API
// +kubebuilder:printcolumn:name="Nodes",type=integer,JSONPath=".status.nodesCount"
type Swarm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SwarmSpec   `json:"spec,omitempty"`
	Status SwarmStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SwarmList contains a list of Swarm
type SwarmList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Swarm `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Swarm{}, &SwarmList{})
}
