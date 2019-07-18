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

IMG_VERSION=1.2
TEST_REGISTRY=${TEST_REGISTRY:-pc2-dtr.dev.purestorage.com}
BUILDER_IMAGE=${BUILDER_IMAGE:-pure-go-builder:${IMG_VERSION}}
GOBUILDER=${GOBUILDER:-${TEST_REGISTRY}/purestorage/pure-go-builder:${IMG_VERSION}}

docker pull ${GOBUILDER}
docker tag ${GOBUILDER} ${BUILDER_IMAGE}
