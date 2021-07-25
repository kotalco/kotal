package v1alpha1

import (
	"fmt"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

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

// Validate validates network genesis block spec
func (g *Genesis) Validate() field.ErrorList {

	var allErrors field.ErrorList

	// don't use existing network chain id
	if chain := ChainByID[g.ChainID]; chain != "" {
		err := field.Invalid(field.NewPath("spec").Child("genesis").Child("chainId"), fmt.Sprintf("%d", g.ChainID), fmt.Sprintf("can't use chain id of %s network to avoid tx replay", chain))
		allErrors = append(allErrors, err)
	}

	// validate forks order
	allErrors = append(allErrors, g.ValidateForksOrder()...)
	return allErrors
}

// ValidateForksOrder validates that forks are in correct order
func (g *Genesis) ValidateForksOrder() field.ErrorList {
	var orderErrors field.ErrorList
	forks := g.Forks

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
		"berlin",
		"london",
	}

	// milestones at the correct order
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
		forks.Berlin,
		forks.London,
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
