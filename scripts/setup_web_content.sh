#!/usr/bin/env bash
# Used to set up any project dependencies

set -ex
SCRIPT_DIR=$(dirname "$0")

${SCRIPT_DIR}/print_console_label.sh "Installing Web Content Dependencies"
cd ${SCRIPT_DIR}/../gui && npm ci
