#!/usr/bin/env bash
# Installs the helm charts

set -e

SCRIPT_DIR=$(dirname "$0")
REPO_ROOT=${SCRIPT_DIR}/../../
VERSION=$(${REPO_ROOT}/scripts/generate_version.sh ${REPO_ROOT})

${REPO_ROOT}/scripts/print_console_label.sh "Starting helm install for pure1-unplugged"

IP_OR_FQDN=$1
if [[ -z "${IP_OR_FQDN}" ]]; then
    echo "Usage: $0 IP_OR_FQDN"
    echo "  ex: $0 \$(minikube ip)"
    exit 1
fi

${REPO_ROOT}/scripts/print_console_label.sh "Creating pure1-unplugged namespace"
# Do this before the helm install, they are expecting the certs to be installed.
# We have to create the namespace before running these script though
if [[ -z "$(kubectl get namespaces | grep pure1-unplugged)" ]]; then
    kubectl create namespace pure1-unplugged
fi

${REPO_ROOT}/scripts/print_console_label.sh "Generating SSL Certificates for Dex"

${REPO_ROOT}/scripts/print_console_label.sh "Helm init"

# Make sure helm is available
helm init
kubectl get pod -l name=tiller -n kube-system
# wait for tiller pod is up and running
while true; do
    sleep 5
    readyTillers=$(kubectl get rs -l name=tiller -n kube-system -o json | jq -r '.items[].status.readyReplicas')
    if [[ -n "${readyTillers}" ]]; then
        [ ${readyTillers} -gt 0 ] && break
    fi
done

${REPO_ROOT}/scripts/print_console_label.sh "Install pure1-unplugged"
echo "This might take a few minutes..."
helm \
    install \
    ${REPO_ROOT}/build/chart/*-chart.tgz \
    --name pure1-unplugged \
    --namespace pure1-unplugged \
    --set "global.publicAddress=${IP_OR_FQDN}" \
    --wait

echo "Finished Deploying!"
echo ""
echo "Dashboard should be available at: https://${IP_OR_FQDN}/"
echo ""
