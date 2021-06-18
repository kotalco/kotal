package v1alpha1

import (
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

// validate validates network config
func (n *NetworkConfig) validate() field.ErrorList {
	var validateErrors field.ErrorList

	// consensus: can't specify consensus while joining existing network
	if n.Network != "" && n.Consensus != "" {
		err := field.Invalid(field.NewPath("spec").Child("consensus"), n.Consensus, "must be none while joining a network")
		validateErrors = append(validateErrors, err)
	}

	// genesis: must specify genesis if there's no network to join
	if n.Network == "" && n.Genesis == nil {
		err := field.Invalid(field.NewPath("spec").Child("genesis"), "", "must be specified if spec.network is none")
		validateErrors = append(validateErrors, err)
	}

	// id: must be provided if join is none
	if n.Network == "" && n.ID == 0 {
		err := field.Invalid(field.NewPath("spec").Child("id"), "", "must be specified if spec.network is none")
		validateErrors = append(validateErrors, err)
	}

	// id: must be none if network is provided
	if n.Network != "" && n.ID != 0 {
		err := field.Invalid(field.NewPath("spec").Child("id"), fmt.Sprintf("%d", n.ID), "must be none if spec.network is provided")
		validateErrors = append(validateErrors, err)
	}

	// consensus: must be provided if genesis is provided
	if n.Genesis != nil && n.Consensus == "" {
		err := field.Invalid(field.NewPath("spec").Child("consensus"), "", "must be specified if spec.genesis is provided")
		validateErrors = append(validateErrors, err)
	}

	// validate non nil genesis
	if n.Genesis != nil {
		validateErrors = append(validateErrors, n.Genesis.Validate(n)...)
	}

	return validateErrors
}

// ValidateCreate validates network config
func (n *NetworkConfig) ValidateCreate() field.ErrorList {
	var errors field.ErrorList

	errors = append(errors, n.validate()...)

	return errors
}

// ValidateUpdate validates network config update
func (n *NetworkConfig) ValidateUpdate(oldConfig *NetworkConfig) field.ErrorList {
	var updateErrors field.ErrorList

	updateErrors = append(updateErrors, n.validate()...)

	if oldConfig.ID != n.ID {
		err := field.Invalid(field.NewPath("spec").Child("id"), fmt.Sprintf("%d", n.ID), "field is immutable")
		updateErrors = append(updateErrors, err)
	}

	if oldConfig.Network != n.Network {
		err := field.Invalid(field.NewPath("spec").Child("network"), n.Network, "field is immutable")
		updateErrors = append(updateErrors, err)
	}

	if oldConfig.Consensus != n.Consensus {
		err := field.Invalid(field.NewPath("spec").Child("consensus"), n.Consensus, "field is immutable")
		updateErrors = append(updateErrors, err)
	}

	// TODO: move to genesis.ValidateUpdate
	// TODO: genesis forks can change, new forks can be scheduled in the future
	if !reflect.DeepEqual(n.Genesis, oldConfig.Genesis) {
		err := field.Invalid(field.NewPath("spec").Child("genesis"), "", "field is immutable")
		updateErrors = append(updateErrors, err)
	}

	return updateErrors

}
