#!/usr/bin/env bash

if uname | grep -q 'Darwin'; then
    echo "ERROR: Unable to build on OS X! Aborting.."
    exit 1
fi

set -ex

SCRIPT_DIR=$(dirname "$0")
REPO_ROOT=${SCRIPT_DIR}/../..
VERSION=$(${REPO_ROOT}/scripts/generate_version.sh ${REPO_ROOT})

${REPO_ROOT}/scripts/print_console_label.sh "Building pure1-unplugged installer image"


# Use realpath for this so we can map it into a container
HOST_RESULTS_DIR=${REPO_ROOT}/build/iso
CONTAINER_REPO_ROOT=/pure1-unplugged

function cleanup {
    # use sudo since the directory is potentially owned by root, fix that too..
    sudo chown -R $(id -un):$(id -gn) ${HOST_RESULTS_DIR}
    echo "Finished! Build artifacts are in: ${HOST_RESULTS_DIR}"
}

trap cleanup EXIT

docker run \
    --rm \
    --privileged \
    -e CONTAINER_REPO_ROOT=${CONTAINER_REPO_ROOT} \
    -v $(realpath ${REPO_ROOT}):${CONTAINER_REPO_ROOT} \
    lorax-build:${VERSION} \
    ${CONTAINER_REPO_ROOT}/scripts/build/build_installer_iso.sh
