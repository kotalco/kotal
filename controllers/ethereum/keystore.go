package controllers

import (
	"io/ioutil"
	"os"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
)

// KeyStoreFromPrivateKey generates key store from private key (hex without 0x)
func KeyStoreFromPrivateKey(key, password string) (content []byte, err error) {
	dir, err := ioutil.TempDir(os.TempDir(), "tmp")
	if err != nil {
		return
	}
	defer os.RemoveAll(dir)

	ks := keystore.NewKeyStore(dir, keystore.StandardScryptN, keystore.StandardScryptP)
	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return
	}

	acc, err := ks.ImportECDSA(privateKey, password)
	if err != nil {
		return
	}

	content, err = ks.Export(acc, password, password)

	return
}
