package v1alpha1

import (
	"fmt"
	"math/big"
	"reflect"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/util/validation/field"
)

var (
	// ChainByID is public chains indexed by ID
	ChainByID = map[uint]string{
		1:        MainNetwork,
		3:        RopstenNetwork,
		4:        RinkebyNetwork,
		5:        GoerliNetwork,
		6:        KottiNetwork,
		61:       ClassicNetwork,
		63:       MordorNetwork,
		2018:     DevNetwork,
		11155111: SepoliaNetwork,
	}
)

// EnabledConsensusConfigs returns enabled consensus configs
func (g *Genesis) EnabledConsensusConfigs() []string {
	configs := map[string]bool{
		"ethash": g.Ethash != nil,
		"clique": g.Clique != nil,
		"ibft2":  g.IBFT2 != nil,
	}

	enabledConfigs := []string{}

	for consensus, enabled := range configs {
		if enabled {
			enabledConfigs = append(enabledConfigs, consensus)
		}
	}

	return enabledConfigs
}

// ReservedAccountIsUsed returns true if reserved account is used
// reserved accounts are accounts from 0x00...01 to 0x00...ff (1 to 256)
// reserved accounts are used for precompiles
func (g *Genesis) ReservedAccountIsUsed() (bool, string) {
	// space of reserved addresses
	space := new(big.Int)
	space.SetInt64(256)

	for _, account := range g.Accounts {
		address := string(account.Address)
		i := new(big.Int)
		i.SetString(address[2:], 16)
		// address must be outside (with greater int value) reserved space
		if i.Cmp(space) != 1 {
			return true, address
		}
	}
	return false, ""
}

// validate validates network genesis block spec
func (g *Genesis) validate() field.ErrorList {

	var allErrors field.ErrorList

	// validate accounts from 0x00...01 to 0x00...ff are reserved
	if used, address := g.ReservedAccountIsUsed(); used {
		err := field.Invalid(field.NewPath("spec").Child("genesis").Child("accounts"), address, "reserved account is used")
		allErrors = append(allErrors, err)
	}

	// validate consensus config (ethash, clique, ibft2) is not missing
	// validate only one consensus configuration can be set
	// TODO: update this validation after suporting new consensus algorithm
	configs := g.EnabledConsensusConfigs()
	if len(configs) == 0 {
		err := field.Invalid(field.NewPath("spec").Child("genesis"), "", "consensus configuration (ethash, clique, or ibft2) is missing")
		allErrors = append(allErrors, err)
	} else if len(configs) > 1 {
		sort.Strings(configs)
		err := field.Invalid(field.NewPath("spec").Child("genesis"), "", fmt.Sprintf("multiple consensus configurations (%s) are enabled", strings.Join(configs, ", ")))
		allErrors = append(allErrors, err)
	}

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
		"arrowglacier",
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
		forks.ArrowGlacier,
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

// ValidateCreate validates genesis block during node creation
func (g *Genesis) ValidateCreate() field.ErrorList {
	var allErrors field.ErrorList

	allErrors = append(allErrors, g.validate()...)

	return allErrors
}

func (g *Genesis) ValidateUpdate(oldGenesis *Genesis) field.ErrorList {
	var allErrors field.ErrorList

	if g.Coinbase != oldGenesis.Coinbase {
		err := field.Invalid(field.NewPath("spec").Child("genesis").Child("coinbase"), g.Coinbase, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if g.Difficulty != oldGenesis.Difficulty {
		err := field.Invalid(field.NewPath("spec").Child("genesis").Child("difficulty"), g.Difficulty, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if g.MixHash != oldGenesis.MixHash {
		err := field.Invalid(field.NewPath("spec").Child("genesis").Child("mixHash"), g.MixHash, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if g.GasLimit != oldGenesis.GasLimit {
		err := field.Invalid(field.NewPath("spec").Child("genesis").Child("gasLimit"), g.GasLimit, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if g.Nonce != oldGenesis.Nonce {
		err := field.Invalid(field.NewPath("spec").Child("genesis").Child("nonce"), g.Nonce, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if g.Timestamp != oldGenesis.Timestamp {
		err := field.Invalid(field.NewPath("spec").Child("genesis").Child("timestamp"), g.Timestamp, "field is immutable")
		allErrors = append(allErrors, err)
	}

	if !reflect.DeepEqual(g.Accounts, oldGenesis.Accounts) {
		err := field.Invalid(field.NewPath("spec").Child("genesis").Child("accounts"), "", "field is immutable")
		allErrors = append(allErrors, err)
	}

	allErrors = append(allErrors, g.validate()...)

	return allErrors
}
