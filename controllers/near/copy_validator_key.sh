#!/bin/sh

set -e

echo "Copying validator key from secrets dir to data dir"

mkdir -p ${KOTAL_DATA_PATH}
cp ${KOTAL_SECRETS_PATH}/validator_key.json ${KOTAL_DATA_PATH}
