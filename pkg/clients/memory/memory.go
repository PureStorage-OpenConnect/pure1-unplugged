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
	"fmt"
	"sync"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
)

// Type guard: check that this struct actually implements the interface
var _ resources.APITokenStorage = (*tokenStorage)(nil)

// NewInMemoryTokenStorage creates an implementation of token.Storage that uses
// in-memory maps. Therefore it is not persistent, but it's useful as a testing
// implementation.
func NewInMemoryTokenStorage() resources.APITokenStorage {
	return &tokenStorage{
		mapLock: &sync.Mutex{},
		tokens:  map[string]string{},
	}
}

// SaveToken associates the given API token with the given device ID, and returns
// if there's an error in the saving process.
func (t *tokenStorage) SaveToken(deviceID string, token string) error {
	t.mapLock.Lock()
	defer t.mapLock.Unlock()

	t.tokens[deviceID] = token
	return nil
}

// HasToken checks if there's an API token associated with the given device ID.
// An error is not thrown if the device ID doesn't exist, but an error is thrown
// if there is an issue in the checking process itself.
func (t *tokenStorage) HasToken(deviceID string) (bool, error) {
	t.mapLock.Lock()
	defer t.mapLock.Unlock()

	_, ok := t.tokens[deviceID]
	return ok, nil
}

// GetToken fetches the API token associated with the given device ID. An
// error is thrown if the key doesn't exist or there's an issue in the fetching
// process.
func (t *tokenStorage) GetToken(deviceID string) (string, error) {
	t.mapLock.Lock()
	defer t.mapLock.Unlock()

	token, ok := t.tokens[deviceID]

	if !ok {
		return "", fmt.Errorf("Token not found for device ID %s", deviceID)
	}

	return token, nil
}

// DeleteToken deletes the token with the given device ID from this storage.
// Note that a nonexistent ID should *not* be considered an error, and the
// error return value is reserved for an issue with the actual deletion process
// itself.
func (t *tokenStorage) DeleteToken(deviceID string) error {
	t.mapLock.Lock()
	defer t.mapLock.Unlock()

	delete(t.tokens, deviceID)

	return nil
}
