#!/bin/sh

set -e

validator accounts import --accept-terms-of-use \
--${KOTAL_NETWORK} \
--wallet-dir=${KOTAL_DATA_PATH}/prysm-wallet \
--keys-dir=${KOTAL_KEY_DIR}/keystore-${KOTAL_KEYSTORE_INDEX}.json \
--account-password-file=${KOTAL_KEY_DIR}/password.txt \
--wallet-password-file=${KOTAL_SECRETS_PATH}/prysm-wallet/prysm-wallet-password.txt