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
	Join string `json:"join,omitempty"`

	// Consensus is the consensus algorithm to be used by the network nodes to reach consensus
	Consensus ConsensusAlgorithm `json:"consensus,omitempty"`

	// Nodes is array of node specifications
	// +kubebuilder:validation:MinItems=1
	Nodes []Node `json:"nodes"`
}

// ConsensusAlgorithm is the algorithm nodes use to reach consensus
// +kubebuilder:validation:Enum=poa;pow;ibft2;quorum
type ConsensusAlgorithm string

const (
	// ProofOfAuthority is proof of authority consensus algorithm
	ProofOfAuthority ConsensusAlgorithm = "poa"

	// ProofOfWork is proof of work (nakamoto consensus) consensus algorithm
	ProofOfWork ConsensusAlgorithm = "pow"

	// IBFT2 is Istanbul Byzantine Fault Tolerant consensus algorithm
	IBFT2 ConsensusAlgorithm = "ibft2"

	//Quorum is Quorum IBFT consensus algorithm
	Quorum ConsensusAlgorithm = "quorum"
)

// SynchronizationMode is the node synchronization mode
// +kubebuilder:validation:Enum=fast;full;archive
type SynchronizationMode string

// String returns the string value of the synchronization mode
func (sm SynchronizationMode) String() string {
	return string(sm)
}

const (
	//FastSynchronization is the full (archive) synchronization mode, alias for archive
	FastSynchronization SynchronizationMode = "fast"

	//ArchiveSynchronization is the archive synchronization mode, alias for full
	ArchiveSynchronization SynchronizationMode = "archive"

	//FullSynchronization is the fast (non-archival) synchronization mode
	FullSynchronization SynchronizationMode = "full"
)

// RPCService is name of the RPC service
// +kubebuilder:validation:Enum=admin;clique;debug;eea;eth;ibft;miner;net;perm;plugins;priv;txpool;web3
type RPCService string

// String returns string value of the RPC service
func (rs RPCService) String() string {
	return string(rs)
}

const (
	// AdminService is administration service
	AdminService RPCService = "admin"

	// CliqueService is clique (Proof of Authority consensus) service
	CliqueService RPCService = "clique"

	// DebugService is debugging service
	DebugService RPCService = "debug"

	// EEAService is EEA (Enterprise Ethereum Alliance) service
	EEAService RPCService = "eea"

	// ETHService is ethereum service
	ETHService RPCService = "eth"

	// IBFTService is IBFT consensus service
	IBFTService RPCService = "ibft"

	// MinerService is miner service
	MinerService RPCService = "miner"

	// NetworkService is network service
	NetworkService RPCService = "net"

	// PermissionService is permission service
	PermissionService RPCService = "perm"

	// PluginsService is plugins service
	PluginsService RPCService = "plugins"

	// PrivacyService is privacy service
	PrivacyService RPCService = "privacy"

	// TransactionPoolService is transaction pool service
	TransactionPoolService RPCService = "txpool"

	// Web3Service is web3 service
	Web3Service RPCService = "web3"
)

//Node is the specification of the node
type Node struct {
	// Name is the node name
	Name string `json:"name"`

	// SyncMode is the node synchronization mode
	SyncMode SynchronizationMode `json:"syncMode,omitempty"`

	// Miner is whether node is mining/validating blocks or no
	Miner bool `json:"miner,omitempty"`

	// MinerAccount is the account to which mining rewards are paid
	MinerAccount string `json:"minerAccount,omitempty"`

	// Hosts is a list of hostnames to to whitelist for RPC access
	Hosts []string `json:"hosts,omitempty"`

	// CORSDomains is the domains from which to accept cross origin requests
	CORSDomains []string `json:"corsDomains,omitempty"`

	// RPC is whether HTTP-RPC server is enabled or no
	RPC bool `json:"rpc,omitempty"`

	// RPCPort is HTTP-RPC server listening port
	RPCPort uint `json:"rpcPort,omitempty"`

	// RPCHost is HTTP-RPC server host address
	RPCHost string `json:"rpcHost,omitempty"`

	// RPCServices is a list of rpc services to enable
	RPCServices []RPCService `json:"rpcServices,omitempty"`
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
