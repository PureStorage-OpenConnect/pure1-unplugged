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

package flasharray

import (
	"net"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
)

// Client and collector

// ArrayClient is the interface for the FlashArray client
type ArrayClient interface {
	GetAlertsClosed() ([]*AlertResponse, error)
	GetAlertsFlagged() ([]*AlertResponse, error)
	GetAlertsOpen() ([]*AlertResponse, error)
	GetArrayCapacityMetrics() (*ArrayCapacityMetricsResponse, error)
	GetArrayInfo() (*ArrayInfoResponse, error)
	GetArrayPerformanceMetrics() (*ArrayPerformanceMetricsResponse, error)
	GetHostCount() (uint32, error)
	GetModel() (string, error)
	GetVolumeCapacityMetrics() ([]*VolumeCapacityMetricsResponse, error)
	GetVolumeCount() (uint32, error)
	GetVolumePerformanceMetrics() ([]*VolumePerformanceMetricsResponse, error)
	GetVolumePendingEradicationCount() (uint32, error)
	GetVolumeSnapshotCount() (uint32, error)
	GetVolumeSnapshots() ([]*VolumeSnapshotResponse, error)
}

// Client is a FlashArray client that handles specific REST API requests
type Client struct {
	APIToken     string
	APIVersion   string
	DisplayName  string
	ManagementIP net.IP
}

// Collector is a FlashArray collector that uses the client to make requests
type Collector struct {
	ArrayID        string
	ArrayType      string
	Client         ArrayClient
	DisplayName    string
	MgmtEndpoint   string
	metaConnection resources.ArrayMetadata
}

// Response bundles used by the collector to group requests

// AlertResponseBundle is used to return all alerts responses together
type AlertResponseBundle struct {
	ClosedResponse  []*AlertResponse
	FlaggedResponse []*AlertResponse
	OpenResponse    []*AlertResponse
}

// ArrayMetricsResponseBundle is used to return all array metric responses together
type ArrayMetricsResponseBundle struct {
	CapacityMetricsResponse    *ArrayCapacityMetricsResponse
	PerformanceMetricsResponse *ArrayPerformanceMetricsResponse
}

// ObjectCountResponseBundle is used to return all object count responses together
type ObjectCountResponseBundle struct {
	HostCount                     uint32
	SnapshotCount                 uint32
	VolumeCount                   uint32
	VolumePendingEradicationCount uint32
}

// Responses returned by the client

// AlertResponse is from /message regardless of parameters
type AlertResponse struct {
	Actual          string `json:"actual"`
	Category        string `json:"category"`
	Code            uint16 `json:"code"`
	ComponentName   string `json:"component_name"`
	ComponentType   string `json:"component_type"`
	CurrentSeverity string `json:"current_severity"`
	Details         string `json:"details"`
	Event           string `json:"event"`
	Expected        string `json:"expected"`
	ID              uint64 `json:"id"`
	Opened          string `json:"opened"`
}

// APIVersionResponse is from /api/api_versions
type APIVersionResponse struct {
	Version []string `json:"version"`
}

// ArrayCapacityMetricsResponse is from /array with parameters space=true
type ArrayCapacityMetricsResponse struct {
	Capacity       uint64  `json:"capacity"`
	DataReduction  float64 `json:"data_reduction"`
	SharedSpace    uint64  `json:"shared_space"`
	Snapshots      uint64  `json:"snapshots"`
	SystemSpace    uint64  `json:"system"`
	TotalReduction float64 `json:"total_reduction"`
	TotalSpace     uint64  `json:"total"`
	VolumeSpace    uint64  `json:"volumes"`
}

// ArrayControllersResponse is from /array with parameters controllers=true
type ArrayControllersResponse struct {
	Mode  string `json:"mode"`
	Model string `json:"model"`
}

// ArrayInfoResponse is from /array with no parameters
type ArrayInfoResponse struct {
	ArrayName string `json:"array_name"`
	ID        string `json:"id"`
	Version   string `json:"version"`
}

// ArrayPerformanceMetricsResponse is from /array with parameters action=monitor, size=true
type ArrayPerformanceMetricsResponse struct {
	BytesPerRead  uint64 `json:"bytes_per_read"`
	BytesPerWrite uint64 `json:"bytes_per_write"`
	BytesPerOp    uint64 `json:"bytes_per_op"`
	InputPerSec   uint64 `json:"input_per_sec"`
	OutputPerSec  uint64 `json:"output_per_sec"`
	QueueDepth    uint16 `json:"queue_depth"`
	ReadLatency   uint64 `json:"usec_per_read_op"`
	ReadsPerSec   uint64 `json:"reads_per_sec"`
	WriteLatency  uint64 `json:"usec_per_write_op"`
	WritesPerSec  uint64 `json:"writes_per_sec"`
}

// EmptyResponse is from any endpoint where we only read the headers
type EmptyResponse struct{}

// VolumeCapacityMetricsResponse is from /volume with parameters space=true
type VolumeCapacityMetricsResponse struct {
	DataReduction  float64 `json:"data_reduction"`
	Name           string  `json:"name"`
	Size           uint64  `json:"size"`
	TotalReduction float64 `json:"total_reduction"`
}

// VolumePerformanceMetricsResponse is from /volume with parameters action=monitor
type VolumePerformanceMetricsResponse struct {
	InputPerSec  uint64 `json:"input_per_sec"`
	OutputPerSec uint64 `json:"output_per_sec"`
	Name         string `json:"name"`
	ReadLatency  uint64 `json:"usec_per_read_op"`
	ReadsPerSec  uint64 `json:"reads_per_sec"`
	WriteLatency uint64 `json:"usec_per_write_op"`
	WritesPerSec uint64 `json:"writes_per_sec"`
}

// VolumeSnapshotResponse is from /volume with parameters snap=true
type VolumeSnapshotResponse struct {
	Source string `json:"source"`
}
