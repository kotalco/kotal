/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NetworkSpec defines the desired state of Network
type NetworkSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Join specifies the network to join
	// +optional
	Join string `json:"join,omitempty"`

	// Consensus is the consensus algorithm to be used by the network nodes to reach consensus
	// +optional
	Consensus ConsensusAlgorithm `json:"consensus,omitempty"`

	// Nodes is array of node specifications
	// +kubebuilder:validation:MinItems=1
	Nodes []Node `json:"nodes"`
}

//ConsensusAlgorithm is the algorithm nodes use to reach consensus
// +kubebuilder:validation:Enum=poa;pow;ibft2;quorum
type ConsensusAlgorithm string

const (
	//ProofOfAuthority is proof of authority consensus algorithm
	ProofOfAuthority ConsensusAlgorithm = "poa"
	//ProofOfWork is proof of work (satoshi consensus) consensus algorithm
	ProofOfWork ConsensusAlgorithm = "pow"
	//IBFT2 is Istanbul Byzantine Fault Tolerant consensus algorithm
	IBFT2 ConsensusAlgorithm = "ibft2"
	//Quorum is Quorum IBFT consensus algorithm
	Quorum ConsensusAlgorithm = "quorum"
)

//Node is the specification of the node
type Node struct {
	Name string `json:"name"`
}

// NetworkStatus defines the observed state of Network
type NetworkStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// Network is the Schema for the networks API
type Network struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetworkSpec   `json:"spec,omitempty"`
	Status NetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkList contains a list of Network
type NetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Network `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Network{}, &NetworkList{})
}
