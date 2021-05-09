#!/bin/sh

set -e

if [ -z "$(ls -A $DATA_PATH/keystore)" ]
then
	echo "importing account"
	geth account import --datadir $DATA_PATH --password $SECRETS_PATH/account.password $SECRETS_PATH/account.key
else
	echo "account has been imported before!"
fi