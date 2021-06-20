package v1alpha1

import (
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// +kubebuilder:webhook:verbs=create;update,path=/validate-ethereum2-kotal-io-v1alpha1-beaconnode,mutating=false,failurePolicy=fail,groups=ethereum2.kotal.io,resources=beaconnodes,versions=v1alpha1,name=validate-ethereum2-v1alpha1-beaconnode.kb.io,sideEffects=None,admissionReviewVersions=v1

var _ webhook.Validator = &BeaconNode{}

// validate is the shared validate create and update logic
func (r *BeaconNode) validate() field.ErrorList {
	var nodeErrors field.ErrorList

	path := field.NewPath("spec")

	// rest is supported by teku and lighthouse only
	if r.Spec.REST && r.Spec.Client != TekuClient && r.Spec.Client != LighthouseClient {
		err := field.Invalid(path.Child("rest"), r.Spec.REST, fmt.Sprintf("not supported by %s client", r.Spec.Client))
		nodeErrors = append(nodeErrors, err)
	}

	// rpc is supported by nimbus and prysm only
	if r.Spec.RPC && r.Spec.Client != NimbusClient && r.Spec.Client != PrysmClient {
		err := field.Invalid(path.Child("rpc"), r.Spec.RPC, fmt.Sprintf("not supported by %s client", r.Spec.Client))
		nodeErrors = append(nodeErrors, err)
	}

	// grpc is supported by prysm only
	if r.Spec.GRPC && r.Spec.Client != PrysmClient {
		err := field.Invalid(path.Child("grpc"), r.Spec.GRPC, fmt.Sprintf("not supported by %s client", r.Spec.Client))
		nodeErrors = append(nodeErrors, err)
	}

	// rpc is always on in prysm
	if r.Spec.Client == PrysmClient && !r.Spec.RPC {
		err := field.Invalid(path.Child("rpc"), r.Spec.RPC, "can't be disabled in prysm client")
		nodeErrors = append(nodeErrors, err)
	}

	// eth1 endpoint is required by prysm if network is not mainnet
	if r.Spec.Client == PrysmClient && len(r.Spec.Eth1Endpoints) == 0 && r.Spec.Network != "mainnet" {
		err := field.Invalid(path.Child("eth1Endpoints"), "", fmt.Sprintf("required by %s client if network is not mainnet", r.Spec.Client))
		nodeErrors = append(nodeErrors, err)
	}

	// teku and nimbus doesn't support multiple Ethereum 1 endpoints
	if len(r.Spec.Eth1Endpoints) > 1 && r.Spec.Client == NimbusClient {
		err := field.Invalid(path.Child("eth1Endpoints"), strings.Join(r.Spec.Eth1Endpoints, ", "), fmt.Sprintf("multiple Ethereum 1 endpoints not supported by %s client", r.Spec.Client))
		nodeErrors = append(nodeErrors, err)
	}

	return nodeErrors
}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *BeaconNode) ValidateCreate() error {
	var allErrors field.ErrorList

	nodelog.Info("validate create", "name", r.Name)

	allErrors = append(allErrors, r.validate()...)
	allErrors = append(allErrors, r.Spec.Resources.ValidateCreate()...)

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)

}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *BeaconNode) ValidateUpdate(old runtime.Object) error {
	var allErrors field.ErrorList
	oldNode := old.(*BeaconNode)
	path := field.NewPath("spec")

	nodelog.Info("validate update", "name", r.Name)

	allErrors = append(allErrors, r.validate()...)
	allErrors = append(allErrors, r.Spec.Resources.ValidateUpdate(&oldNode.Spec.Resources)...)

	if oldNode.Spec.Client != r.Spec.Client {
		err := field.Invalid(path.Child("client"), r.Spec.Client, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if oldNode.Spec.Network != r.Spec.Network {
		err := field.Invalid(path.Child("network"), r.Spec.Network, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if len(allErrors) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{}, r.Name, allErrors)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *BeaconNode) ValidateDelete() error {
	nodelog.Info("validate delete", "name", r.Name)

	return nil
}
