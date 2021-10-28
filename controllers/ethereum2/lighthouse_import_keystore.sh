#!/bin/sh

set -e

lighthouse account validator import --datadir ${KOTAL_DATA_PATH} --network ${KOTAL_NETWORK} \
--keystore ${KOTAL_KEY_DIR}/keystore-${KOTAL_KEYSTORE_INDEX}.json \
--reuse-password \
--password-file ${KOTAL_KEY_DIR}/password.txt