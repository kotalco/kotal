package v1alpha1

import (
	"fmt"
	"strings"

	"github.com/kotalco/kotal/helpers"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-ethereum-kotal-io-v1alpha1-node,mutating=false,failurePolicy=fail,groups=ethereum.kotal.io,resources=nodes,versions=v1alpha1,name=validate-ethereum-v1alpha1-node.kb.io

var _ webhook.Validator = &Node{}

// Validate validates a node with a given path
func (n *Node) Validate(path *field.Path, validateNetworkConfig bool) field.ErrorList {
	var nodeErrors field.ErrorList

	privateNetwork := n.Spec.Genesis != nil

	if validateNetworkConfig {
		nodeErrors = append(nodeErrors, n.Spec.NetworkConfig.Validate()...)
	}

	// validate nodekey is provided if node is bootnode
	if n.Spec.Bootnode == true && n.Spec.Nodekey == "" {
		err := field.Invalid(path.Child("nodekey"), n.Spec.Nodekey, "must provide nodekey if bootnode is true")
		nodeErrors = append(nodeErrors, err)
	}

	// validate fatal and trace logging not supported by geth
	if n.Spec.Client == GethClient && (n.Spec.Logging == FatalLogs || n.Spec.Logging == TraceLogs) {
		err := field.Invalid(path.Child("logging"), n.Spec.Logging, fmt.Sprintf("not supported by client %s", n.Spec.Client))
		nodeErrors = append(nodeErrors, err)
	}

	// validate off, fatal and all logs not supported by parity
	if n.Spec.Client == ParityClient && (n.Spec.Logging == NoLogs || n.Spec.Logging == FatalLogs || n.Spec.Logging == AllLogs) {
		err := field.Invalid(path.Child("logging"), n.Spec.Logging, fmt.Sprintf("not supported by client %s", n.Spec.Client))
		nodeErrors = append(nodeErrors, err)
	}

	// if cpu and cpulimit are string equal, no need to compare quantity
	if n.Spec.Resources.CPU != n.Spec.Resources.CPULimit {
		cpu := resource.MustParse(n.Spec.Resources.CPU)
		cpuLimit := resource.MustParse(n.Spec.Resources.CPULimit)

		// validate cpuLimit can't be less than cpu request
		if cpuLimit.Cmp(cpu) == -1 {
			err := field.Invalid(path.Child("resources").Child("cpuLimit"), n.Spec.Resources.CPULimit, fmt.Sprintf("must be greater than or equal to cpu %s", string(n.Spec.Resources.CPU)))
			nodeErrors = append(nodeErrors, err)
		}
	}

	// if memory and memoryLimit are string equal, no need to compare quantity
	if n.Spec.Resources.Memory != n.Spec.Resources.MemoryLimit {
		memory := resource.MustParse(n.Spec.Resources.Memory)
		memoryLimit := resource.MustParse(n.Spec.Resources.MemoryLimit)

		// validate memoryLimit can't be less than memory request
		if memoryLimit.Cmp(memory) == -1 {
			err := field.Invalid(path.Child("resources").Child("memoryLimit"), n.Spec.Resources.MemoryLimit, fmt.Sprintf("must be greater than or equal to memory %s", string(n.Spec.Resources.Memory)))
			nodeErrors = append(nodeErrors, err)
		}
	}

	// validate coinbase is provided if node is miner
	if n.Spec.Miner && n.Spec.Coinbase == "" {
		err := field.Invalid(path.Child("coinbase"), "", "must provide coinbase if miner is true")
		nodeErrors = append(nodeErrors, err)
	}

	// validate coinbase can't be set if miner is not set explicitly as true
	if n.Spec.Coinbase != "" && n.Spec.Miner == false {
		err := field.Invalid(path.Child("miner"), false, "must set miner to true if coinbase is provided")
		nodeErrors = append(nodeErrors, err)
	}

	// validate only geth client can import accounts
	if n.Spec.Client == BesuClient && n.Spec.Import != nil {
		err := field.Invalid(path.Child("client"), n.Spec.Client, "must be geth or parity if import is provided")
		nodeErrors = append(nodeErrors, err)
	}

	// validate rpc must be enabled if grapql is enabled and geth is used
	if n.Spec.Client == GethClient && n.Spec.GraphQL && !n.Spec.RPC {
		err := field.Invalid(path.Child("rpc"), n.Spec.RPC, "must enable rpc if client is geth and graphql is enabled")
		nodeErrors = append(nodeErrors, err)
	}

	// validate parity doesn't support graphql
	if n.Spec.Client == ParityClient && n.Spec.GraphQL {
		err := field.Invalid(path.Child("client"), n.Spec.Client, "client doesn't support graphQL")
		nodeErrors = append(nodeErrors, err)
	}

	// validate only geth client supports light sync mode
	if n.Spec.Client != GethClient && n.Spec.SyncMode == LightSynchronization {
		err := field.Invalid(path.Child("client"), n.Spec.Client, "must be geth if syncMode is light")
		nodeErrors = append(nodeErrors, err)
	}

	// validate geth supports only pow and poa
	if privateNetwork && n.Spec.Consensus == IstanbulBFT && n.Spec.Client != BesuClient {
		err := field.Invalid(path.Child("client"), n.Spec.Client, fmt.Sprintf("client doesn't support %s consensus", n.Spec.Consensus))
		nodeErrors = append(nodeErrors, err)
	}

	// validate besu only support fixed difficulty ethash networks
	if privateNetwork && n.Spec.Consensus == ProofOfWork && n.Spec.Genesis.Ethash.FixedDifficulty != nil && n.Spec.Client != BesuClient {
		err := field.Invalid(path.Child("client"), n.Spec.Client, "client doesn't support fixed difficulty pow networks")
		nodeErrors = append(nodeErrors, err)
	}

	// validate account must be imported if coinbase is provided
	if n.Spec.Client != BesuClient && n.Spec.Coinbase != "" && n.Spec.Import == nil {
		err := field.Invalid(path.Child("import"), "", "must import coinbase account")
		nodeErrors = append(nodeErrors, err)
	}

	// validate parity doesn't support PoW mining
	if n.Spec.Client == ParityClient && n.Spec.Consensus == ProofOfWork && n.Spec.Miner {
		err := field.Invalid(path.Child("client"), n.Spec.Client, "client doesn't support mining")
		nodeErrors = append(nodeErrors, err)
	}

	// validate imported account private key is valid and coinbase account is derived from it
	// TODO: cache private address -> address results
	if n.Spec.Client != BesuClient && n.Spec.Coinbase != "" && n.Spec.Import != nil {
		privateKey := n.Spec.Import.PrivateKey[2:]
		address, err := helpers.DeriveAddress(string(privateKey))
		if err != nil {
			err := field.Invalid(path.Child("import").Child("privatekey"), "<private key>", "invalid private key")
			nodeErrors = append(nodeErrors, err)
		}

		if strings.ToLower(string(n.Spec.Coinbase)) != strings.ToLower(address) {
			err := field.Invalid(path.Child("import").Child("privatekey"), "<private key>", "private key doesn't correspond to the coinbase address")
			nodeErrors = append(nodeErrors, err)
		}
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

	allErrors = append(allErrors, n.Validate(field.NewPath("spec"), true)...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, n.Name, allErrors)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (n *Node) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList

	nodelog.Info("validate update", "name", n.Name)

	allErrors = append(allErrors, n.Validate(field.NewPath("spec"), true)...)

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
