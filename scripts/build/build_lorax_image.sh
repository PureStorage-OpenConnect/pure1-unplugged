#!/usr/bin/env bash

set -ex

SCRIPT_DIR=$(dirname "$0")
REPO_ROOT=${SCRIPT_DIR}/../..
VERSION=$(${REPO_ROOT}/scripts/generate_version.sh ${REPO_ROOT})


${SCRIPT_DIR}/../print_console_label.sh "Building 'lorax-build' Docker Image"

# kubernetes images from kubeadm in centos container (same version we will install/use in the deployment)
KUBEVERSION=$(grep 'const KubeVersion' ${REPO_ROOT}/pkg/puctl/kube/types.go | awk -F ' = ' '{print $2}')
KUBEVERSION=$(sed -e 's/^"//' -e 's/"$//' <<<"${KUBEVERSION}")  # remove quotes
KUBEVERSION=$(sed -e 's/^v//'  <<<"${KUBEVERSION}")  # remove 'v' from 'v1.2.3' syntax
if [[ -z "${KUBEVERSION}" ]]; then
    echo "Unable to parse KUBEVERSION!"
    exit 1
fi

docker build \
    --no-cache \
    -t lorax-build:${VERSION} \
    --build-arg "KUBERNETES_VERSION=${KUBEVERSION}" \
    -f ${REPO_ROOT}/images/lorax-build/Dockerfile \
    ${REPO_ROOT}
