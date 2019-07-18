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

package db

import (
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
)

// MetadataConnection provides a unified class to access metadata information through
// any source
type MetadataConnection struct {
	Tokens resources.APITokenStorage
	DAO    resources.ArrayDatabase
}

// BulkResponse provides a basic template for anything that returns an array of objects, and is
// designed to be easy to unmarshall
type BulkResponse struct {
	Response []map[string]interface{} `json:"response"`
}
