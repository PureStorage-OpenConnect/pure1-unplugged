#!/usr/bin/env bash
# Used to lint the Angular static content
set -ex
SCRIPT_DIR=$(dirname "$0")

${SCRIPT_DIR}/../print_console_label.sh "Linting Web Content"

cd ${SCRIPT_DIR}/../../gui
node_modules/@angular/cli/bin/ng lint gui

