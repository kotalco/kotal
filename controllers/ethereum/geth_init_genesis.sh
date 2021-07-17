#!/bin/sh

set -e

if [ ! -d $DATA_PATH/geth ]
then
	echo "initializing geth genesis block"
	geth init --datadir $DATA_PATH $CONFIG_PATH/genesis.json
else
	echo "genesis block has been initialized before!"
fi