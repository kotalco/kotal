#!/bin/sh

set -e

ipfs config Addresses.API /ip4/$IPFS_API_HOST/tcp/$IPFS_API_PORT
