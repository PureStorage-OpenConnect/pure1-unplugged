#!/usr/bin/env bash

set -ex
SCRIPT_DIR=$(dirname "$0")
REPO_ROOT=${SCRIPT_DIR}/../..

${SCRIPT_DIR}/../print_console_label.sh "Building pure1-unplugged Image"

if [[ $1 = "minikube" ]]
then
    eval $(minikube docker-env)
fi

VERSION=$(${REPO_ROOT}/scripts/generate_version.sh ${REPO_ROOT})

docker build -t purestorage/pure1-unplugged:${VERSION} -f ${REPO_ROOT}/images/pure1-unplugged/Dockerfile ${REPO_ROOT}
