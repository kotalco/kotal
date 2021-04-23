#!/bin/sh

set -e

if [ -e /data/ipfs-cluster/service.json ]
then
	echo "ipfs cluster config has already been initialized"
else
	echo "initializing ipfs cluster config"
	ipfs-cluster-service init
fi