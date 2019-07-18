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

package mock

import "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"

// Type guard: ensure this implements the interface
var _ resources.ArrayDatabase = (*ArrayDatabaseImpl)(nil)

// FindArrays is a mocked implementation
func (a *ArrayDatabaseImpl) FindArrays(query *resources.ArrayQuery) ([]*resources.Array, error) {
	args := a.Called(query)
	return args.Get(0).([]*resources.Array), args.Error(1)
}

// PatchArray is a mocked implementation
func (a *ArrayDatabaseImpl) PatchArray(device *resources.Array) (*resources.Array, error) {
	args := a.Called(device)
	return args.Get(0).(*resources.Array), args.Error(1)
}

// PatchArrayTags is a mocked implementation
func (a *ArrayDatabaseImpl) PatchArrayTags(device *resources.Array) (*resources.Array, error) {
	args := a.Called(device)
	return args.Get(0).(*resources.Array), args.Error(1)
}

// InsertArray is a mocked implementation
func (a *ArrayDatabaseImpl) InsertArray(device *resources.Array) error {
	args := a.Called(device)
	return args.Error(0)
}

// DeleteArray is a mocked implementation
func (a *ArrayDatabaseImpl) DeleteArray(query *resources.ArrayQuery) ([]string, error) {
	args := a.Called(query)
	return args.Get(0).([]string), args.Error(1)
}
