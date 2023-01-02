package v1alpha1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-ethereum-kotal-io-v1alpha1-node,mutating=false,failurePolicy=fail,groups=ethereum.kotal.io,resources=nodes,versions=v1alpha1,name=validate-ethereum-v1alpha1-node.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &Node{}

// validate validates a node with a given path
func (n *Node) validate() field.ErrorList {
	var nodeErrors field.ErrorList

	privateNetwork := n.Spec.Genesis != nil

	path := field.NewPath("spec")

	// network: can't specifiy genesis while joining existing network
	if n.Spec.Network != "" && n.Spec.Genesis != nil {
		err := field.Invalid(field.NewPath("spec").Child("network"), n.Spec.Network, "must be none if spec.genesis is specified")
		nodeErrors = append(nodeErrors, err)
	}

	// genesis: must specify genesis if there's no network to join
	if n.Spec.Network == "" && n.Spec.Genesis == nil {
		err := field.Invalid(field.NewPath("spec").Child("genesis"), "", "must be specified if spec.network is none")
		nodeErrors = append(nodeErrors, err)
	}

	if !n.Spec.Client.SupportsVerbosityLevel(n.Spec.Logging) {
		err := field.Invalid(path.Child("logging"), n.Spec.Logging, fmt.Sprintf("not supported by client %s", n.Spec.Client))
		nodeErrors = append(nodeErrors, err)
	}

	// validate coinbase is provided if node is miner
	if n.Spec.Miner && n.Spec.Coinbase == "" {
		err := field.Invalid(path.Child("coinbase"), "", "must provide coinbase if miner is true")
		nodeErrors = append(nodeErrors, err)
	}

	// validate jwtSecretName is provided if engine is enabled
	if n.Spec.Engine && n.Spec.JWTSecretName == "" {
		err := field.Invalid(path.Child("jwtSecretName"), "", "must provide jwtSecretName if engine is true")
		nodeErrors = append(nodeErrors, err)
	}

	// validate coinbase can't be set if miner is not set explicitly as true
	if n.Spec.Coinbase != "" && !n.Spec.Miner {
		err := field.Invalid(path.Child("miner"), false, "must set miner to true if coinbase is provided")
		nodeErrors = append(nodeErrors, err)
	}

	// validate that besu doesn't support importing ethereum accounts
	// Netermind, go-ethereum, and OpenEthereum support importing accounts
	if n.Spec.Client == BesuClient && n.Spec.Import != nil {
		err := field.Invalid(path.Child("client"), n.Spec.Client, "client doesn't support importing accounts")
		nodeErrors = append(nodeErrors, err)
	}

	// validate rpc must be enabled if grapql is enabled and geth is used
	if n.Spec.Client == GethClient && n.Spec.GraphQL && !n.Spec.RPC {
		err := field.Invalid(path.Child("rpc"), n.Spec.RPC, "must enable rpc if client is geth and graphql is enabled")
		nodeErrors = append(nodeErrors, err)
	}

	// validate nethermind doesn't support GraphQL
	if n.Spec.GraphQL && n.Spec.Client == NethermindClient {
		err := field.Invalid(path.Child("client"), n.Spec.Client, "client doesn't support GraphQL")
		nodeErrors = append(nodeErrors, err)
	}

	// validate nethermind doesn't support hosts whitelisting
	if len(n.Spec.Hosts) > 0 && n.Spec.Client == NethermindClient {
		err := field.Invalid(path.Child("client"), n.Spec.Client, "client doesn't support hosts whitelisting")
		nodeErrors = append(nodeErrors, err)
	}

	// validate nethermind doesn't support CORS domains
	if len(n.Spec.CORSDomains) > 0 && n.Spec.Client == NethermindClient {
		err := field.Invalid(path.Child("client"), n.Spec.Client, "client doesn't support CORS domains")
		nodeErrors = append(nodeErrors, err)
	}

	// validate only geth client supports light sync mode
	if (n.Spec.SyncMode == LightSynchronization || n.Spec.SyncMode == SnapSynchronization) && n.Spec.Client != GethClient {
		err := field.Invalid(path.Child("syncMode"), n.Spec.SyncMode, fmt.Sprintf("not supported by client %s", n.Spec.Client))
		nodeErrors = append(nodeErrors, err)
	}

	// validate geth supports only pow and poa
	if privateNetwork && n.Spec.Genesis.IBFT2 != nil && n.Spec.Client != BesuClient {
		err := field.Invalid(path.Child("client"), n.Spec.Client, "client doesn't support ibft2 consensus")
		nodeErrors = append(nodeErrors, err)
	}

	// validate besu only support fixed difficulty ethash networks
	if privateNetwork && n.Spec.Genesis.Ethash != nil && n.Spec.Genesis.Ethash.FixedDifficulty != nil && n.Spec.Client != BesuClient {
		err := field.Invalid(path.Child("client"), n.Spec.Client, "client doesn't support fixed difficulty pow networks")
		nodeErrors = append(nodeErrors, err)
	}

	// validate account must be imported if coinbase is provided
	if n.Spec.Client != BesuClient && n.Spec.Coinbase != "" && n.Spec.Import == nil {
		err := field.Invalid(path.Child("import"), "", "must import coinbase account")
		nodeErrors = append(nodeErrors, err)
	}

	// validate rpc can't be enabled for node with imported account
	if n.Spec.Client != BesuClient && n.Spec.Import != nil && n.Spec.RPC {
		err := field.Invalid(path.Child("rpc"), n.Spec.RPC, "must be false if import is provided")
		nodeErrors = append(nodeErrors, err)
	}

	// validate ws can't be enabled for node with imported account
	if n.Spec.Client != BesuClient && n.Spec.Import != nil && n.Spec.WS {
		err := field.Invalid(path.Child("ws"), n.Spec.WS, "must be false if import is provided")
		nodeErrors = append(nodeErrors, err)
	}

	// validate graphql can't be enabled for node with imported account
	if n.Spec.Client != BesuClient && n.Spec.Import != nil && n.Spec.GraphQL {
		err := field.Invalid(path.Child("graphql"), n.Spec.GraphQL, "must be false if import is provided")
		nodeErrors = append(nodeErrors, err)
	}

	return nodeErrors
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (n *Node) ValidateCreate() error {
	var allErrors field.ErrorList

	nodelog.Info("validate create", "name", n.Name)

	allErrors = append(allErrors, n.validate()...)
	allErrors = append(allErrors, n.Spec.Resources.ValidateCreate()...)

	// validate genesis block
	if n.Spec.Genesis != nil {
		allErrors = append(allErrors, n.Spec.Genesis.ValidateCreate()...)
	}

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, n.Name, allErrors)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (n *Node) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList
	oldNode := old.(*Node)

	nodelog.Info("validate update", "name", n.Name)

	if n.Spec.Client != oldNode.Spec.Client {
		err := field.Invalid(field.NewPath("spec").Child("client"), n.Spec.Client, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if oldNode.Spec.Network != n.Spec.Network {
		err := field.Invalid(field.NewPath("spec").Child("network"), n.Spec.Network, "field is immutable")
		allErrors = append(allErrors, err)
	}

	// validate genesis block
	if oldNode.Spec.Genesis != nil {
		allErrors = append(allErrors, n.Spec.Genesis.ValidateUpdate(oldNode.Spec.Genesis)...)
	}

	allErrors = append(allErrors, n.validate()...)
	allErrors = append(allErrors, n.Spec.Resources.ValidateUpdate(&oldNode.Spec.Resources)...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, n.Name, allErrors)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (n *Node) ValidateDelete() error {
	nodelog.Info("validate delete", "name", n.Name)

	return nil
}
