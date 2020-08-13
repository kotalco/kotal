package controllers

const initScriptTemplate = `
#!/bin/sh

set -e

if [ -e /data/ipfs/config ]
then
	echo "ipfs repo has already been initialized"
else 
	echo "initializing ipfs repo"
	ipfs init
fi

echo "adding bootstrap swarm peers"
{{ range .Peers }}
	ipfs bootstrap add {{ . }}
{{ end }}

echo "applying profiles"
{{ range .Profiles }}
	ipfs config profile apply {{ . }}
{{ end }}
`
