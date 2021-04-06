#!/bin/sh

set -e

# TODO: replace with data dir var
if [ -e /data/ipfs/config ]
then
	echo "ipfs config has already been initialized"
else 
	echo "initializing ipfs config"
	ipfs init --empty-repo
fi