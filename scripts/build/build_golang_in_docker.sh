#!/usr/bin/env bash

set -ex
BASIC_SCRIPT_DIR="$(dirname "$0")"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)" # Gets the current absolute directory (bash wizardry)

BUILDER_IMG="pure-go-builder:1.2"
ROOT_GITHUB_REPO="github.com/PureStorage-OpenConnect/pure1-unplugged"
REPO_PATH="/go/src/${ROOT_GITHUB_REPO}"

${BASIC_SCRIPT_DIR}/../print_console_label.sh "Building Golang Content in Docker"

docker run --rm -v "${SCRIPT_DIR}/../../":"${REPO_PATH}/" -u $(id -u):$(id -g) -w "${REPO_PATH}/" ${BUILDER_IMG} make -f ${REPO_PATH}/Makefile-golang $@
