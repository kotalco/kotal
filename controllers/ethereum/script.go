package controllers

import (
	"bytes"
	"fmt"
	"text/template"

	ethereumv1alpha1 "github.com/kotalco/kotal/apis/ethereum/v1alpha1"
)

const initGenesisScriptTemplate = `
#!/bin/sh

set -e

if [ ! -d {{.DataDir}}/geth ]
then
	echo "initializing geth genesis block"
	geth init --datadir {{.DataDir}} {{.GenesisDir}}/genesis.json
else
	echo "genesis block has been initialized before!"
fi
`

// InitGenesisInput is the input for init genesis block script
type InitGenesisInput struct {
	DataDir    string
	GenesisDir string
}

// generateInitGenesisScript generates init genesis block script
func generateInitGenesisScript() (script string, err error) {

	input := &InitGenesisInput{
		DataDir:    PathBlockchainData,
		GenesisDir: PathConfig,
	}

	tmpl, err := template.New("master").Parse(initGenesisScriptTemplate)
	if err != nil {
		return
	}

	buff := new(bytes.Buffer)
	if err = tmpl.Execute(buff, input); err != nil {
		return
	}

	script = buff.String()

	return
}

func generateImportAccountScript(client ethereumv1alpha1.EthereumClient) (script string, err error) {
	switch client {
	case ethereumv1alpha1.GethClient:
		return generateGethImportAccountScript()
	case ethereumv1alpha1.ParityClient:
		return generateParityImportAccountScript()
	}

	err = fmt.Errorf("generating init genesis for client %s is not supported", client)
	return
}

const importGethAccountScriptTemplate = `
#!/bin/sh

set -e

if [ -z "$(ls -A {{.DataDir}}/keystore)" ]
then
	echo "importing account"
	geth account import --datadir {{.DataDir}} --password {{.SecretsDir}}/account.password {{.SecretsDir}}/account.key
else
	echo "account has been imported before!"
fi
`

// ImportGethAccountInput is the input for importing an account into geth client
type ImportGethAccountInput struct {
	DataDir    string
	SecretsDir string
}

// generateGethImportAccountScript generates init genesis block script
func generateGethImportAccountScript() (script string, err error) {

	input := &ImportGethAccountInput{
		DataDir:    PathBlockchainData,
		SecretsDir: PathSecrets,
	}

	tmpl, err := template.New("master").Parse(importGethAccountScriptTemplate)
	if err != nil {
		return
	}

	buff := new(bytes.Buffer)
	if err = tmpl.Execute(buff, input); err != nil {
		return
	}

	script = buff.String()

	return
}

const importParityAccountScriptTemplate = `
#!/bin/sh

set -e

if [ ! -f {{.DataDir}}/keys/ethereum/UTC* ]
then
	echo "importing account"
	/home/openethereum/openethereum account import --chain {{.ConfigDir}}/genesis.json --base-path {{.DataDir}} --password {{.SecretsDir}}/account.password {{.SecretsDir}}/account
else
	echo "account has been imported before!"
fi
`

// ImportParityAccountInput is the input for init genesis block script
type ImportParityAccountInput struct {
	ConfigDir  string
	DataDir    string
	SecretsDir string
}

// generateParityImportAccountScript generates init genesis block script
func generateParityImportAccountScript() (script string, err error) {

	input := &ImportParityAccountInput{
		ConfigDir:  PathConfig,
		DataDir:    PathBlockchainData,
		SecretsDir: PathSecrets,
	}

	tmpl, err := template.New("master").Parse(importParityAccountScriptTemplate)
	if err != nil {
		return
	}

	buff := new(bytes.Buffer)
	if err = tmpl.Execute(buff, input); err != nil {
		return
	}

	script = buff.String()

	return
}
