#!/usr/bin/env bash

set -ex

SCRIPT_DIR=$(dirname "$0")
REPO_ROOT=$(realpath ${SCRIPT_DIR}/../..)
VERSION=$(${REPO_ROOT}/scripts/generate_version.sh ${REPO_ROOT})

${REPO_ROOT}/scripts/print_console_label.sh "Building pure1-unplugged install bundle"

# Use realpath for this so we can map it into a container
HOST_RESULTS_DIR=${REPO_ROOT}/build/bundle
CONTAINER_REPO_ROOT=/root/pure1-unplugged

function cleanup {
    # use sudo since the directory is potentially owned by root, fix that too..
    sudo chown -R $(id -un):$(id -gn) ${HOST_RESULTS_DIR}
    echo "Build artifacts are in: ${HOST_RESULTS_DIR}"
}

trap cleanup EXIT

docker run \
    --rm \
    --privileged \
    -v ${REPO_ROOT}:${CONTAINER_REPO_ROOT} \
    -v /var/run/docker.sock:/var/run/docker.sock \
    lorax-build:${VERSION} \
    ${CONTAINER_REPO_ROOT}/scripts/build/build_install_bundle.sh
