#!/bin/sh

set -e

if [ -e $IPFS_CLUSTER_PATH/service.json ]
then
	echo "ipfs cluster config has already been initialized"
else
	echo "initializing ipfs cluster config"
	ipfs-cluster-service init --consensus $IPFS_CLUSTER_CONSENSUS
fi