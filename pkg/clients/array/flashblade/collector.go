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
	"fmt"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources/metrics"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/timing"
	log "github.com/sirupsen/logrus"
)

// NewCollector creates both a new array collector and its underlying array client
func NewCollector(arrayID string, displayName string, managementEndpoint string, apiToken string, metaConnection resources.ArrayMetadata) (resources.ArrayCollector, error) {
	timer := timing.NewStageTimer("flashblade.NewCollector", log.Fields{"display_name": displayName})
	defer timer.Finish()

	arrayClient, err := NewClient(displayName, managementEndpoint, apiToken)
	if err != nil {
		log.WithFields(log.Fields{
			"display_name": displayName,
		}).Error("Could not create FlashBlade Collector")
		return nil, err
	}

	collector := Collector{
		ArrayID:        arrayID,
		ArrayType:      common.FlashBlade,
		Client:         arrayClient,
		DisplayName:    displayName,
		MgmtEndpoint:   managementEndpoint,
		metaConnection: metaConnection,
	}

	log.WithFields(log.Fields{
		"array_type":   collector.ArrayType,
		"display_name": collector.DisplayName,
	}).Info("Successfully created FlashBlade Collector")
	return &collector, nil
}

// GetAllArrayData makes multiple underlying requests to get metric data for alerts and the array
func (collector *Collector) GetAllArrayData() (*metrics.AllArrayData, error) {
	log.WithFields(log.Fields{
		"display_name": collector.DisplayName,
	}).Trace("Getting all data")
	timer := timing.NewStageTimer("flashblade.Collector.GetAllArrayData", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	// Get the basic array info
	arrayInfo, err := collector.Client.GetArrayInfo()
	if err != nil {
		return nil, err // Can't mark the data without the array name
	}

	log.WithField("display_name", collector.DisplayName).Trace("Array info collected, starting fetch goroutines")

	// Fetch all alerts
	alertsChan := make(chan []*AlertResponse)
	go collector.fetchAllAlerts(alertsChan)

	// Fetch the array metrics
	arrayMetricsChan := make(chan ArrayMetricsResponseBundle)
	go collector.fetchArrayMetrics(arrayMetricsChan)

	// Fetch the object counts
	objectCountChan := make(chan ObjectCountResponseBundle)
	go collector.fetchObjectCounts(objectCountChan)

	// Fetch the array tags
	arrayTags, err := collector.GetArrayTags()
	if err != nil {
		log.WithError(err).WithField("display_name", collector.DisplayName).Warn("Failed to collect array tags, setting to empty map instead and proceeding")
		arrayTags = map[string]string{}
	}

	// Await all responses before proceeded to process them
	var alertResponses []*AlertResponse
	var arrayMetricsResponseBundle ArrayMetricsResponseBundle
	var objectCountResponseBundle ObjectCountResponseBundle
	for i := 0; i < 3; i++ {
		select {
		case message1 := <-alertsChan:
			alertResponses = message1
			log.WithField("info", arrayInfo).Trace("Received alerts bundle")
		case message2 := <-arrayMetricsChan:
			arrayMetricsResponseBundle = message2
			log.WithField("info", arrayInfo).Trace("Received array metrics bundle")
		case message3 := <-objectCountChan:
			objectCountResponseBundle = message3
			log.WithField("info", arrayInfo).Trace("Received object counts bundle")
		}
	}

	timer.Stage("parse_responses")
	log.WithField("display_name", collector.DisplayName).Trace("Received all metrics bundles, parsing and converting")

	// Record the current time for the metrics
	creationTime := time.Now().Unix()

	// Convert the alerts
	alerts := make([]*metrics.Alert, 0, len(alertResponses))
	for _, alert := range alertResponses {
		alerts = append(alerts, convertAlertsResponse(alert, collector.ArrayID, arrayInfo.Name, collector.DisplayName, collector.MgmtEndpoint))
	}

	// Count the open and flagged alerts
	openFlaggedAlertCount := 0
	for _, alert := range alertResponses {
		if alert.Flagged == true || alert.State == "open" {
			openFlaggedAlertCount++
		}
	}

	// Combine the array objects metric
	arrayObjectsMetric := &metrics.ArrayObjectsMetric{
		AlertMessageCount:             uint32(openFlaggedAlertCount),
		FileSystemCount:               objectCountResponseBundle.FileSystemCount,
		HostCount:                     0, // Does not apply to FlashBlade
		SnapshotCount:                 objectCountResponseBundle.SnapshotCount,
		VolumeCount:                   0, // Does not apply to FlashBlade
		VolumePendingEradicationCount: 0, // Does not apply to FlashBlade
	}

	// Convert and combine the array metrics
	arrayCapacityMetric := convertArrayCapacityMetricsResponse(arrayMetricsResponseBundle.CapacityMetricsResponse)
	arrayPerformanceMetric := convertArrayPerformanceMetricsResponse(arrayMetricsResponseBundle.PerformanceMetricsResponse)
	arrayMetric := &metrics.ArrayMetric{
		ArrayCapacityMetric:    arrayCapacityMetric,
		ArrayObjectsMetric:     arrayObjectsMetric,
		ArrayPerformanceMetric: arrayPerformanceMetric,
		ArrayID:                collector.ArrayID,
		ArrayName:              arrayInfo.Name,
		ArrayType:              collector.ArrayType,
		CreatedAt:              creationTime,
		DisplayName:            collector.DisplayName,
		Tags:                   arrayTags,
	}

	// Combine all metrics together
	allArrayData := metrics.AllArrayData{
		Alerts:      alerts,
		ArrayMetric: arrayMetric,
	}
	return &allArrayData, nil
}

// GetAllVolumeData makes multiple underlying requests to get metric data for all volumes (file systems)
func (collector *Collector) GetAllVolumeData(timeWindow int64) (*metrics.AllVolumeData, error) {
	log.WithFields(log.Fields{
		"display_name": collector.DisplayName,
	}).Trace("Getting all data")
	timer := timing.NewStageTimer("flashblade.Collector.GetAllVolumeData", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	// Get the basic array info
	arrayInfo, err := collector.Client.GetArrayInfo()
	if err != nil {
		return nil, err // Can't mark the data without the array name
	}

	log.WithField("display_name", collector.DisplayName).Trace("Array info collected, beginning to fetch filesystem capacity metrics")

	// Get the various file system metrics
	timer.Stage("GetFileSystemCapacityMetrics")
	capacityResponse, err := collector.Client.GetFileSystemCapacityMetrics()
	if err != nil {
		collector.logIncompleteData(err, "GetFileSystemCapacityMetrics")
		capacityResponse = []*FileSystemCapacityMetricsResponse{}
	}

	log.WithField("display_name", collector.DisplayName).Trace("Beginning to fetch filesystem performance metrics")

	timer.Stage("GetFileSystemPerformanceMetrics")
	performanceResponse, err := collector.Client.GetFileSystemPerformanceMetrics(timeWindow)
	if err != nil {
		return nil, err // If we are missing performance data, it's too messy to tie things together
	}

	log.WithField("display_name", collector.DisplayName).Trace("Beginning to fetch filesystem snapshot metrics")

	timer.Stage("GetFileSystemSnapshots")
	snapshotsResponse, err := collector.Client.GetFileSystemSnapshots()
	if err != nil {
		collector.logIncompleteData(err, "GetFileSystemSnapshots")
		snapshotsResponse = []*FileSystemSnapshotResponse{}
	}

	// Fetch the array tags
	arrayTags, err := collector.GetArrayTags()
	if err != nil {
		log.WithError(err).WithField("display_name", collector.DisplayName).Warn("Failed to collect array tags, setting to empty map instead and proceeding")
		arrayTags = map[string]string{}
	}

	timer.Stage("parse_responses")
	log.WithField("display_name", collector.DisplayName).Trace("Fetched all metrics, parsing and converting")

	// Map the snapshots to their file systems to get a count
	volumeSnapshotCountMap := make(map[string]uint32)
	for _, snapshot := range snapshotsResponse {
		if _, ok := volumeSnapshotCountMap[snapshot.Source]; !ok {
			volumeSnapshotCountMap[snapshot.Source] = 0
		}
		volumeSnapshotCountMap[snapshot.Source]++
	}

	// Convert the capacity metrics (we want a map so we can look up a metric by name)
	var capacityMetricsMap = make(map[string]*metrics.VolumeCapacityMetric)
	for _, response := range capacityResponse {
		metric := convertVolumeCapacityMetricsResponse(response, volumeSnapshotCountMap[response.Name])
		capacityMetricsMap[response.Name] = metric
	}

	// Convert the performance metrics and combine the metrics together
	// Note: each file system may have multiple data points here, order does not matter
	var combinedVolumeMetrics []*metrics.VolumeMetric
	for _, response := range performanceResponse {
		performanceMetric := convertVolumePerformanceMetricsResponse(response)
		var capacityMetric *metrics.VolumeCapacityMetric
		if _, ok := capacityMetricsMap[response.Name]; ok {
			capacityMetric = capacityMetricsMap[response.Name]
		}
		volumeMetric := &metrics.VolumeMetric{
			VolumeCapacityMetric:    capacityMetric,
			VolumePerformanceMetric: performanceMetric,
			ArrayDisplayName:        collector.DisplayName,
			ArrayID:                 collector.ArrayID,
			ArrayName:               arrayInfo.Name,
			ArrayTags:               arrayTags,
			CreatedAt:               int64(response.Time) / 1000, // sec
			Type:                    "FileSystem",
			VolumeName:              response.Name,
		}
		combinedVolumeMetrics = append(combinedVolumeMetrics, volumeMetric)
	}

	return &metrics.AllVolumeData{
		VolumeMetricsTimeSeries: combinedVolumeMetrics,
	}, nil
}

// GetArrayID returns the ID of the array
func (collector *Collector) GetArrayID() string {
	return collector.ArrayID
}

// GetArrayModel returns the model of the array (always "FlashBlade")
func (collector *Collector) GetArrayModel() (string, error) {
	return common.FlashBlade, nil
}

// GetArrayName returns the name of the array
func (collector *Collector) GetArrayName() (string, error) {
	timer := timing.NewStageTimer("flashblade.Collector.GetArrayName", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	arrayInfo, err := collector.Client.GetArrayInfo()
	if err != nil {
		return "", err
	}
	return arrayInfo.Name, nil
}

// GetArrayTags returns the tags of the array from the API server
func (collector *Collector) GetArrayTags() (map[string]string, error) {
	timer := timing.NewStageTimer("flashblade.Collector.GetArrayTags", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	return collector.metaConnection.GetTags(collector.ArrayID)
}

// GetArrayType returns the type of array
func (collector *Collector) GetArrayType() string {
	return collector.ArrayType
}

// GetArrayVersion returns the version of the array
func (collector *Collector) GetArrayVersion() (string, error) {
	timer := timing.NewStageTimer("flashblade.Collector.GetArrayVersion", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	arrayInfo, err := collector.Client.GetArrayInfo()
	if err != nil {
		return "", err
	}

	return arrayInfo.Version, nil
}

// GetDisplayName returns the display name for the array
func (collector *Collector) GetDisplayName() string {
	return collector.DisplayName
}

// fetchAllAlerts is a helper function that makes a large request for all alerts and adds them to the channel
func (collector *Collector) fetchAllAlerts(alertsChan chan []*AlertResponse) {
	timer := timing.NewStageTimer("flashblade.Collector.fetchAllAlerts", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	log.WithField("display_name", collector.DisplayName).Trace("Beginning to fetch alerts")
	alertsResponse, err := collector.Client.GetAlerts()
	if err != nil {
		collector.logIncompleteData(err, "GetAlerts")
		alertsResponse = []*AlertResponse{}
	}

	log.WithField("display_name", collector.DisplayName).Trace("All alerts fetched, returning bundle to channel")
	alertsChan <- alertsResponse
}

// fetchArrayMetrics is a helper function that makes requests for array metrics and adds a bundled response to the channel
func (collector *Collector) fetchArrayMetrics(arrayMetricsChan chan ArrayMetricsResponseBundle) {
	timer := timing.NewStageTimer("flashblade.Collector.fetchArrayMetrics", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	log.WithField("display_name", collector.DisplayName).Trace("Beginning to fetch array capacity metrics")

	capacityResponse, err := collector.Client.GetArrayCapacityMetrics()
	if err != nil {
		collector.logIncompleteData(err, "GetArrayCapacityMetrics")
		capacityResponse = &ArrayCapacityMetricsResponse{}
	}

	log.WithField("display_name", collector.DisplayName).Trace("Beginning to fetch array performance metrics")

	performanceResponse, err := collector.Client.GetArrayPerformanceMetrics()
	if err != nil {
		collector.logIncompleteData(err, "GetArrayPerformanceMetrics")
		performanceResponse = &ArrayPerformanceMetricsResponse{}
	}

	log.WithField("display_name", collector.DisplayName).Trace("All array metrics fetched, returning bundle to channel")

	responseBundle := ArrayMetricsResponseBundle{
		CapacityMetricsResponse:    capacityResponse,
		PerformanceMetricsResponse: performanceResponse,
	}
	arrayMetricsChan <- responseBundle
}

// fetchObjectCounts is a helper function that makes requests for various object counts and adds
// a bundled response to the channel
func (collector *Collector) fetchObjectCounts(itemCountChan chan ObjectCountResponseBundle) {
	timer := timing.NewStageTimer("flashblade.Collector.fetchObjectCounts", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	log.WithField("display_name", collector.DisplayName).Trace("Beginning to fetch array filesystem count")

	fileSystemCount, err := collector.Client.GetFileSystemCount()
	if err != nil {
		collector.logIncompleteData(err, "GetFileSystemCount")
	}

	log.WithField("display_name", collector.DisplayName).Trace("Beginning to fetch array filesystem snapshot count")

	snapshotCount, err := collector.Client.GetFileSystemSnapshotCount()
	if err != nil {
		collector.logIncompleteData(err, "GetFileSystemSnapshotCount")
	}

	log.WithField("display_name", collector.DisplayName).Trace("Fetched all array object counts, returning bundle to channel")

	responseBundle := ObjectCountResponseBundle{
		FileSystemCount: fileSystemCount,
		SnapshotCount:   snapshotCount,
	}
	itemCountChan <- responseBundle
}

// logIncompleteData is a helper function to log errors when data gathering failed at some stage
func (collector *Collector) logIncompleteData(err error, subject string) {
	log.WithFields(log.Fields{
		"display_name": collector.DisplayName,
		"error":        err,
		"subject":      subject,
	}).Warn(fmt.Sprintf("Error gathering data; response will be incomplete"))
}

// ConvertAlertsResponse converts an alert response into the desired resource
func convertAlertsResponse(response *AlertResponse, arrayID string, arrayName string, arrayDisplayName string, arrayHostname string) *metrics.Alert {
	alert := &metrics.Alert{
		Action:           response.Action,
		AlertID:          response.Index,
		ArrayDisplayName: arrayDisplayName,
		ArrayHostname:    arrayHostname,
		ArrayID:          arrayID,
		ArrayName:        arrayName,
		Code:             response.Code,
		Created:          int64(response.Created) / 1000, // Convert to seconds
		Description:      response.Description,
		Flagged:          response.Flagged,
		Notified:         int64(response.Notified) / 1000, // Convert to seconds
		Severity:         response.Severity,
		State:            response.State,
		Summary:          response.Subject,
		Updated:          int64(response.Updated) / 1000, // Convert to seconds
		Variables:        response.Variables,
	}
	alert.PopulateSeverityIndex()
	return alert
}

// ConvertArrayCapacityMetricsResponse converts an array capacity metric response into the desired resource
func convertArrayCapacityMetricsResponse(response *ArrayCapacityMetricsResponse) *metrics.ArrayCapacityMetric {
	return &metrics.ArrayCapacityMetric{
		DataReduction:  response.Space.DataReduction,
		PercentFull:    float64(response.Space.TotalPhysical) / float64(response.Capacity),
		SharedSpace:    0, // Does not apply to FlashBlade
		SnapshotSpace:  response.Space.Snapshots,
		SystemSpace:    0, // Does not apply to FlashBlade
		TotalReduction: 0, // Does not apply to FlashBlade
		TotalSpace:     response.Capacity,
		UsedSpace:      response.Space.TotalPhysical,
		VolumeSpace:    response.Space.Unique,
	}
}

// ConvertArrayPerformanceMetricsResponse converts an array performance metric response into the desired resource
func convertArrayPerformanceMetricsResponse(response *ArrayPerformanceMetricsResponse) *metrics.ArrayPerformanceMetric {
	return &metrics.ArrayPerformanceMetric{
		BytesPerOp:     uint64(response.BytesPerOp),
		BytesPerRead:   uint64(response.BytesPerRead),
		BytesPerWrite:  uint64(response.BytesPerWrite),
		OtherIOPS:      uint64(response.OthersPerSec),
		OtherLatency:   uint64(response.UsecPerOtherOp),
		QueueDepth:     0, // Does not apply to FlashBlade
		ReadBandwidth:  uint64(response.OutputPerSec),
		ReadIOPS:       uint64(response.ReadsPerSec),
		ReadLatency:    uint64(response.UsecPerReadOp),
		WriteBandwidth: uint64(response.InputPerSec),
		WriteIOPS:      uint64(response.WritesPerSec),
		WriteLatency:   uint64(response.UsecPerWriteOp),
	}
}

// ConvertVolumeCapacityMetricsResponse converts a file system capacity metric response into the desired volume resource
func convertVolumeCapacityMetricsResponse(response *FileSystemCapacityMetricsResponse, snapshotCount uint32) *metrics.VolumeCapacityMetric {
	return &metrics.VolumeCapacityMetric{
		DataReduction:    response.Space.DataReduction,
		ProvisionedSpace: response.Provisioned,
		SnapshotCount:    snapshotCount,
		TotalReduction:   0, // Does not apply to FlashBlade
		UsedSpace:        response.Space.TotalPhysical,
	}
}

// ConvertVolumePerformanceMetricsResponse converts a file system performance metric response into the desired volume resource
func convertVolumePerformanceMetricsResponse(response *FileSystemPerformanceMetricsResponse) *metrics.VolumePerformanceMetric {
	return &metrics.VolumePerformanceMetric{
		ReadBandwidth:  uint64(response.ReadBytesPerSec),
		ReadIOPS:       uint64(response.ReadsPerSec),
		ReadLatency:    uint64(response.UsecPerReadOp),
		OtherIOPS:      uint64(response.OthersPerSec),
		OtherLatency:   uint64(response.UsecPerOtherOp),
		WriteBandwidth: uint64(response.WriteBytesPerSec),
		WriteIOPS:      uint64(response.WritesPerSec),
		WriteLatency:   uint64(response.UsecPerWriteOp),
	}
}
