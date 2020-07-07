package helpers

import (
	"crypto/ecdsa"
	"errors"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func derive(fromPrivateKey string) (publicKeyECDSA *ecdsa.PublicKey, err error) {
	// private key
	privateKey, err := crypto.HexToECDSA(fromPrivateKey)
	if err != nil {
		return
	}

	// public key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		err = errors.New("publicKey is not of type *ecdsa.PublicKey")
		return
	}

	return
}

// DerivePublicKey drives node public key from private key
func DerivePublicKey(fromPrivateKey string) (publicKeyHex string, err error) {
	publicKeyECDSA, err := derive(fromPrivateKey)
	if err != nil {
		return
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	publicKeyHex = hexutil.Encode(publicKeyBytes)[4:]

	return

}

// DeriveAddress drives ethereum address from private key
func DeriveAddress(fromPrivateKey string) (addressHex string, err error) {
	publicKeyECDSA, err := derive(fromPrivateKey)
	if err != nil {
		return
	}

	addressHex = crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	return

}
