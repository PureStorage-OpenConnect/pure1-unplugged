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

package flashblade

import (
	"net"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
)

// Client and collector

// ArrayClient is the interface for the FlashBlade client
type ArrayClient interface {
	GetAlerts() ([]*AlertResponse, error)
	GetArrayCapacityMetrics() (*ArrayCapacityMetricsResponse, error)
	GetArrayInfo() (*ArrayInfoResponse, error)
	GetArrayPerformanceMetrics() (*ArrayPerformanceMetricsResponse, error)
	GetFileSystemCapacityMetrics() ([]*FileSystemCapacityMetricsResponse, error)
	GetFileSystemCount() (uint32, error)
	GetFileSystemPerformanceMetrics(window int64) ([]*FileSystemPerformanceMetricsResponse, error)
	GetFileSystemSnapshotCount() (uint32, error)
	GetFileSystemSnapshots() ([]*FileSystemSnapshotResponse, error)
}

// Client is a FlashBlade client that handles specific REST API requests
type Client struct {
	APIToken     string
	APIVersion   string
	AuthToken    string
	DisplayName  string
	ManagementIP net.IP
}

// Collector is a FlashBlade collector that uses the client to make requests
type Collector struct {
	ArrayID        string
	ArrayType      string
	Client         ArrayClient
	DisplayName    string
	MgmtEndpoint   string
	metaConnection resources.ArrayMetadata
}

// Response bundles used by the collector to group requests

// ArrayMetricsResponseBundle is used to return all array metric responses together
type ArrayMetricsResponseBundle struct {
	CapacityMetricsResponse    *ArrayCapacityMetricsResponse
	PerformanceMetricsResponse *ArrayPerformanceMetricsResponse
}

// ObjectCountResponseBundle is used to return all object count responses together
type ObjectCountResponseBundle struct {
	FileSystemCount uint32
	SnapshotCount   uint32
}

// Responses returned by the client

// AlertGenericResponse is from /alerts
type AlertGenericResponse struct {
	Items []*AlertResponse `json:"items"`
}

// AlertResponse is a sub-object from /alerts
type AlertResponse struct {
	Action      string                 `json:"action"`
	Code        uint16                 `json:"code"`
	Component   string                 `json:"component"`
	Created     uint64                 `json:"created"`
	Description string                 `json:"description"`
	Flagged     bool                   `json:"flagged"`
	Index       uint64                 `json:"index"`
	Name        string                 `json:"name"`
	Notified    uint64                 `json:"notified"`
	Severity    string                 `json:"severity"`
	State       string                 `json:"state"`
	Subject     string                 `json:"subject"`
	Updated     uint64                 `json:"updated"`
	Variables   map[string]interface{} `json:"variables"`
}

// APIVersionResponse is from /api/api_version
type APIVersionResponse struct {
	Version []string `json:"versions"`
}

// ArrayCapacityMetricsGenericResponse is from /arrays/space
type ArrayCapacityMetricsGenericResponse struct {
	Items []*ArrayCapacityMetricsResponse `json:"items"`
}

// ArrayCapacityMetricsResponse is a sub-object from /arrays/space
type ArrayCapacityMetricsResponse struct {
	Capacity uint64             `json:"capacity"`
	Name     string             `json:"name"`
	Space    ArrayCapacitySpace `json:"space"`
	Time     uint64             `json:"time"`
}

// ArrayCapacitySpace is a sub-object of ArrayCapacityMetricsResponse
type ArrayCapacitySpace struct {
	DataReduction float64 `json:"data_reduction"`
	Snapshots     uint64  `json:"snapshots"`
	TotalPhysical uint64  `json:"total_physical"`
	Unique        uint64  `json:"unique"`
	Virtual       uint64  `json:"virtual"`
}

// ArrayInfoGenericResponse is from /arrays
type ArrayInfoGenericResponse struct {
	Items []*ArrayInfoResponse `json:"items"`
}

// ArrayInfoResponse is a sub-object from /arrays
type ArrayInfoResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ArrayPerformanceMetricsGenericResponse is from /arrays/performance
type ArrayPerformanceMetricsGenericResponse struct {
	Items []*ArrayPerformanceMetricsResponse `json:"items"`
}

// ArrayPerformanceMetricsResponse is a sub-object of /arrays/performance
type ArrayPerformanceMetricsResponse struct {
	BytesPerOp     float64 `json:"bytes_per_op"`
	BytesPerRead   float64 `json:"bytes_per_read"`
	BytesPerWrite  float64 `json:"bytes_per_write"`
	InputPerSec    float64 `json:"input_per_sec"`
	Name           string  `json:"name"`
	OthersPerSec   float64 `json:"others_per_sec"`
	OutputPerSec   float64 `json:"output_per_sec"`
	ReadsPerSec    float64 `json:"reads_per_sec"`
	Time           uint64  `json:"time"`
	UsecPerOtherOp float64 `json:"usec_per_other_op"`
	UsecPerReadOp  float64 `json:"usec_per_read_op"`
	UsecPerWriteOp float64 `json:"usec_per_write_op"`
	WritesPerSec   float64 `json:"writes_per_sec"`
}

// FileSystemCapacityMetricsGenericResponse is from /file-systems
type FileSystemCapacityMetricsGenericResponse struct {
	Items          []*FileSystemCapacityMetricsResponse `json:"items"`
	PaginationInfo PaginationResponse                   `json:"pagination_info"`
}

// FileSystemCapacityMetricsResponse is a sub-object from /file-systems
type FileSystemCapacityMetricsResponse struct {
	Name        string                  `json:"name"`
	Provisioned uint64                  `json:"provisioned"`
	Space       FileSystemCapacitySpace `json:"space"`
}

// FileSystemCapacitySpace is a sub-object of FileSystemCapacityResponse
type FileSystemCapacitySpace struct {
	DataReduction float64 `json:"data_reduction"`
	Snapshots     uint64  `json:"snapshots"`
	TotalPhysical uint64  `json:"total_physical"`
	Unique        uint64  `json:"unique"`
	Virtual       uint64  `json:"virtual"`
}

// FileSystemPerformanceMetricsGenericResponse is from /file-systems/performance
type FileSystemPerformanceMetricsGenericResponse struct {
	Items          []*FileSystemPerformanceMetricsResponse `json:"items"`
	PaginationInfo PaginationResponse                      `json:"pagination_info"`
}

// FileSystemPerformanceMetricsResponse is a sub-object of /file-systems/performance
type FileSystemPerformanceMetricsResponse struct {
	BytesPerOp       float64 `json:"bytes_per_op"`
	BytesPerRead     float64 `json:"bytes_per_read"`
	BytesPerWrite    float64 `json:"bytes_per_write"`
	Name             string  `json:"name"`
	OthersPerSec     float64 `json:"others_per_sec"`
	ReadBytesPerSec  float64 `json:"read_bytes_per_sec"`
	ReadsPerSec      float64 `json:"reads_per_sec"`
	Time             uint64  `json:"time"`
	UsecPerOtherOp   float64 `json:"usec_per_other_op"`
	UsecPerReadOp    float64 `json:"usec_per_read_op"`
	UsecPerWriteOp   float64 `json:"usec_per_write_op"`
	WriteBytesPerSec float64 `json:"write_bytes_per_sec"`
	WritesPerSec     float64 `json:"writes_per_sec"`
}

// FileSystemSnapshotGenericResponse is from /file-system-snapshots
type FileSystemSnapshotGenericResponse struct {
	Items          []*FileSystemSnapshotResponse `json:"items"`
	PaginationInfo PaginationResponse            `json:"pagination_info"`
}

// FileSystemSnapshotResponse is a sub-object of /file-system-snapshots
type FileSystemSnapshotResponse struct {
	Source string `json:"source"`
}

// PaginationResponse is a part of responses from all endpoints
type PaginationResponse struct {
	TotalItemCount    uint32 `json:"total_item_count"`
	ContinuationToken string `json:"continuation_token"`
}
