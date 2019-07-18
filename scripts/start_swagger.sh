#!/usr/bin/env bash
# Starts a swagger server
set -ex

docker rm pure1-unplugged-swagger || true
BASIC_SCRIPT_DIR="$(dirname "$0")"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)" # Gets the current absolute directory (bash wizardry)

docker run -p 80:8080 -v $SCRIPT_DIR/..:/pure1-unplugged -e SWAGGER_JSON=/pure1-unplugged/deploy/helm/pure1-unplugged/charts/swagger-server/api-server-swagger.yaml --name pure1-unplugged-swagger swaggerapi/swagger-ui
