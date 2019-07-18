#!/usr/bin/env bash

set -ex

SCRIPT_DIR=$(dirname "$0")
REPO_ROOT=${SCRIPT_DIR}/../..
VERSION=$(${REPO_ROOT}/scripts/generate_version.sh ${REPO_ROOT})

${REPO_ROOT}/scripts/print_console_label.sh "Cleaning bundle output directory"

BUNDLE_NAME=pure1-unplugged-${VERSION}
BUNDLE_DIR=${REPO_ROOT}/build/bundle
BUNDLE_TMP_DIR=${BUNDLE_DIR}/${BUNDLE_NAME}
rm -rf ${BUNDLE_DIR}
mkdir -p ${BUNDLE_TMP_DIR}

${REPO_ROOT}/scripts/print_console_label.sh "Setup directory structure"
mkdir -p ${BUNDLE_TMP_DIR}/etc/pure1-unplugged
mkdir -p ${BUNDLE_TMP_DIR}/opt/pure1-unplugged
mkdir -p ${BUNDLE_TMP_DIR}/usr/bin

${REPO_ROOT}/scripts/print_console_label.sh "Saving deploy/ contents in bundle"
cp -r ${REPO_ROOT}/deploy/infra ${BUNDLE_TMP_DIR}/opt/pure1-unplugged/
cp ${REPO_ROOT}/build/bin/puctl ${BUNDLE_TMP_DIR}/usr/bin/puctl

# Rename the helm binary to strip off the version info
mv ${BUNDLE_TMP_DIR}/opt/pure1-unplugged/infra/helm/helm-* ${BUNDLE_TMP_DIR}/opt/pure1-unplugged/infra/helm/helm


mkdir -p ${BUNDLE_TMP_DIR}

${REPO_ROOT}/scripts/print_console_label.sh "Saving helm chart in bundle"

# Copy helm chart into bundle
cp ${REPO_ROOT}/build/chart/*-chart.tgz ${BUNDLE_TMP_DIR}/opt/pure1-unplugged/
cp ${REPO_ROOT}/build/config.yaml ${BUNDLE_TMP_DIR}/etc/pure1-unplugged/config.yaml

${REPO_ROOT}/scripts/print_console_label.sh "Generating docker image lists"

function save_docker_images() {
    local IMG_LIST=$1
    local IMG_LOCATION=$2

    mkdir -p "${IMG_LOCATION}"

    # remove duplicates
    awk '!a[$0]++' ${IMG_LIST} > ${IMG_LIST}.tmp
    rm ${IMG_LIST}
    mv ${IMG_LIST}.tmp ${IMG_LIST}

    echo "Image list:"
    cat ${IMG_LIST}

    # For each image do a 'docker save' on it, we'll use a slightly modified name for the files
    for IMG in $(cat ${IMG_LIST}); do
        local IMG_TAR=$(echo $IMG | sed 's#/#__#g' | sed 's/:/--/g').tar

        # If we don't already have the image try and pull it
        if [[ "$(docker images -q ${IMG} 2> /dev/null)" == "" ]]; then
            docker image pull ${IMG}
        fi

        docker save -o ${IMG_LOCATION}/${IMG_TAR} ${IMG}
    done
}

# First grab all our infra images and save them:
INFRA_IMG_LIST=${BUNDLE_TMP_DIR}/infra-image-list

# kubernetes images from kubeadm in centos container (same version we will install/use in the deployment)
KUBEVERSION=$(grep 'const KubeVersion' ${REPO_ROOT}/pkg/puctl/kube/types.go | awk -F ' = ' '{print $2}')
KUBEVERSION=$(sed -e 's/^"//' -e 's/"$//' <<<"${KUBEVERSION}")  # remove quotes
if [[ -z "${KUBEVERSION}" ]]; then
    echo "Unable to parse KUBEVERSION!"
    exit 1
fi
kubeadm config images list --kubernetes-version=${KUBEVERSION} >> ${INFRA_IMG_LIST}

# calico images from the calico yamls
cat ${BUNDLE_TMP_DIR}/opt/pure1-unplugged/infra/calico/*.yaml.template | grep 'image:' | awk '{print $NF}' >> ${INFRA_IMG_LIST}

# helm (tiller) image, grab the one that matches the binary we distribute
HELM_VERSION=$(${BUNDLE_TMP_DIR}/opt/pure1-unplugged/infra/helm/helm version --client --template='{{ .Client.SemVer }}')
echo "gcr.io/kubernetes-helm/tiller:${HELM_VERSION}" >> ${INFRA_IMG_LIST}


PURE1_UNPLUGGED_APP_IMAGE_LIST=${BUNDLE_TMP_DIR}/pure1-unplugged-app-image-list

# our app images from helm, this should pick up the version we injected earlier for the pure1-unplugged image too
helm template ${BUNDLE_TMP_DIR}/opt/pure1-unplugged/*-chart.tgz | grep 'image:' | awk '{print $NF}' | sed -e 's/^"//' -e 's/"$//' >> ${PURE1_UNPLUGGED_APP_IMAGE_LIST}


${REPO_ROOT}/scripts/print_console_label.sh "Saving docker images from list to bundle"

save_docker_images "${INFRA_IMG_LIST}" "${BUNDLE_TMP_DIR}/opt/pure1-unplugged/images/infra"
save_docker_images "${PURE1_UNPLUGGED_APP_IMAGE_LIST}" "${BUNDLE_TMP_DIR}/opt/pure1-unplugged/images/apps/pure1-unplugged"


${REPO_ROOT}/scripts/print_console_label.sh "Generate puctl bash autocomplete script"
mkdir -p ${BUNDLE_TMP_DIR}/etc/bash_completion.d/
${REPO_ROOT}/build/bin/puctl --config=${BUNDLE_TMP_DIR}/etc/pure1-unplugged/config.yaml completion ${BUNDLE_TMP_DIR}/etc/bash_completion.d/puctl_auto_completion

${REPO_ROOT}/scripts/print_console_label.sh "Creating compressed bundle"
BUNDLE_FILE=${BUNDLE_DIR}/${BUNDLE_NAME}.tar.gz
tar -czvf ${BUNDLE_FILE} -C ${BUNDLE_TMP_DIR} .

# cd so that the sha1sum file has the right leading path
pushd ${BUNDLE_DIR}
sha1sum -b ${BUNDLE_NAME}.tar.gz > ${BUNDLE_NAME}.tar.gz.sha1
popd
