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

import "github.com/stretchr/testify/mock"

// ArrayDatabaseImpl provides a mocked implementation of the resources.ArrayDatabase interface for testing
type ArrayDatabaseImpl struct {
	mock.Mock
}

// APITokenStorageImpl provides a mocked implementation of the resources.APITokenStorage interface for testing
type APITokenStorageImpl struct {
	mock.Mock
}

// ArrayMetadataImpl provides a mocked implementation of the resources.ArrayMetadata interface for testing
type ArrayMetadataImpl struct {
	mock.Mock
}
