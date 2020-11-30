package v1alpha1

import (
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
func (n *Network) ValidateMissingBootnodes() *field.Error {
	// it's fine for a network of 1 node to have no bootnodes
	if len(n.Spec.Nodes) == 1 {
		return nil
	}

	if n.Spec.Nodes[0].Bootnode != true {
		msg := "first node must be a bootnode if network has multiple nodes"
		return field.Invalid(field.NewPath("spec").Child("nodes").Index(0).Child("bootnode"), false, msg)
	}

	return nil
}

// ValidateNodes validate network nodes spec
func (n *Network) ValidateNodes() field.ErrorList {
	var allErrors field.ErrorList

	for i := range n.Spec.Nodes {
		path := field.NewPath("spec").Child("nodes").Index(i)
		node := Node{
			Spec: n.Spec.Nodes[i].NodeSpec,
		}
		// No need to pass network and availability config
		// it has already been passed during network defaulting phase
		// no need to validate network config
		allErrors = append(allErrors, node.Validate(path, false)...)
	}

	if err := n.ValidateMissingBootnodes(); err != nil {
		allErrors = append(allErrors, err)
	}

	return allErrors
}

// Validate is the shared validation between create and update
func (n *Network) Validate() field.ErrorList {
	var validateErrors field.ErrorList

	// validate network config (id, genesis, consensus and join)
	validateErrors = append(validateErrors, n.Spec.NetworkConfig.Validate()...)

	// validate nodes
	validateErrors = append(validateErrors, n.ValidateNodes()...)

	return validateErrors
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (n *Network) ValidateCreate() error {
	var allErrors field.ErrorList

	networklog.Info("validate create", "name", n.Name)

	// shared validation rules with update
	allErrors = append(allErrors, n.Validate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, n.Name, allErrors)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (n *Network) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList
	oldNetwork := old.(*Network)

	networklog.Info("validate update", "name", n.Name)

	// shared validation rules with create
	allErrors = append(allErrors, n.Validate()...)

	// shared validation rules with create
	allErrors = append(allErrors, n.Spec.NetworkConfig.ValidateUpdate(&oldNetwork.Spec.NetworkConfig)...)

	// maximum allowed nodes with different name
	var maxDiff int
	// all old nodes names
	oldNodesNames := map[string]bool{}
	// nodes count in the old network spec
	oldNodesCount := len(oldNetwork.Spec.Nodes)
	// nodes count in the new network spec
	newNodesCount := len(n.Spec.Nodes)
	// nodes with different names than the old spec
	differentNodes := map[string]int{}

	if newNodesCount > oldNodesCount {
		maxDiff = newNodesCount - oldNodesCount
	}

	for _, node := range oldNetwork.Spec.Nodes {
		oldNodesNames[node.Name] = true
	}

	for i, node := range n.Spec.Nodes {
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

	return apierrors.NewInvalid(schema.GroupKind{}, n.Name, allErrors)

}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (n *Network) ValidateDelete() error {
	networklog.Info("validate delete", "name", n.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
