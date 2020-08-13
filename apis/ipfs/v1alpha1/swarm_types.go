package v1alpha1

import (
	"fmt"
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
	Resources *NodeResources `json:"resources,omitempty"`
}

// SwarmAddress returns node swarm address
func (n *Node) SwarmAddress(ip string) string {
	// TODO: replace hardcoded 4001 port with node swarm port
	return fmt.Sprintf("/ip4/%s/tcp/4001/p2p/%s", ip, n.ID)
}

// DeploymentName returns name to be used by node deployment
func (n *Node) DeploymentName(swarm string) string {
	return fmt.Sprintf("%s-%s", swarm, n.Name)
}

// PVCName returns name to be used by node pvc
func (n *Node) PVCName(swarm string) string {
	return n.DeploymentName(swarm) // same as deployment name
}

// ConfigName returns name to be used by node config map
func (n *Node) ConfigName(swarm string) string {
	return n.DeploymentName(swarm) // same as deployment name
}

// ServiceName returns name to be used by node service
func (n *Node) ServiceName(swarm string) string {
	return n.DeploymentName(swarm) // same as deployment name
}

// Labels to be used by node resources
func (n *Node) Labels(swarm string) map[string]string {
	return map[string]string{
		"name":     "node",
		"instance": n.Name,
		"swarm":    swarm,
	}
}

// NodeResources is node compute and storage resources
type NodeResources struct {
	// CPU is cpu cores the node requires
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*m?$"
	CPU string `json:"cpu,omitempty"`
	// CPULimit is cpu cores the node is limited to
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*m?$"
	CPULimit string `json:"cpuLimit,omitempty"`
	// Memory is memmory requirements
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*[KMGTPE]i$"
	Memory string `json:"memory,omitempty"`
	// MemoryLimit is cpu cores the node is limited to
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*[KMGTPE]i$"
	MemoryLimit string `json:"memoryLimit,omitempty"`
	// Storage is disk space storage requirements
	// +kubebuilder:validation:Pattern="^[1-9][0-9]*[KMGTPE]i$"
	Storage string `json:"storage,omitempty"`
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
