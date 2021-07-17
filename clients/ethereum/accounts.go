package ethereum

import (
	"fmt"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
)

// genesisAccounts returns genesis config accounts
func genesisAccounts(withBuiltins bool, forks *ethereumv1alpha1.Forks) map[string]interface{} {
	accounts := map[string]interface{}{}
	for i := 0; i < 256; i++ {
		address := fmt.Sprintf("%#040x", i)
		var fn map[string]interface{}
		if i >= 1 && i <= 9 {
			if withBuiltins {
				fn = builtinFunction(i, forks)
			}
			accounts[address] = map[string]interface{}{
				"balance": "0x1",
				"builtin": fn,
			}
		} else {
			accounts[address] = map[string]interface{}{
				"balance": "0x1",
			}
		}
	}
	return accounts
}

// builtinFunction returns built in parity functions
func builtinFunction(i int, forks *ethereumv1alpha1.Forks) map[string]interface{} {
	switch i {
	case 1:
		return ecRecover()
	case 2:
		return sha256()
	case 3:
		return ripemd160()
	case 4:
		return identity()
	case 5:
		return modexp(forks)
	case 6:
		return altBn128Add(forks)
	case 7:
		return altBn128Mul(forks)
	case 8:
		return altBn128Pairing(forks)
	case 9:
		return blake2(forks)
	}
	return nil
}

// ecRecover revcovers public key from elliptic curve signature
func ecRecover() map[string]interface{} {
	return map[string]interface{}{
		"name": "ecrecover",
		"pricing": map[string]interface{}{
			"linear": map[string]int{
				"base": 3000,
				"word": 0,
			},
		},
	}
}

// sha256 is sha256 function
func sha256() map[string]interface{} {
	return map[string]interface{}{
		"name": "sha256",
		"pricing": map[string]interface{}{
			"linear": map[string]int{
				"base": 60,
				"word": 12,
			},
		},
	}
}

// ripemd160 is ripemd160 hash function
func ripemd160() map[string]interface{} {
	return map[string]interface{}{
		"name": "ripemd160",
		"pricing": map[string]interface{}{
			"linear": map[string]int{
				"base": 600,
				"word": 120,
			},
		},
	}
}

// identity function
func identity() map[string]interface{} {
	return map[string]interface{}{
		"name": "identity",
		"pricing": map[string]interface{}{
			"linear": map[string]int{
				"base": 15,
				"word": 3,
			},
		},
	}
}

// modexp is modular exponentiaiton function
func modexp(forks *ethereumv1alpha1.Forks) map[string]interface{} {
	byzantiumBlock := fmt.Sprintf("%#x", forks.Byzantium)
	berlinBlock := fmt.Sprintf("%#x", forks.Berlin)

	return map[string]interface{}{
		"name": "modexp",
		"pricing": map[string]interface{}{
			byzantiumBlock: map[string]interface{}{
				"info": "EIP-198: Big integer modular exponentiation.",
				"price": map[string]interface{}{
					"modexp": map[string]uint{
						"divisor": 20,
					},
				},
			},
			berlinBlock: map[string]interface{}{
				"info": "EIP-2565: ModExp Gas Cost.",
				"price": map[string]interface{}{
					"modexp2565": map[string]interface{}{},
				},
			},
		},
	}
}

// altBn128Add is elliptic curve add function
func altBn128Add(forks *ethereumv1alpha1.Forks) map[string]interface{} {

	byzantiumBlock := fmt.Sprintf("%#x", forks.Byzantium)
	istanbulBlock := fmt.Sprintf("%#x", forks.Istanbul)

	return map[string]interface{}{
		"name": "alt_bn128_add",
		"pricing": map[string]interface{}{
			byzantiumBlock: map[string]interface{}{
				"price": map[string]interface{}{
					"alt_bn128_const_operations": map[string]int{
						"price": 500,
					},
				},
			},
			istanbulBlock: map[string]interface{}{
				"info": "EIP 1108 transition",
				"price": map[string]interface{}{
					"alt_bn128_const_operations": map[string]int{
						"price": 150,
					},
				},
			},
		},
	}
}

// altBn128Mul is elliptic function Multiplication funciton
func altBn128Mul(forks *ethereumv1alpha1.Forks) map[string]interface{} {

	byzantiumBlock := fmt.Sprintf("%#x", forks.Byzantium)
	istanbulBlock := fmt.Sprintf("%#x", forks.Istanbul)

	return map[string]interface{}{
		"name": "alt_bn128_mul",
		"pricing": map[string]interface{}{
			byzantiumBlock: map[string]interface{}{
				"price": map[string]interface{}{
					"alt_bn128_const_operations": map[string]int{
						"price": 40000,
					},
				},
			},
			istanbulBlock: map[string]interface{}{
				"info": "EIP 1108 transition",
				"price": map[string]interface{}{
					"alt_bn128_const_operations": map[string]int{
						"price": 6000,
					},
				},
			},
		},
	}
}

// altBn128Pairing is elliptic curve pairing function
func altBn128Pairing(forks *ethereumv1alpha1.Forks) map[string]interface{} {

	byzantiumBlock := fmt.Sprintf("%#x", forks.Byzantium)
	istanbulBlock := fmt.Sprintf("%#x", forks.Istanbul)

	return map[string]interface{}{
		"name": "alt_bn128_pairing",
		"pricing": map[string]interface{}{
			byzantiumBlock: map[string]interface{}{
				"price": map[string]interface{}{
					"alt_bn128_pairing": map[string]int{
						"base": 100000,
						"pair": 80000,
					},
				},
			},
			istanbulBlock: map[string]interface{}{
				"info": "EIP 1108 transition",
				"price": map[string]interface{}{
					"alt_bn128_pairing": map[string]int{
						"base": 45000,
						"pair": 34000,
					},
				},
			},
		},
	}
}

// blake2 is blake hash function
func blake2(forks *ethereumv1alpha1.Forks) map[string]interface{} {

	return map[string]interface{}{
		"name":        "blake2_f",
		"activate_at": fmt.Sprintf("%#x", forks.Istanbul),
		"pricing": map[string]interface{}{
			"blake2_f": map[string]interface{}{
				"gas_per_round": 1,
			},
		},
	}
}
