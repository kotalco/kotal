package controllers

import (
	"fmt"
)

// genesisAccounts returns genesis config accounts
func genesisAccounts(withBuiltins bool) map[string]interface{} {
	accounts := map[string]interface{}{}
	for i := 0; i < 256; i++ {
		address := fmt.Sprintf("%#040x", i)
		var fn map[string]interface{}
		if withBuiltins && i >= 0 && i <= 9 {
			fn = builtinFunction(i)
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
func builtinFunction(i int) map[string]interface{} {
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
		return modexp()
	case 6:
		return altBn128Add()
	case 7:
		return altBn128Mul()
	case 8:
		return altBn128Pairing()
	case 9:
		return blake2()
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

func modexp() map[string]interface{} {
	return map[string]interface{}{
		"name":        "modexp",
		"activate_at": "0x0",
		"pricing": map[string]interface{}{
			"modexp": map[string]int{
				"divisor": 20,
			},
		},
	}
}

func altBn128Add() map[string]interface{} {
	return map[string]interface{}{
		"name": "alt_bn128_add",
		"pricing": map[string]interface{}{
			"0": map[string]interface{}{
				"price": map[string]interface{}{
					"alt_bn128_const_operations": map[string]int{
						"price": 500,
					},
				},
			},
			"0x17d433": map[string]interface{}{
				"info": "EIP 1108 transition at block 1_561_651 (0x17d433)",
				"price": map[string]interface{}{
					"alt_bn128_const_operations": map[string]int{
						"price": 150,
					},
				},
			},
		},
	}
}

func altBn128Mul() map[string]interface{} {
	return map[string]interface{}{
		"name": "alt_bn128_mul",
		"pricing": map[string]interface{}{
			"0": map[string]interface{}{
				"price": map[string]interface{}{
					"alt_bn128_const_operations": map[string]int{
						"price": 40000,
					},
				},
			},
			"0x17d433": map[string]interface{}{
				"info": "EIP 1108 transition at block 1_561_651 (0x17d433)",
				"price": map[string]interface{}{
					"alt_bn128_const_operations": map[string]int{
						"price": 6000,
					},
				},
			},
		},
	}
}

func altBn128Pairing() map[string]interface{} {
	return map[string]interface{}{
		"name": "alt_bn128_pairing",
		"pricing": map[string]interface{}{
			"0": map[string]interface{}{
				"price": map[string]interface{}{
					"alt_bn128_pairing": map[string]int{
						"base": 100000,
						"pair": 80000,
					},
				},
			},
			"0x17d433": map[string]interface{}{
				"info": "EIP 1108 transition at block 1_561_651 (0x17d433)",
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

func blake2() map[string]interface{} {
	return map[string]interface{}{
		"name":        "blake2_f",
		"activate_at": "0x17d433",
		"pricing": map[string]interface{}{
			"blake2_f": map[string]interface{}{
				"gas_per_round": 1,
			},
		},
	}
}
