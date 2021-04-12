#!/bin/sh

set -e

if [ -e $IPFS_PATH/config ]
then
	echo "ipfs config has already been initialized"
else 
	echo "initializing ipfs config"
	ipfs init --empty-repo --profile $IPFS_INIT_PROFILES
fi