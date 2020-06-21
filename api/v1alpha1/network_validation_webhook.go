package v1alpha1

import (
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-ethereum-kotal-io-v1alpha1-network,mutating=false,failurePolicy=fail,groups=ethereum.kotal.io,resources=networks,versions=v1alpha1,name=vnetwork.kb.io

var _ webhook.Validator = &Network{}

var (
	// ChainByID is public chains indexed by ID
	ChainByID = map[uint]string{
		1:    MainNetwork,
		3:    RopstenNetwork,
		4:    RinkebyNetwork,
		5:    GoerliNetwork,
		6:    KottiNetwork,
		61:   ClassicNetwork,
		63:   MordorNetwork,
		2018: DevNetwork,
	}
)

// ValidateMissingBootnodes validates that at least one bootnode in the network
func (r *Network) ValidateMissingBootnodes() *field.Error {
	// it's fine for a network of 1 node to have no bootnodes
	if len(r.Spec.Nodes) == 1 {
		return nil
	}

	if !r.Spec.Nodes[0].IsBootnode() {
		msg := "first node must be a bootnode if network has multiple nodes"
		return field.Invalid(field.NewPath("spec").Child("nodes").Index(0).Child("bootnode"), false, msg)
	}

	return nil
}

// ValidateNodeNameUniqeness validates that all node names are unique
func (r *Network) ValidateNodeNameUniqeness() field.ErrorList {

	var uniquenessErrors field.ErrorList
	names := map[string]int{}
	msg := "already used by spec.nodes[%d].name"
	nodesPath := field.NewPath("spec").Child("nodes")

	for i, node := range r.Spec.Nodes {
		if j, exists := names[node.Name]; exists {
			path := nodesPath.Index(i).Child("name")
			err := field.Invalid(path, node.Name, fmt.Sprintf(msg, j))
			uniquenessErrors = append(uniquenessErrors, err)
		} else {
			names[node.Name] = i
		}
	}
	return uniquenessErrors
}

// ValidateNode validates a single node
func (r *Network) ValidateNode(i int) field.ErrorList {
	node := r.Spec.Nodes[i]
	var nodeErrors field.ErrorList

	// validate nodekey is provided if node is bootnode
	if node.IsBootnode() && node.Nodekey == "" {
		err := field.Invalid(field.NewPath("spec").Child("nodes").Index(i).Child("nodekey"), node.Nodekey, "must provide nodekey if bootnode is true")
		nodeErrors = append(nodeErrors, err)
	}

	// validate coinbase is provided if node is miner
	if node.Miner && node.Coinbase == "" {
		err := field.Invalid(field.NewPath("spec").Child("nodes").Index(i).Child("coinbase"), "", "must provide coinbase if miner is true")
		nodeErrors = append(nodeErrors, err)
	}

	// validate coinbase can't be set if miner is not set explicitly as true
	if node.Coinbase != "" && node.Miner == false {
		err := field.Invalid(field.NewPath("spec").Child("nodes").Index(i).Child("miner"), false, "must set miner to true if coinbase is provided")
		nodeErrors = append(nodeErrors, err)
	}

	return nodeErrors
}

// ValidateNodes validate network nodes spec
func (r *Network) ValidateNodes() field.ErrorList {
	var allErrors field.ErrorList

	for i := range r.Spec.Nodes {
		allErrors = append(allErrors, r.ValidateNode(i)...)
	}

	allErrors = append(allErrors, r.ValidateNodeNameUniqeness()...)

	if err := r.ValidateMissingBootnodes(); err != nil {
		allErrors = append(allErrors, err)
	}

	return allErrors
}

// ValidateGenesis validates network genesis block spec
func (r *Network) ValidateGenesis() field.ErrorList {

	var allErrors field.ErrorList

	// join: can't specifiy genesis while joining existing network
	if r.Spec.Join != "" {
		err := field.Invalid(field.NewPath("spec").Child("join"), r.Spec.Join, "must be none if spec.genesis is specified")
		allErrors = append(allErrors, err)
	}

	// don't use existing network chain id
	if chain := ChainByID[r.Spec.Genesis.ChainID]; chain != "" {
		err := field.Invalid(field.NewPath("spec").Child("genesis").Child("chainId"), fmt.Sprintf("%d", r.Spec.Genesis.ChainID), fmt.Sprintf("can't use chain id of %s network to avoid tx replay", chain))
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
	return allErrors
}

// Validate is the shared validation between create and update
func (r *Network) Validate() field.ErrorList {
	var validateErrors field.ErrorList

	// consensus: can't specify consensus while joining existing network
	if r.Spec.Join != "" && r.Spec.Consensus != "" {
		err := field.Invalid(field.NewPath("spec").Child("consensus"), r.Spec.Consensus, "must be none while joining a network")
		validateErrors = append(validateErrors, err)
	}

	// genesis: must specify genesis if there's no network to join
	if r.Spec.Join == "" && r.Spec.Genesis == nil {
		err := field.Invalid(field.NewPath("spec").Child("genesis"), "", "must be specified if spec.join is none")
		validateErrors = append(validateErrors, err)
	}

	// validate non nil genesis
	if r.Spec.Genesis != nil {
		validateErrors = append(validateErrors, r.ValidateGenesis()...)
	}

	// validate nodes
	validateErrors = append(validateErrors, r.ValidateNodes()...)

	return validateErrors
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Network) ValidateCreate() error {
	var allErrors field.ErrorList

	networklog.Info("validate create", "name", r.Name)

	// shared validation rules with update
	allErrors = append(allErrors, r.Validate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Network) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList

	networklog.Info("validate update", "name", r.Name)

	// shared validation rules with create
	allErrors = append(allErrors, r.Validate()...)

	oldNetwork := old.(*Network)

	if oldNetwork.Spec.Join != r.Spec.Join {
		err := field.Invalid(field.NewPath("spec").Child("join"), r.Spec.Join, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if oldNetwork.Spec.Consensus != r.Spec.Consensus {
		err := field.Invalid(field.NewPath("spec").Child("consensus"), r.Spec.Consensus, "field is immutable")
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
