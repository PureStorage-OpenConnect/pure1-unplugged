#!/usr/bin/env bash
# Used to test the static content in Chrome (not headless)

set -ex
SCRIPT_DIR=$(dirname "$0")

${SCRIPT_DIR}/../print_console_label.sh "Testing Web Content in Chrome"

cd ${SCRIPT_DIR}/../../gui
node_modules/@angular/cli/bin/ng test gui --browsers Chrome --watch=false

