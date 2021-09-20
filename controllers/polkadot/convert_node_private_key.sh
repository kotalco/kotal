#!/bin/sh

set -e

mkdir -p $DATA_PATH

# convert node private key to binary format
xxd -r -p -c 32 $SECRETS_PATH/nodekey > $DATA_PATH/kotal_nodekey