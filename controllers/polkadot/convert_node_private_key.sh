#!/bin/sh

set -e

mkdir -p $KOTAL_DATA_PATH

# convert node private key to binary format
xxd -r -p -c 32 $KOTAL_SECRETS_PATH/nodekey > $KOTAL_DATA_PATH/kotal_nodekey