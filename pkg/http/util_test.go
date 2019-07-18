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

package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeysFilledNoKeys(t *testing.T) {
	assert.NoError(t, EnsureKeysAreFilled(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}))
}

func TestKeysFilled(t *testing.T) {
	assert.NoError(t, EnsureKeysAreFilled(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}, "key1", "key2"))
}

func TestKeysFilledMissing(t *testing.T) {
	assert.Error(t, EnsureKeysAreFilled(map[string]interface{}{
		"key1": "value1",
	}, "key1", "key2"))
}

func TestKeysFilledWhitespace(t *testing.T) {
	assert.Error(t, EnsureKeysAreFilled(map[string]interface{}{
		"key1": "value1",
		"key2": "   ",
	}, "key1", "key2"))
}
