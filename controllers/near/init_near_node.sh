#!/bin/sh

set -e


if [ -z "$(ls -A $KOTAL_DATA_PATH/genesis.json)" ]
then
    echo "Initializing NEAR node"
	neard init --chain-id $KOTAL_NEAR_NETWORK --download-genesis --download-config
else
	echo "NEAR node has already been initialized before!"
fi