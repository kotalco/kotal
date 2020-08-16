package v1alpha1

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/kotalco/kotal/helpers"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
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
	nodePath := field.NewPath("spec").Child("nodes").Index(i)
	var nodeErrors field.ErrorList

	// validate nodekey is provided if node is bootnode
	if node.IsBootnode() && node.Nodekey == "" {
		err := field.Invalid(nodePath.Child("nodekey"), node.Nodekey, "must provide nodekey if bootnode is true")
		nodeErrors = append(nodeErrors, err)
	}

	cpu := resource.MustParse(node.Resources.CPU)
	cpuLimit := resource.MustParse(node.Resources.CPULimit)

	// validate cpuLimit can't be less than cpu request
	if cpuLimit.Cmp(cpu) == -1 {
		err := field.Invalid(nodePath.Child("resources").Child("cpuLimit"), node.Resources.CPULimit, fmt.Sprintf("must be greater than or equal to cpu %s", string(node.Resources.CPU)))
		nodeErrors = append(nodeErrors, err)
	}

	memory := resource.MustParse(node.Resources.Memory)
	memoryLimit := resource.MustParse(node.Resources.MemoryLimit)

	// validate memoryLimit can't be less than memory request
	if memoryLimit.Cmp(memory) == -1 {
		err := field.Invalid(nodePath.Child("resources").Child("memoryLimit"), node.Resources.MemoryLimit, fmt.Sprintf("must be greater than or equal to memory %s", string(node.Resources.Memory)))
		nodeErrors = append(nodeErrors, err)
	}

	// validate coinbase is provided if node is miner
	if node.Miner && node.Coinbase == "" {
		err := field.Invalid(nodePath.Child("coinbase"), "", "must provide coinbase if miner is true")
		nodeErrors = append(nodeErrors, err)
	}

	// validate coinbase can't be set if miner is not set explicitly as true
	if node.Coinbase != "" && node.Miner == false {
		err := field.Invalid(nodePath.Child("miner"), false, "must set miner to true if coinbase is provided")
		nodeErrors = append(nodeErrors, err)
	}

	// validate only geth client can import accounts
	if node.Client != GethClient && node.Import != nil {
		err := field.Invalid(nodePath.Child("client"), node.Client, "must be geth if import is provided")
		nodeErrors = append(nodeErrors, err)
	}

	// validate only geth client supports light sync mode
	if node.Client != GethClient && node.SyncMode == LightSynchronization {
		err := field.Invalid(nodePath.Child("client"), node.Client, "must be geth if syncMode is light")
		nodeErrors = append(nodeErrors, err)
	}

	// Validate geth node
	if node.Client == GethClient {
		nodeErrors = append(nodeErrors, r.ValidateGethNode(&node, i)...)
	}

	return nodeErrors
}

// ValidateGethNode validates a node with client geth
func (r *Network) ValidateGethNode(node *Node, i int) field.ErrorList {
	var gethErrors field.ErrorList
	nodePath := field.NewPath("spec").Child("nodes").Index(i)

	// validate geth supports only pow and poa
	if r.Spec.Join == "" && r.Spec.Consensus != ProofOfWork && r.Spec.Consensus != ProofOfAuthority {
		err := field.Invalid(nodePath.Child("client"), node.Client, fmt.Sprintf("client doesn't support %s consensus", r.Spec.Consensus))
		gethErrors = append(gethErrors, err)
	}

	// validate geth doesn't support fixed difficulty ethash networks
	if r.Spec.Join == "" && r.Spec.Consensus == ProofOfWork && r.Spec.Genesis.Ethash.FixedDifficulty != nil {
		err := field.Invalid(nodePath.Child("client"), node.Client, "client doesn't support fixed difficulty pow networks")
		gethErrors = append(gethErrors, err)
	}

	// validate account must be imported if coinbase is provided
	if node.Coinbase != "" && node.Import == nil {
		err := field.Invalid(nodePath.Child("import"), "", "must import coinbase account")
		gethErrors = append(gethErrors, err)
	}

	// validate imported account private key is valid and coinbase account is derived from it
	if node.Coinbase != "" && node.Import != nil {
		privateKey := node.Import.PrivateKey[2:]
		address, err := helpers.DeriveAddress(string(privateKey))
		if err != nil {
			err := field.Invalid(nodePath.Child("import").Child("privatekey"), "<private key>", "invalid private key")
			gethErrors = append(gethErrors, err)
		}

		if strings.ToLower(string(node.Coinbase)) != strings.ToLower(address) {
			err := field.Invalid(nodePath.Child("import").Child("privatekey"), "<private key>", "private key doesn't correspond to the coinbase address")
			gethErrors = append(gethErrors, err)
		}
	}

	// validate rpc can't be enabled for node with imported account
	if node.Import != nil && node.RPC {
		err := field.Invalid(nodePath.Child("rpc"), node.RPC, "must be false if import is provided")
		gethErrors = append(gethErrors, err)
	}

	// validate ws can't be enabled for node with imported account
	if node.Import != nil && node.WS {
		err := field.Invalid(nodePath.Child("ws"), node.WS, "must be false if import is provided")
		gethErrors = append(gethErrors, err)
	}

	// validate graphql can't be enabled for node with imported account
	if node.Import != nil && node.GraphQL {
		err := field.Invalid(nodePath.Child("graphql"), node.GraphQL, "must be false if import is provided")
		gethErrors = append(gethErrors, err)
	}

	return gethErrors
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

	// validate forks order
	allErrors = append(allErrors, r.ValidateForksOrder()...)
	return allErrors
}

// ValidateForksOrder validates that forks are in correct order
func (r *Network) ValidateForksOrder() field.ErrorList {
	var orderErrors field.ErrorList
	forks := r.Spec.Genesis.Forks

	forkNames := []string{
		"homestead",
		"eip150",
		"eip155",
		"eip155",
		"byzantium",
		"constantinople",
		"petersburg",
		"istanbul",
		"muirglacier",
	}

	milestones := []uint{
		forks.Homestead,
		forks.EIP150,
		forks.EIP155,
		forks.EIP155,
		forks.Byzantium,
		forks.Constantinople,
		forks.Petersburg,
		forks.Istanbul,
		forks.MuirGlacier,
	}

	for i := 1; i < len(milestones); i++ {
		if milestones[i] < milestones[i-1] {
			path := field.NewPath("spec").Child("genesis").Child("forks").Child(forkNames[i])
			msg := fmt.Sprintf("Fork %s can't be activated (at block %d) before fork %s (at block %d)", forkNames[i], milestones[i], forkNames[i-1], milestones[i-1])
			orderErrors = append(orderErrors, field.Invalid(path, fmt.Sprintf("%d", milestones[i]), msg))
		}
	}

	return orderErrors

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

	// id: must be provided if join is none
	if r.Spec.Join == "" && r.Spec.ID == 0 {
		err := field.Invalid(field.NewPath("spec").Child("id"), "", "must be specified if spec.join is none")
		validateErrors = append(validateErrors, err)
	}

	// id: must be none if join is provided
	if r.Spec.Join != "" && r.Spec.ID != 0 {
		err := field.Invalid(field.NewPath("spec").Child("id"), fmt.Sprintf("%d", r.Spec.ID), "must be none if spec.join is provided")
		validateErrors = append(validateErrors, err)
	}

	// consensus: must be provided if genesis is provided
	if r.Spec.Genesis != nil && r.Spec.Consensus == "" {
		err := field.Invalid(field.NewPath("spec").Child("consensus"), "", "must be specified if spec.genesis is provided")
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

	if oldNetwork.Spec.ID != r.Spec.ID {
		err := field.Invalid(field.NewPath("spec").Child("id"), fmt.Sprintf("%d", r.Spec.ID), "field is immutable")
		allErrors = append(allErrors, err)
	}

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
