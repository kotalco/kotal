package controllers

import (
	"bytes"
	"text/template"
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
		GenesisDir: PathGenesisFile,
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

const importAccountScriptTemplate = `
#!/bin/sh

set -e

if [ -z "$(ls -A {{.DataDir}}/keystore)" ]
then
	echo "importing account"
	geth account import --datadir {{.DataDir}} --password {{.ImportDir}}/account.password {{.ImportDir}}/account.key
else
	echo "account has been imported before!"
fi
`

// ImportAccountInput is the input for init genesis block script
type ImportAccountInput struct {
	DataDir   string
	ImportDir string
}

// generateImportAccountScript generates init genesis block script
func generateImportAccountScript() (script string, err error) {

	input := &ImportAccountInput{
		DataDir:   PathBlockchainData,
		ImportDir: PathGenesisFile,
	}

	tmpl, err := template.New("master").Parse(importAccountScriptTemplate)
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
