#!/usr/bin/env bash

# Copyright 2017, Pure Storage Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -xe
SCRIPT_DIR=$(dirname "$0")
REPO_ROOT=${SCRIPT_DIR}/../
VERSION=$(${REPO_ROOT}/scripts/generate_version.sh ${REPO_ROOT})

TEST_REGISTRY=${TEST_REGISTRY:-pc2-dtr.dev.purestorage.com}
IMG_REPO=purestorage
IMG_NAME=pure1-unplugged:${VERSION}

docker tag ${IMG_REPO}/${IMG_NAME} ${TEST_REGISTRY}/${IMG_REPO}/${IMG_NAME}
docker push ${TEST_REGISTRY}/${IMG_REPO}/${IMG_NAME}
