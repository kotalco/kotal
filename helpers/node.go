package helpers

import (
	"crypto/ecdsa"
	"errors"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// CreateNodeKeypair creates private key for node to be used for enodeURL
func CreateNodeKeypair(hex string) (privateKeyHex, publicKeyHex string, err error) {
	// private key
	var privateKey *ecdsa.PrivateKey

	if hex != "" {
		privateKey, err = crypto.HexToECDSA(hex)
		if err != nil {
			return
		}
		privateKeyHex = hex
	} else {
		privateKey, err = crypto.GenerateKey()
		if err != nil {
			return
		}
		privateKeyBytes := crypto.FromECDSA(privateKey)
		privateKeyHex = hexutil.Encode(privateKeyBytes)[2:]
	}

	// public key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		err = errors.New("publicKey is not of type *ecdsa.PublicKey")
		return
	}
	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	publicKeyHex = hexutil.Encode(publicKeyBytes)[4:]

	return

}
