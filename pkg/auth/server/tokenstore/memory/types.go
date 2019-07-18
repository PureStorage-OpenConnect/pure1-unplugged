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

package memory

import (
	"sync"

	"golang.org/x/oauth2"
)

// inMemoryAPITokenStore is an implementation of the APITokenStore interface using only process memory (no persistence)
type inMemoryAPITokenStore struct {
	namesLock  *sync.Mutex
	tokensLock *sync.Mutex
	usersLock  *sync.Mutex
	names      map[string]string        // name -> API token (1:1)
	tokens     map[string]string        // API token -> user ID (many:1)
	users      map[string]*oauth2.Token // user ID -> oauth token (1:1)
}
