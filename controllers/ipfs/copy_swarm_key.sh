#!/bin/sh

set -e

mkdir -p $IPFS_PATH &&
cp $SECRETS_PATH/swarm.key $IPFS_PATH