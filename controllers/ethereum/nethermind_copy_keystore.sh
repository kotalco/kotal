#!/bin/sh

set -e

mkdir -p $KOTAL_DATA_PATH/keystore

cp $KOTAL_SECRETS_PATH/account  $KOTAL_DATA_PATH/keystore/key-$KOTAL_COINBASE