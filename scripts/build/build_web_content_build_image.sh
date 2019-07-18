#!/usr/bin/env bash
# Used to build the Angular build image

set -ex
SCRIPT_DIR=$(dirname "$0")

${SCRIPT_DIR}/../print_console_label.sh "Building Web Content Build Image"

docker build -t pure-angular-builder:1.0 ${SCRIPT_DIR}/../../images/angular-build
