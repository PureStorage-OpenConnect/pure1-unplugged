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

package resources

import (
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources/metrics"
)

// ArrayDatabase provides an interface to Elastic (or mocked data, or something else) to store
// and access array metadata
type ArrayDatabase interface {
	FindArrays(query *ArrayQuery) ([]*Array, error)
	PatchArray(array *Array) (*Array, error)
	PatchArrayTags(array *Array) (*Array, error)
	InsertArray(array *Array) error
	DeleteArray(query *ArrayQuery) ([]string, error) // Returns list of IDs deleted
}

// APITokenStorage defines a type that can be used to save array API tokens
// with their ID. This should ideally be separate from the main API
// server database, as the whole intent is for API tokens to be protected.
type APITokenStorage interface {
	// SaveToken associates the given API token with the given array ID, and returns
	// if there's an error in the saving process.
	SaveToken(arrayID string, token string) error

	// HasToken checks if there's an API token associated with the given array ID.
	// An error is not thrown if the array ID doesn't exist, but an error is thrown
	// if there is an issue in the checking process itself.
	HasToken(arrayID string) (bool, error)

	// GetToken fetches the API token associated with the given array ID. An
	// error is thrown if the key doesn't exist or there's an issue in the fetching
	// process.
	GetToken(arrayID string) (string, error)

	// DeleteToken deletes the token with the given array ID from this storage.
	// Note that a nonexistent ID should *not* be considered an error, and the
	// error return value is reserved for an issue with the actual deletion process
	// itself.
	DeleteToken(arrayID string) error
}

// ArrayCollector is an interface for a FlashArray or FlashBlade collector that uses an underlying
// client to make requests and package the responses into the desired resources
type ArrayCollector interface {
	GetAllArrayData() (*metrics.AllArrayData, error)
	GetAllVolumeData(timeWindow int64) (*metrics.AllVolumeData, error)
	GetArrayID() string
	GetArrayModel() (string, error)
	GetArrayName() (string, error)
	GetArrayTags() (map[string]string, error)
	GetArrayType() string
	GetArrayVersion() (string, error)
	GetDisplayName() string
}

// ArrayDiscovery represents a connection to fetch a list of arrays
// from an external source, whether it's a json file, a database, or
// a server. It fetches the struct of raw information which is then
// used to construct a Collector of whatever desired backend connection type
// (see CollectorFactory.InitializeCollector).
type ArrayDiscovery interface {
	GetArrays() ([]*ArrayRegistrationInfo, error)
}

// CollectorFactory represents a factory that converts from an array
// metadata struct into a backend connection. Mainly a way to allow
// passing mocks for testing easier.
type CollectorFactory interface {
	InitializeCollector(*ArrayRegistrationInfo) (ArrayCollector, error)
}

// ArrayMetadata represents a connection to modify a array or get more information
// about it from an external source, whether it's a json file, a database, or
// a server.
type ArrayMetadata interface {
	Patch(arrayID string, body *ArrayPatchInfo) error
	GetTags(arrayID string) (map[string]string, error)
}

// ArrayQuery provides a struct holding possible query parameters for a request
type ArrayQuery struct {
	Ids            []string
	Names          []string
	Versions       []string
	Models         []string
	Sort           string
	SortDescending bool // false: ascending, true: descending
	Offset         int
	Limit          int
}

// Array provides a struct for unified FlashArray/FlashBlade metadata
type Array struct {
	InternalID   string              `json:"InternalID,omitempty"`
	Name         string              `json:"Name,omitempty"`
	MgmtEndPoint string              `json:"MgmtEndpoint,omitempty"`
	APIToken     string              `json:"APIToken,omitempty"`
	Status       string              `json:"Status,omitempty"`
	Lastseen     time.Time           `json:"AsOf,omitempty"`
	Lastupdated  time.Time           `json:"LastUpdated,omitempty"`
	DeviceType   string              `json:"DeviceType,omitempty"`
	Version      string              `json:"Version,omitempty"`
	Model        string              `json:"Model,omitempty"`
	Tags         []map[string]string `json:"Tags,omitempty"`
}

// ArrayPatchInfo provides the data that is commonly patched on
// the API server
type ArrayPatchInfo struct {
	Status  string `json:"status,omitempty"`
	Model   string `json:"model,omitempty"`
	Version string `json:"version,omitempty"`
	AsOf    string `json:"_as_of,omitempty"`
}

// ArrayRegistrationInfo provides all the info needed to open a
// connection with an array.
type ArrayRegistrationInfo struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	MgmtEndpoint string `json:"mgmt_endpoint"`
	APIToken     string `json:"api_token"`
	DeviceType   string `json:"device_type"`
}
