#!/usr/bin/env bash

set -e

SCRIPT_DIR=$(dirname "$0")
REPO_ROOT=${SCRIPT_DIR}/../../
CHART_DIR=${REPO_ROOT}/build/chart
VERSION=$(${REPO_ROOT}/scripts/generate_version.sh ${REPO_ROOT})

rm -rf ${CHART_DIR}
mkdir -p ${CHART_DIR}

TMP_CHART_DIR=${CHART_DIR}/tmp
mkdir -p ${TMP_CHART_DIR}

cp -r ${REPO_ROOT}/deploy/helm/* ${TMP_CHART_DIR}/

# Inject in our current version to the helm chart we will build
sed -i "s/@PURE1_UNPLUGGED_VERSION@/${VERSION}/g" ${TMP_CHART_DIR}/pure1-unplugged/*.yaml

HELM_HOME=$(mktemp -d)

# Manually specify stable repo URL since it's moved: https://helm.sh/blog/new-location-stable-incubator-charts/
helm init --client-only --home ${HELM_HOME} --stable-repo-url "https://charts.helm.sh/stable"
helm package --home ${HELM_HOME} --destination ${CHART_DIR} ${TMP_CHART_DIR}/pure1-unplugged

# rename our output
mv ${CHART_DIR}/pure1-unplugged-${VERSION}.tgz ${CHART_DIR}/pure1-unplugged-${VERSION}-chart.tgz

# Copy over our base yaml file too
cp ${REPO_ROOT}/deploy/config.yaml ${REPO_ROOT}/build/config.yaml
