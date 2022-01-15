#!/bin/sh

set -e

echo "Copying node key from secrets dir to data dir"

mkdir -p ${KOTAL_DATA_PATH}
cp ${KOTAL_SECRETS_PATH}/node_key.json ${KOTAL_DATA_PATH}
