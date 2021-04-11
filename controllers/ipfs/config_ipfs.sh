#!/bin/sh

set -e

ipfs config Addresses.API /ip4/0.0.0.0/tcp/$IPFS_API_PORT
