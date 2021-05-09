#!/bin/sh

set -e

if [ ! -f $DATA_PATH/keys/ethereum/UTC* ]
then
	echo "importing account"
	/home/openethereum/openethereum account import --chain $CONFIG_PATH/genesis.json --base-path $DATA_PATH --password $SECRETS_PATH/account.password $SECRETS_PATH/account
else
	echo "account has been imported before!"
fi