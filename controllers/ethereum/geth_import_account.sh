#!/bin/sh

set -e

if [ -z "$(ls -A $KOTAL_DATA_PATH/keystore)" ]
then
	echo "importing account"
	geth account import --datadir $KOTAL_DATA_PATH --password $KOTAL_SECRETS_PATH/account.password $KOTAL_SECRETS_PATH/account.key
else
	echo "account has been imported before!"
fi