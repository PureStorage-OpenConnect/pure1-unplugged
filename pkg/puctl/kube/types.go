// Copyright 2019, Pure Storage Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kube

// KubeVersion is the version of kubernetes we will install and assume to be working with
const KubeVersion = "v1.13.4"

// Pure1UnpluggedNamespace Kubernetes namespace for running the Pure1 Unplugged application
const Pure1UnpluggedNamespace = "pure1-unplugged"

// SystemNamespace Kubernetes namespace for kubernetes infrastructure components
const SystemNamespace = "kube-system"
