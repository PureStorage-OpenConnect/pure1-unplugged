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

package metrics

// Database represents a generic connection to a backend that stores metrics data
type Database interface {
	// Bulk add array metrics (usually not necessary, usually just one at a time, but just in case)
	AddArrayMetrics(metrics []*ArrayMetric) error
	// Bulk add volume metrics
	AddVolumeMetrics(metrics []*VolumeMetric) error
	// Bulk update/upsert alerts
	UpdateAlerts(alerts []*Alert) error

	// Purge old array metrics by age: metrics older than the given age in days will be deleted, metrics from before today will be marked read-only
	CleanArrayMetrics(maxAgeInDays int) error
	// Clean old volume metrics by age: metrics older than the given age in days will be deleted, metrics from before today will be marked read-only
	CleanVolumeMetrics(maxAgeInDays int) error
	// Clean alerts by age: alerts older than the given age in days (by created date) will be deleted, and steps will be taken to mitigate any performance
	// issues resulting from the method of deletion ("delete by query" in the case of Elastic should also come along with a force merge)
	CleanAlerts(maxAgeInDays int) error
	// Clean old error logs by age: logs older than the given age in days will be deleted, logs from before today will be marked read-only
	CleanErrorLogs(maxAgeInDays int) error
	// Clean old timer logs by age: logs older than the given age in days will be deleted, metrics from before today will be marked read-only
	CleanTimerLogs(maxAgeInDays int) error
}

// Alert is unified between FlashArray and FlashBlade and stores all relevant information
type Alert struct {
	AlertID          uint64 `json:"AlertID"`
	ArrayDisplayName string `json:"ArrayDisplayName"`
	ArrayHostname    string `json:"ArrayHostname"`
	ArrayID          string `json:"ArrayID"`
	ArrayName        string `json:"ArrayName"`
	Code             uint16 `json:"Code"`
	Created          int64  `json:"Created"`
	Severity         string `json:"Severity"`
	SeverityIndex    byte   `json:"SeverityIndex"`
	State            string `json:"State"`
	Summary          string `json:"Summary"`
	// Optional params depending on array type
	Action      string                 `json:"Action"`
	Component   string                 `json:"Component"`
	Description string                 `json:"Description"`
	Flagged     bool                   `json:"Flagged"`
	Notified    int64                  `json:"Notified"`
	Updated     int64                  `json:"Updated"`
	Variables   map[string]interface{} `json:"Variables"`
}

// AllArrayData represents all metrics for the array and alerts in one response
type AllArrayData struct {
	Alerts      []*Alert
	ArrayMetric *ArrayMetric
}

// AllVolumeData represents all metrics for volumes for multiple points in time in one response
type AllVolumeData struct {
	VolumeMetricsTimeSeries []*VolumeMetric
}

// ArrayMetric represents a full array metric (capacity, performance, counts, and metadata)
type ArrayMetric struct {
	*ArrayCapacityMetric
	*ArrayObjectsMetric
	*ArrayPerformanceMetric
	ArrayID     string            `json:"ArrayID"`
	ArrayName   string            `json:"ArrayName"`
	ArrayType   string            `json:"ArrayType"`
	CreatedAt   int64             `json:"CreatedAt"` // Unix seconds since epoch
	DisplayName string            `json:"DisplayName"`
	Tags        map[string]string `json:"Tags"`
}

// ArrayCapacityMetric represents all relevant capacity metrics for an array
type ArrayCapacityMetric struct {
	DataReduction  float64 `json:"DataReduction"`
	PercentFull    float64 `json:"PercentFull"`
	SharedSpace    uint64  `json:"SharedSpace"`
	SnapshotSpace  uint64  `json:"SnapshotSpace"`
	SystemSpace    uint64  `json:"SystemSpace"`
	TotalReduction float64 `json:"TotalReduction"`
	TotalSpace     uint64  `json:"TotalSpace"`
	UsedSpace      uint64  `json:"UsedSpace"`
	VolumeSpace    uint64  `json:"VolumeSpace"`
}

// ArrayPerformanceMetric represents all relevant performance metrics for an array
type ArrayPerformanceMetric struct {
	BytesPerOp     uint64 `json:"BytesPerOp"`
	BytesPerRead   uint64 `json:"BytesPerRead"`
	BytesPerWrite  uint64 `json:"BytesPerWrite"`
	OtherIOPS      uint64 `json:"OtherIOPS"`
	OtherLatency   uint64 `json:"OtherLatency"`
	QueueDepth     uint16 `json:"QueueDepth"`
	ReadBandwidth  uint64 `json:"ReadBandwidth"`
	ReadIOPS       uint64 `json:"ReadIOPS"`
	ReadLatency    uint64 `json:"ReadLatency"`
	WriteBandwidth uint64 `json:"WriteBandwidth"`
	WriteIOPS      uint64 `json:"WriteIOPS"`
	WriteLatency   uint64 `json:"WriteLatency"`
}

// ArrayObjectsMetric represents any object count information for a device
type ArrayObjectsMetric struct {
	AlertMessageCount             uint32 `json:"AlertMessageCount"` // Only open and flagged alerts
	FileSystemCount               uint32 `json:"FileSystemCount"`
	HostCount                     uint32 `json:"HostCount"`
	SnapshotCount                 uint32 `json:"SnapshotCount"`
	VolumeCount                   uint32 `json:"VolumeCount"`
	VolumePendingEradicationCount uint32 `json:"VolumePendingEradicationCount"`
}

// VolumeCapacityMetric represents all relevant capacity information for a device volume
type VolumeCapacityMetric struct {
	DataReduction    float64 `json:"DataReduction"`
	ProvisionedSpace uint64  `json:"ProvisionedSpace"`
	SnapshotCount    uint32  `json:"SnapshotCount"`
	TotalReduction   float64 `json:"TotalReduction"`
	UsedSpace        uint64  `json:"UsedSpace"`
}

// VolumePerformanceMetric represents all relevant performance information for a device volume
type VolumePerformanceMetric struct {
	ReadBandwidth  uint64 `json:"ReadBandwidth"`
	ReadIOPS       uint64 `json:"ReadIOPS"`
	ReadLatency    uint64 `json:"ReadLatency"`
	OtherIOPS      uint64 `json:"OtherIOPS"`
	OtherLatency   uint64 `json:"OtherLatency"`
	WriteBandwidth uint64 `json:"WriteBandwidth"`
	WriteIOPS      uint64 `json:"WriteIOPS"`
	WriteLatency   uint64 `json:"WriteLatency"`
}

// VolumeMetric represents a full volume metric (capacity, performance, and metadata)
type VolumeMetric struct {
	*VolumeCapacityMetric
	*VolumePerformanceMetric
	ArrayID          string            `json:"ArrayID"`
	ArrayName        string            `json:"ArrayName"`
	ArrayDisplayName string            `json:"ArrayDisplayName"`
	ArrayTags        map[string]string `json:"ArrayTags"`
	CreatedAt        int64             `json:"CreatedAt"` // Unix seconds since epoch
	Type             string            `json:"Type"`
	VolumeName       string            `json:"VolumeName"`
}
