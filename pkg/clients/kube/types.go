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

import (
	"sync"

	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

type kubeSecretDeviceTokenStorage struct {
	secretAccess typev1.SecretInterface

	mapLock *sync.Mutex

	secretWriteChan chan bool // channel to trigger a write
	stopChan        chan bool // channel to stop the goroutine

	tokens map[string]string // device ID -> API token
}
