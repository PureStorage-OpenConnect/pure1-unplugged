#!/usr/bin/env bash
# Upgrade the helm charts

set -e

SCRIPT_DIR=$(dirname "$0")
REPO_ROOT=${SCRIPT_DIR}/../../
VERSION=$(${REPO_ROOT}/scripts/generate_version.sh ${REPO_ROOT})

IP_OR_FQDN=$1
if [[ -z "${IP_OR_FQDN}" ]]; then
    echo "Usage: $0 IP_OR_FQDN"
    echo "  ex: $0 \$(minikube ip)"
    exit 1
fi

${REPO_ROOT}/scripts/print_console_label.sh "Upgrading pure1-unplugged"
echo "This might take a few minutes..."
# Install into pure1-unplugged deployment and namespace, override the publicAddress
helm \
    upgrade \
    pure1-unplugged \
    ${REPO_ROOT}/build/chart/*-chart.tgz \
    --namespace pure1-unplugged \
    --set "global.publicAddress=${IP_OR_FQDN}" \
    --recreate-pods \
    --wait

echo "Finished Upgrading!"
echo ""
echo "Dashboard should be available at: https://${IP_OR_FQDN}/"
echo ""
