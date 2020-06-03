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
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var networklog = logf.Log.WithName("network-resource")

// SetupWebhookWithManager sets up the webook with a given controller manager
func (r *Network) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-ethereum-kotal-io-v1alpha1-network,mutating=true,failurePolicy=fail,groups=ethereum.kotal.io,resources=networks,verbs=create;update,versions=v1alpha1,name=mnetwork.kb.io

var _ webhook.Defaulter = &Network{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Network) Default() {
	networklog.Info("default", "name", r.Name)

	for i := range r.Spec.Nodes {
		r.DefaultNode(&r.Spec.Nodes[i])
	}

	if r.Spec.Genesis != nil {
		r.DefaultGenesis()
	}

}

// DefaultNode defaults a single node
func (r *Network) DefaultNode(node *Node) {

	defaultAPIs := []API{Web3API, ETHAPI, NetworkAPI}
	anyAddress := "0.0.0.0"
	allOrigins := []string{"*"}

	if node.P2PPort == 0 {
		node.P2PPort = 30303
	}

	if node.SyncMode == "" {
		node.SyncMode = FullSynchronization
	}

	if node.RPC || node.WS || node.GraphQL {
		if len(node.Hosts) == 0 {
			node.Hosts = allOrigins
		}

		if len(node.CORSDomains) == 0 {
			node.CORSDomains = allOrigins
		}
	}

	if node.RPC {
		if node.RPCHost == "" {
			node.RPCHost = anyAddress
		}

		if node.RPCPort == 0 {
			node.RPCPort = 8545
		}

		if len(node.RPCAPI) == 0 {
			node.RPCAPI = defaultAPIs
		}
	}

	if node.WS {
		if node.WSHost == "" {
			node.WSHost = anyAddress
		}

		if node.WSPort == 0 {
			node.WSPort = 8546
		}

		if len(node.WSAPI) == 0 {
			node.WSAPI = defaultAPIs
		}
	}

	if node.GraphQL {
		if node.GraphQLHost == "" {
			node.GraphQLHost = anyAddress
		}

		if node.GraphQLPort == 0 {
			node.GraphQLPort = 8547
		}
	}

}

// DefaultGenesis defaults genesis block parameters
func (r *Network) DefaultGenesis() {
	if r.Spec.Genesis.Coinbase == "" {
		r.Spec.Genesis.Coinbase = "0x0000000000000000000000000000000000000000"
	}

	if r.Spec.Genesis.Difficulty == "" {
		r.Spec.Genesis.Difficulty = "0x1"
	}

	if r.Spec.Genesis.Forks == nil {
		// all milestones will be activated at block 0
		r.Spec.Genesis.Forks = &Forks{
			EIP150Hash: "0x2086799aeebeae135c246c65021c82b4e15a2c451340993aacfd2751886514f0",
		}
	}

	if r.Spec.Genesis.MixHash == "" {
		r.Spec.Genesis.MixHash = "0x0000000000000000000000000000000000000000000000000000000000000000"
	}

	if r.Spec.Genesis.GasLimit == "" {
		r.Spec.Genesis.GasLimit = "0x47b760"
	}

	if r.Spec.Genesis.Nonce == "" {
		r.Spec.Genesis.Nonce = "0x0"
	}

	if r.Spec.Genesis.Timestamp == "" {
		r.Spec.Genesis.Timestamp = "0x0"
	}

	if r.Spec.Consensus == ProofOfAuthority {
		if r.Spec.Genesis.Clique.BlockPeriod == 0 {
			r.Spec.Genesis.Clique.BlockPeriod = 15
		}
		if r.Spec.Genesis.Clique.EpochLength == 0 {
			r.Spec.Genesis.Clique.EpochLength = 3000
		}
	}

	if r.Spec.Consensus == IstanbulBFT {
		if r.Spec.Genesis.IBFT2.BlockPeriod == 0 {
			r.Spec.Genesis.IBFT2.BlockPeriod = 15
		}
		if r.Spec.Genesis.IBFT2.EpochLength == 0 {
			r.Spec.Genesis.IBFT2.EpochLength = 3000
		}
		if r.Spec.Genesis.IBFT2.RequestTimeout == 0 {
			r.Spec.Genesis.IBFT2.RequestTimeout = 10
		}
		if r.Spec.Genesis.IBFT2.MessageQueueLimit == 0 {
			r.Spec.Genesis.IBFT2.MessageQueueLimit = 1000
		}
		if r.Spec.Genesis.IBFT2.DuplicateMesageLimit == 0 {
			r.Spec.Genesis.IBFT2.DuplicateMesageLimit = 100
		}
		if r.Spec.Genesis.IBFT2.FutureMessagesLimit == 0 {
			r.Spec.Genesis.IBFT2.FutureMessagesLimit = 1000
		}
		if r.Spec.Genesis.IBFT2.FutureMessagesMaxDistance == 0 {
			r.Spec.Genesis.IBFT2.FutureMessagesMaxDistance = 10
		}

	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-ethereum-kotal-io-v1alpha1-network,mutating=false,failurePolicy=fail,groups=ethereum.kotal.io,resources=networks,versions=v1alpha1,name=vnetwork.kb.io

var _ webhook.Validator = &Network{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Network) ValidateCreate() error {
	var allErrors field.ErrorList

	networklog.Info("validate create", "name", r.Name)

	// join: can't specifiy genesis while joining existing network
	if r.Spec.Join != "" && r.Spec.Genesis != nil {
		err := field.Invalid(field.NewPath("spec").Child("join"), r.Spec.Join, "must be None if spec.genesis is specified")
		allErrors = append(allErrors, err)
	}

	// consensus: can't specify consensus while joining existing network
	if r.Spec.Join != "" && r.Spec.Consensus != "" {
		err := field.Invalid(field.NewPath("spec").Child("consensus"), r.Spec.Consensus, "must be None while joining a network")
		allErrors = append(allErrors, err)
	}

	chainByID := map[uint]string{
		1:    "mainnet",
		3:    "ropsten",
		4:    "rinkeby",
		5:    "goerli",
		6:    "kotti",
		61:   "classic",
		63:   "mordor",
		2018: "dev",
	}

	// genesis
	if r.Spec.Genesis != nil {
		// don't use existing network chain id
		if chain := chainByID[r.Spec.Genesis.ChainID]; chain != "" {
			err := field.Invalid(field.NewPath("spec").Child("genesis").Child("chainId"), r.Spec.Genesis.ChainID, fmt.Sprintf("can't use chain id of %s network to avoid tx replay", chain))
			allErrors = append(allErrors, err)
		}

		// ethash must be nil of consensus is not Pow
		if r.Spec.Consensus != ProofOfWork && r.Spec.Genesis.Ethash != nil {
			err := field.Invalid(field.NewPath("spec").Child("consensus"), r.Spec.Consensus, fmt.Sprintf("must be %s if spec.genesis.ethash is specified", ProofOfWork))
			allErrors = append(allErrors, err)
		}

		// clique must be nil of consensus is not PoA
		if r.Spec.Consensus != ProofOfAuthority && r.Spec.Genesis.Clique != nil {
			err := field.Invalid(field.NewPath("spec").Child("consensus"), r.Spec.Consensus, fmt.Sprintf("must be %s if spec.genesis.clique is specified", ProofOfAuthority))
			allErrors = append(allErrors, err)
		}

		// ibft2 must be nil of consensus is not ibft2
		if r.Spec.Consensus != IstanbulBFT && r.Spec.Genesis.IBFT2 != nil {
			err := field.Invalid(field.NewPath("spec").Child("consensus"), r.Spec.Consensus, fmt.Sprintf("must be %s if spec.genesis.ibft2 is specified", IstanbulBFT))
			allErrors = append(allErrors, err)
		}
	}

	if len(r.Spec.Nodes) > 1 {
		// TODO: move to validateNetworkBootnodes
		missingBootnodes := true
		firstNode := r.Spec.Nodes[0]

		// TODO: move to validateNodeNamesUniqeness
		// unique node names and their index in spec.nodes[]
		nodeNames := map[string]int{}

		for i, node := range r.Spec.Nodes {
			bootnode := node.IsBootnode()

			// validate node name is not used more than once
			if ii, exists := nodeNames[node.Name]; exists {
				err := field.Invalid(field.NewPath("spec").Child("nodes").Index(i).Child("name"), node.Name, fmt.Sprintf("provided name already used by spec.nodes[%d].name", ii))
				allErrors = append(allErrors, err)
			} else {
				nodeNames[node.Name] = i
			}

			if bootnode {
				missingBootnodes = false
			}

			// validate nodekey is provided if node is bootnode
			if bootnode && node.Nodekey == "" {
				err := field.Invalid(field.NewPath("spec").Child("nodes").Index(i).Child("nodekey"), node.Nodekey, "must provide nodekey if bootnode is true")
				allErrors = append(allErrors, err)
			}

			// validate coinbase is provided if node is miner
			if node.Miner && node.Coinbase == "" {
				err := field.Invalid(field.NewPath("spec").Child("nodes").Index(i).Child("coinbase"), "", "must provide coinbase if miner is true")
				allErrors = append(allErrors, err)
			}

		}

		// first node must be a bootnode or it will be orphaned
		if !firstNode.IsBootnode() {
			err := field.Invalid(field.NewPath("spec").Child("nodes").Index(0).Child("bootnode"), false, "must be true or it will be orphaned")
			allErrors = append(allErrors, err)
		}

		//at least one node should be a bootnode
		if missingBootnodes {
			err := field.Invalid(field.NewPath("spec").Child("nodes"), nil, "at least one node must be a bootnode")
			allErrors = append(allErrors, err)
		}

	}

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)

}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Network) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList

	networklog.Info("validate update", "name", r.Name)

	oldNetwork := old.(*Network)

	if oldNetwork.Spec.Join != r.Spec.Join {
		err := field.Invalid(field.NewPath("spec").Child("join"), r.Spec.Join, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if oldNetwork.Spec.Consensus != r.Spec.Consensus {
		err := field.Invalid(field.NewPath("spec").Child("consensus"), r.Spec.Join, "field is immutable")
		allErrors = append(allErrors, err)
	}

	// TODO: move to validate genesis
	if !reflect.DeepEqual(r.Spec.Genesis, oldNetwork.Spec.Genesis) {
		err := field.Invalid(field.NewPath("spec").Child("genesis"), "", "field is immutable")
		allErrors = append(allErrors, err)
	}

	// maximum allowed nodes with different name
	var maxDiff int
	// all old nodes names
	oldNodesNames := map[string]bool{}
	// nodes count in the old network spec
	oldNodesCount := len(oldNetwork.Spec.Nodes)
	// nodes count in the new network spec
	newNodesCount := len(r.Spec.Nodes)
	// nodes with different names than the old spec
	differentNodes := map[string]int{}

	if newNodesCount > oldNodesCount {
		maxDiff = newNodesCount - oldNodesCount
	}

	for _, node := range oldNetwork.Spec.Nodes {
		oldNodesNames[node.Name] = true
	}

	for i, node := range r.Spec.Nodes {
		if exists := oldNodesNames[node.Name]; !exists {
			differentNodes[node.Name] = i
		}
	}

	if len(differentNodes) > maxDiff {
		for nodeName, i := range differentNodes {
			err := field.Invalid(field.NewPath("spec").Child("nodes").Index(i).Child("name"), nodeName, "field is immutable")
			allErrors = append(allErrors, err)
		}
	}

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)

}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Network) ValidateDelete() error {
	networklog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
