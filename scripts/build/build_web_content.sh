#!/usr/bin/env bash
# Used to build the Angular static content
set -ex
SCRIPT_DIR=$(realpath $(dirname "$0"))
REPO_ROOT=${SCRIPT_DIR}/../..
VERSION=$(${REPO_ROOT}/scripts/generate_version.sh ${REPO_ROOT})

${REPO_ROOT}/scripts/print_console_label.sh "Building Web Content"
mkdir -p ${REPO_ROOT}/build/gui

mkdir -p ${REPO_ROOT}/gui/src/environments
cat <<EOF > ${REPO_ROOT}/gui/src/environments/environment.prod.ts
export const environment = {
  production: true,
  version: '${VERSION}'
};
EOF

pushd ${REPO_ROOT}/gui
node_modules/@angular/cli/bin/ng build gui --prod --outputPath=${REPO_ROOT}/build/gui
popd
