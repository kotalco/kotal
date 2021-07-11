#!/bin/sh

set -e

mkdir -p $DATA_PATH/keystore

cp $SECRETS_PATH/account  $DATA_PATH/keystore/key-$COINBASE