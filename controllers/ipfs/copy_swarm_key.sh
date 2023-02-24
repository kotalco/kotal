#!/bin/sh

set -e

mkdir -p $IPFS_PATH &&
cp $KOTAL_SECRETS_PATH/swarm.key $IPFS_PATH