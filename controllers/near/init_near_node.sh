#!/bin/sh

set -e


if [ -z "$(ls -A $KOTAL_DATA_PATH/genesis.json)" ]
then
    echo "Initializing NEAR node"
	neard --home $KOTAL_DATA_PATH init --chain-id $KOTAL_NEAR_NETWORK --download-genesis --download-config --account-id validator
else
	echo "NEAR node has already been initialized before!"
fi