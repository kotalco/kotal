package ethereum

import (
	"bytes"
	"encoding/hex"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/kotalco/kotal/apis/shared"
)

// createExtraDataFromSigners creates extraDta genesis field value from initial signers
func createExtraDataFromSigners(signers []shared.EthereumAddress) string {
	extraData := "0x"
	// vanity data
	extraData += strings.Repeat("00", 32)
	// signers
	for _, signer := range signers {
		// append address without the 0x
		extraData += string(signer)[2:]
	}
	// proposer signature
	extraData += strings.Repeat("00", 65)
	return extraData
}

// createExtraDataFromValidators creates extraDta genesis field value from initial validators
func createExtraDataFromValidators(validators []shared.EthereumAddress) (string, error) {
	data := []interface{}{}
	extraData := "0x"

	// empty vanity bytes
	vanity := bytes.Repeat([]byte{0x00}, 32)

	// validator addresses bytes
	decodedValidators := []interface{}{}
	for _, validator := range validators {
		validatorBytes, err := hex.DecodeString(string(validator)[2:])
		if err != nil {
			return extraData, err
		}
		decodedValidators = append(decodedValidators, validatorBytes)
	}

	// no vote
	var vote []byte

	// round 0, must be 4 bytes
	round := bytes.Repeat([]byte{0x00}, 4)

	// no committer seals
	committers := []interface{}{}

	// pack all required info into data
	data = append(data, vanity)
	data = append(data, decodedValidators)
	data = append(data, vote)
	data = append(data, round)
	data = append(data, committers)

	// rlp encode data
	payload, err := rlp.EncodeToBytes(data)
	if err != nil {
		return extraData, err
	}

	return extraData + common.Bytes2Hex(payload), nil

}
