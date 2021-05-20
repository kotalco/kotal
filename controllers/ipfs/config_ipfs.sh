#!/bin/sh

set -e

ipfs config Addresses.API /ip4/$IPFS_API_HOST/tcp/$IPFS_API_PORT
ipfs config Addresses.Gateway /ip4/$IPFS_GATEWAY_HOST/tcp/$IPFS_GATEWAY_PORT

export IFS=";"
for profile in $IPFS_PROFILES; do
    ipfs config profile apply $profile
    echo "$profile profile has been applied"
done
