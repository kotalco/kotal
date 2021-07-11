#!/bin/sh

set -e

mkdir -p $DATA_PATH

# convert enode private key to binary format
# nethermind doesn't accept text format
# more info: https://discord.com/channels/629004402170134531/629004402170134537/862516237477347338
xxd -r -p -c 32 $SECRETS_PATH/nodekey > $DATA_PATH/kotal_nodekey