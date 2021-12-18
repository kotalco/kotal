package ethereum

import (
	"encoding/json"
	"fmt"
	"math/big"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
)

type ParityGenesis struct{}

// NormalizeNonce normalizes nonce to be 8 bytes (16 hex digits)
func (p *ParityGenesis) NormalizeNonce(data string) string {
	n := new(big.Int)
	i, _ := n.SetString(data, 16)
	return fmt.Sprintf("%#0.16x", i)
}

// Genesis returns genesis config parameter
func (p *ParityGenesis) Genesis(node *ethereumv1alpha1.Node) (content string, err error) {
	genesis := node.Spec.Genesis
	extraData := "0x00"
	var engineConfig map[string]interface{}

	// clique PoA settings
	if genesis.Clique != nil {
		extraData = createExtraDataFromSigners(genesis.Clique.Signers)
		engineConfig = map[string]interface{}{
			"clique": map[string]interface{}{
				"params": map[string]interface{}{
					"period": genesis.Clique.BlockPeriod,
					"epoch":  genesis.Clique.EpochLength,
				},
			},
		}
	}

	hex := func(n uint) string {
		return fmt.Sprintf("%#x", n)
	}

	tingerineWhistleBlock := hex(genesis.Forks.EIP150)
	spuriousDragonBlock := hex(genesis.Forks.EIP155)
	homesteadBlock := hex(genesis.Forks.Homestead)
	byzantiumBlock := hex(genesis.Forks.Byzantium)
	constantinopleBlock := hex(genesis.Forks.Constantinople)
	petersburgBlock := hex(genesis.Forks.Petersburg)
	istanbulBlock := hex(genesis.Forks.Istanbul)
	muirGlacierBlock := hex(genesis.Forks.MuirGlacier)
	berlinBlock := hex(genesis.Forks.Berlin)
	londonBlock := hex(genesis.Forks.London)
	arrowGlacierBlock := hex(genesis.Forks.ArrowGlacier)

	// ethash PoW settings
	if genesis.Ethash != nil {
		params := map[string]interface{}{
			"minimumDifficulty":      "0x020000",
			"difficultyBoundDivisor": "0x0800",
			"durationLimit":          "0x0d",
			"blockReward": map[string]string{
				tingerineWhistleBlock: "0x4563918244f40000",
				byzantiumBlock:        "0x29a2241af62c0000",
				constantinopleBlock:   "0x1bc16d674ec80000",
			},
			"homesteadTransition": homesteadBlock,
			"eip100bTransition":   byzantiumBlock,
			"difficultyBombDelays": map[string]string{
				byzantiumBlock:      "0x2dc6c0",
				constantinopleBlock: "0x1e8480",
				muirGlacierBlock:    "0x3d0900",
				londonBlock:         "0xaae60",
				arrowGlacierBlock:   "0xf4240",
			},
		}

		if genesis.Forks.DAO != nil {
			params["daoHardforkTransition"] = hex(*genesis.Forks.DAO)
		}

		engineConfig = map[string]interface{}{
			"Ethash": map[string]interface{}{
				"params": params,
			},
		}
	}

	genesisConfig := map[string]interface{}{
		"seal": map[string]interface{}{
			"ethereum": map[string]interface{}{
				"nonce":   p.NormalizeNonce(string(genesis.Nonce)[2:]),
				"mixHash": genesis.MixHash,
			},
		},
		"parentHash": "0x0000000000000000000000000000000000000000000000000000000000000000",
		"timestamp":  genesis.Timestamp,
		"gasLimit":   genesis.GasLimit,
		"difficulty": genesis.Difficulty,
		"author":     genesis.Coinbase,
		"extraData":  extraData,
	}

	// specify base fee per gas if london fork is activated at genesis block
	// https://github.com/openethereum/openethereum/issues/440
	if genesis.Forks.London == 0 {
		genesisConfig["baseFeePerGas"] = "0x3B9ACA00"
	}

	paramsConfig := map[string]interface{}{
		// other non fork parameters
		"chainID":              hex(genesis.ChainID),
		"accountStartNonce":    "0x00",
		"gasLimitBoundDivisor": "0x0400",
		"maximumExtraDataSize": "0xffff",
		"minGasLimit":          "0x1388",
		"networkID":            hex(node.Spec.Genesis.NetworkID),
		// Tingerine Whistle
		"eip150Transition": tingerineWhistleBlock,
		// Spurious Dragon
		"eip155Transition":      spuriousDragonBlock,
		"eip160Transition":      spuriousDragonBlock,
		"eip161abcTransition":   spuriousDragonBlock,
		"eip161dTransition":     spuriousDragonBlock,
		"maxCodeSizeTransition": spuriousDragonBlock, //eip170
		"maxCodeSize":           "0x6000",
		// Byzantium
		"eip140Transition": byzantiumBlock,
		"eip211Transition": byzantiumBlock,
		"eip214Transition": byzantiumBlock,
		"eip658Transition": byzantiumBlock,
		// Constantinople
		"eip145Transition":  constantinopleBlock,
		"eip1014Transition": constantinopleBlock,
		"eip1052Transition": constantinopleBlock,
		"eip1283Transition": constantinopleBlock,
		// PetersBurg
		"eip1283DisableTransition": petersburgBlock,
		// Istanbul
		"eip1283ReenableTransition": istanbulBlock,
		"eip1344Transition":         istanbulBlock,
		"eip1706Transition":         istanbulBlock,
		"eip1884Transition":         istanbulBlock,
		"eip2028Transition":         istanbulBlock,
		// Berlin
		"eip2315Transition": berlinBlock, // Simple Subroutines for the EVM
		"eip2929Transition": berlinBlock, // Gas cost increases for state access opcodes
		"eip2930Transition": berlinBlock, // Access lists. Requires eips 2718 (Typed Transaction Envelope), and 2929
		// London
		"eip1559Transition":                  londonBlock, // Fee market
		"eip3198Transition":                  londonBlock, // BASEFEE opcode
		"eip3541Transition":                  londonBlock, // Reject new contracts starting with the 0xEF byte
		"eip3529Transition":                  londonBlock, // Reduction in refunds
		"eip1559BaseFeeMaxChangeDenominator": "0x8",
		"eip1559ElasticityMultiplier":        "0x2",
		"eip1559BaseFeeInitialValue":         "0x3B9ACA00",
	}

	alloc := genesisAccounts(true, genesis.Forks)
	for _, account := range genesis.Accounts {
		m := map[string]interface{}{
			"balance": account.Balance,
		}

		if account.Code != "" {
			m["code"] = account.Code
		}

		if account.Storage != nil {
			m["storage"] = account.Storage
		}

		alloc[string(account.Address)] = m
	}

	result := map[string]interface{}{
		"name":     "network",
		"genesis":  genesisConfig,
		"params":   paramsConfig,
		"engine":   engineConfig,
		"accounts": alloc,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return
	}

	content = string(data)

	return
}
