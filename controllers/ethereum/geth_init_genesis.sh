#!/bin/sh

set -e

if [ ! -d $KOTAL_DATA_PATH/geth ]
then
	echo "initializing geth genesis block"
	geth init --datadir $KOTAL_DATA_PATH $KOTAL_CONFIG_PATH/genesis.json
else
	echo "genesis block has been initialized before!"
fi