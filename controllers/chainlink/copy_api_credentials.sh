#!/bin/sh

set -e

mkdir -p ${KOTAL_DATA_PATH}
echo $KOTAL_API_EMAIL > ${KOTAL_DATA_PATH}/.api
cat ${KOTAL_SECRETS_PATH}/api-password >> ${KOTAL_DATA_PATH}/.api
