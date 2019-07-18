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

	"golang.org/x/oauth2"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// kubeSecretAPITokenStore is an implementation of the APITokenStore interface using a k8s secret
// to store all of the data. Note that this has a 1MiB size limit, but unless there's an obscenely large
// number of different people accessing this, this should hopefully never come up (if it does, it's time
// to consider using a separate product like Hashicorp Vault anyways)
type kubeSecretAPITokenStore struct {
	secretAccess typev1.SecretInterface

	mapLock         *sync.Mutex
	secretWriteChan chan bool // channel to trigger a write
	stopChan        chan bool // channel to stop the goroutine

	names  map[string]string        // name -> API token (1:1)
	tokens map[string]string        // API token -> user ID (many:1)
	users  map[string]*oauth2.Token // user ID -> oauth token (1:1)
}

type secretData struct {
	Names  map[string]string        `json:"names,omitEmpty"`  // name -> API token (1:1)
	Tokens map[string]string        `json:"tokens,omitEmpty"` // API token -> user ID (many:1)
	Users  map[string]*oauth2.Token `json:"users,omitEmpty"`  // user ID -> oauth token (1:1)
}
