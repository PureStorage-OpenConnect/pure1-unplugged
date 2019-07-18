#!/usr/bin/env bash

set -xe

# !! This needs to be run in CentOS, matching the targeted version !!
if [[ -z "$(grep CentOS /etc/*-release)" ]]; then
    echo "ERROR: This script MUST be run on CentOS to build the CentOS based ISO."
    exit 1
fi

SCRIPT_DIR=$(dirname "$0")
REPO_ROOT=${SCRIPT_DIR}/../..
VERSION=$(${REPO_ROOT}/scripts/generate_version.sh ${REPO_ROOT})

BUILD_DIR=${REPO_ROOT}/build/rpm/
if [[ -d "${BUILD_DIR}" ]]; then
    rm -rf "${BUILD_DIR}"
fi
mkdir -p ${BUILD_DIR}

BUILD_TMP_DIR=${BUILD_DIR}/tmp
mkdir -p ${BUILD_TMP_DIR}

# Generate the pure1-unplugged RPM using fpm (aka easy-mode)
fpm \
    --workdir ${BUILD_TMP_DIR} \
    --debug \
    -s dir \
    -t rpm \
    -n pure1-unplugged \
    -v ${VERSION} \
    -p ${BUILD_DIR}/pure1-unplugged-${VERSION}-x86_64.rpm \
    -d python \
    -d ansible \
    -m "support@purestorage.com" \
    --url "pure1.purestorage.com" \
    --directories /opt/pure1-unplugged \
    --directories /etc/pure1-unplugged \
    --config-files /etc/pure1-unplugged/config.yaml \
    ${REPO_ROOT}/build/bundle/pure1-unplugged-${VERSION}/=/

# If we made it this far the tmp dir is empty, we can remove it
rm -rf ${BUILD_TMP_DIR}
