#!/bin/sh

set -e

mkdir -p ${KOTAL_VALIDATORS_PATH}
cp -RL ${KOTAL_SECRETS_PATH}/validator-keys ${KOTAL_VALIDATORS_PATH}
cp -RL ${KOTAL_SECRETS_PATH}/validator-secrets ${KOTAL_VALIDATORS_PATH}