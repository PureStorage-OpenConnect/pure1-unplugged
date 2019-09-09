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
	timer := timing.NewStageTimer("flasharray.NewCollector", log.Fields{"display_name": displayName})
	defer timer.Finish()

	arrayClient, err := NewClient(displayName, managementEndpoint, apiToken)
	if err != nil {
		log.WithFields(log.Fields{
			"display_name": displayName,
		}).Error("Could not create FlashArray Collector")
		return nil, err
	}

	collector := Collector{
		ArrayID:        arrayID,
		ArrayType:      common.FlashArray,
		Client:         arrayClient,
		DisplayName:    displayName,
		MgmtEndpoint:   managementEndpoint,
		metaConnection: metaConnection,
	}

	log.WithFields(log.Fields{
		"array_type":   collector.ArrayType,
		"display_name": collector.DisplayName,
	}).Info("Successfully created FlashArray Collector")
	return &collector, nil
}

// GetAllArrayData makes multiple underlying requests to get metric data for alerts and the array
func (collector *Collector) GetAllArrayData() (*metrics.AllArrayData, error) {
	log.WithFields(log.Fields{
		"display_name": collector.DisplayName,
	}).Trace("Getting all array data")
	timer := timing.NewStageTimer("flasharray.Collector.GetAllArrayData", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	// Get the basic array info
	arrayInfo, err := collector.Client.GetArrayInfo()
	if err != nil {
		return nil, err // Can't mark the data without the array name
	}

	// Fetch the opened, closed, and flagged alerts
	alertsChan := make(chan AlertResponseBundle)
	go collector.fetchAllAlerts(alertsChan)

	// Fetch the array metrics
	arrayMetricsChan := make(chan ArrayMetricsResponseBundle)
	go collector.fetchArrayMetrics(arrayMetricsChan)

	// Fetch the item counts
	objectCountChan := make(chan ObjectCountResponseBundle)
	go collector.fetchObjectCounts(objectCountChan)

	// Fetch the array tags
	// Note: this is NOT a fatal error if this fails. We shouldn't stop all metrics collection just because
	// we don't have a bit of metadata for it.
	arrayTags, err := collector.GetArrayTags()
	if err != nil {
		arrayTags = map[string]string{}
	}

	// Await all responses before proceeded to process them
	var alertResponseBundle AlertResponseBundle
	var arrayMetricsResponseBundle ArrayMetricsResponseBundle
	var objectCountResponseBundle ObjectCountResponseBundle
	for i := 0; i < 3; i++ {
		select {
		case message1 := <-alertsChan:
			alertResponseBundle = message1
		case message2 := <-arrayMetricsChan:
			arrayMetricsResponseBundle = message2
		case message3 := <-objectCountChan:
			objectCountResponseBundle = message3
		}
	}

	timer.Stage("parse_responses")

	// Record the current time for the metrics
	creationTime := time.Now().Unix()

	// Parse and combine the alerts
	alertsMap := make(map[string]*metrics.Alert)
	// Add the timeline alerts to the map
	for _, response := range alertResponseBundle.TimelineResponse {
		alert := convertAlertsResponse(response, collector.ArrayID, arrayInfo.ArrayName, collector.DisplayName, collector.MgmtEndpoint, false)
		alertsMap[string(alert.AlertID)] = alert
	}
	// Mark the flagged alerts as flagged and leave the rest unflagged
	for _, response := range alertResponseBundle.FlaggedResponse {
		alertsMap[string(response.ID)].Flagged = true
	}
	// Convert the map to a list of values
	var combinedAlerts []*metrics.Alert
	for _, value := range alertsMap {
		combinedAlerts = append(combinedAlerts, value)
	}

	// Combine the array objects metric
	arrayObjectsMetric := &metrics.ArrayObjectsMetric{
		AlertMessageCount:             uint32(len(alertResponseBundle.TimelineResponse)),
		FileSystemCount:               0, // Does not apply to FlashArray
		HostCount:                     objectCountResponseBundle.HostCount,
		SnapshotCount:                 objectCountResponseBundle.SnapshotCount,
		VolumeCount:                   objectCountResponseBundle.VolumeCount,
		VolumePendingEradicationCount: objectCountResponseBundle.VolumePendingEradicationCount,
	}

	// Convert and combine the array metrics
	arrayCapacityMetric := convertArrayCapacityMetricsResponse(arrayMetricsResponseBundle.CapacityMetricsResponse)
	arrayPerformanceMetric := convertArrayPerformanceMetricsResponse(arrayMetricsResponseBundle.PerformanceMetricsResponse)
	arrayMetric := &metrics.ArrayMetric{
		ArrayCapacityMetric:    arrayCapacityMetric,
		ArrayObjectsMetric:     arrayObjectsMetric,
		ArrayPerformanceMetric: arrayPerformanceMetric,
		ArrayID:                collector.ArrayID,
		ArrayName:              collector.DisplayName,
		ArrayType:              collector.ArrayType,
		CreatedAt:              creationTime,
		DisplayName:            collector.DisplayName,
		Tags:                   arrayTags,
	}

	// Combine all metrics together
	allArrayData := metrics.AllArrayData{
		Alerts:      combinedAlerts,
		ArrayMetric: arrayMetric,
	}
	return &allArrayData, nil
}

// GetAllVolumeData makes multiple underlying requests to get metric data for volumes
// Note that timeWindow is ignored for FlashArray, and only gets the latest data
func (collector *Collector) GetAllVolumeData(timeWindow int64) (*metrics.AllVolumeData, error) {
	log.WithFields(log.Fields{
		"display_name": collector.DisplayName,
	}).Trace("Getting all volume data")
	timer := timing.NewStageTimer("flasharray.Collector.GetAllVolumeData", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	// Get the basic array info
	arrayInfo, err := collector.Client.GetArrayInfo()
	if err != nil {
		return nil, err // Can't mark the data without the array name
	}

	// Fetch the the various volume metrics
	timer.Stage("GetVolumeCapacityMetrics")
	capacityResponse, err := collector.Client.GetVolumeCapacityMetrics()
	if err != nil {
		collector.logIncompleteData(err, "GetVolumeCapacityMetrics")
		capacityResponse = []*VolumeCapacityMetricsResponse{}
	}

	timer.Stage("GetVolumePerformanceMetrics")
	performanceResponse, err := collector.Client.GetVolumePerformanceMetrics()
	if err != nil {
		collector.logIncompleteData(err, "GetVolumePerformanceMetrics")
		performanceResponse = []*VolumePerformanceMetricsResponse{}
	}

	timer.Stage("GetVolumeSnapshots")
	snapshotResponse, err := collector.Client.GetVolumeSnapshots()
	if err != nil {
		collector.logIncompleteData(err, "GetVolumeSnapshots")
		snapshotResponse = []*VolumeSnapshotResponse{}
	}

	// Fetch the array tags
	arrayTags, err := collector.GetArrayTags()
	if err != nil {
		arrayTags = map[string]string{}
	}

	timer.Stage("parse_responses")

	// Record the current time for the metrics
	creationTime := time.Now().Unix()

	// Map the volume snapshots to their volumes
	volumeSnapshotCountMap := make(map[string]uint32)
	for _, snapshot := range snapshotResponse {
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
			ArrayName:               arrayInfo.ArrayName,
			ArrayTags:               arrayTags,
			CreatedAt:               creationTime,
			Type:                    "Volume",
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

// GetArrayModel returns the model of the array
func (collector *Collector) GetArrayModel() (string, error) {
	timer := timing.NewStageTimer("flasharray.Collector.GetArrayModel", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	return collector.Client.GetModel()
}

// GetArrayName returns the name of the array
func (collector *Collector) GetArrayName() (string, error) {
	timer := timing.NewStageTimer("flasharray.Collector.GetArrayName", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	arrayInfo, err := collector.Client.GetArrayInfo()
	if err != nil {
		return "", err
	}

	return arrayInfo.ArrayName, nil
}

// GetArrayTags returns the tags of the array from the API server
func (collector *Collector) GetArrayTags() (map[string]string, error) {
	timer := timing.NewStageTimer("flasharray.Collector.GetArrayTags", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	return collector.metaConnection.GetTags(collector.ArrayID)
}

// GetArrayType returns the type of array
func (collector *Collector) GetArrayType() string {
	return collector.ArrayType
}

// GetArrayVersion returns the version of the array
func (collector *Collector) GetArrayVersion() (string, error) {
	timer := timing.NewStageTimer("flasharray.Collector.GetArrayVersion", log.Fields{"display_name": collector.DisplayName})
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

// fetchAllAlerts is a helper function that makes requests for the various alert types and adds
// a bundled response to the channel
func (collector *Collector) fetchAllAlerts(alertsChan chan AlertResponseBundle) {
	timer := timing.NewStageTimer("flasharray.Collector.fetchAllAlerts", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	alertsFlaggedResponse, err := collector.Client.GetAlertsFlagged()
	if err != nil {
		collector.logIncompleteData(err, "GetAlertsFlagged")
		alertsFlaggedResponse = []*AlertResponse{}
	}

	alertsTimelineResponse, err := collector.Client.GetAlertsTimeline()
	if err != nil {
		collector.logIncompleteData(err, "GetAlertsTimeline")
		alertsTimelineResponse = []*AlertResponse{}
	}

	responseBundle := AlertResponseBundle{
		FlaggedResponse:  alertsFlaggedResponse,
		TimelineResponse: alertsTimelineResponse,
	}
	alertsChan <- responseBundle
}

// fetchArrayMetrics is a helper function that makes requests for array metrics and adds a bundled
// response to the channel
func (collector *Collector) fetchArrayMetrics(arrayMetricsChan chan ArrayMetricsResponseBundle) {
	timer := timing.NewStageTimer("flasharray.Collector.fetchArrayMetrics", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	arrayCapacityResponse, err := collector.Client.GetArrayCapacityMetrics()
	if err != nil {
		collector.logIncompleteData(err, "GetArrayCapacityMetrics")
		arrayCapacityResponse = &ArrayCapacityMetricsResponse{}
	}

	arrayPerformanceResponse, err := collector.Client.GetArrayPerformanceMetrics()
	if err != nil {
		collector.logIncompleteData(err, "GetArrayPerformanceMetrics")
		arrayPerformanceResponse = &ArrayPerformanceMetricsResponse{}
	}

	responseBundle := ArrayMetricsResponseBundle{
		CapacityMetricsResponse:    arrayCapacityResponse,
		PerformanceMetricsResponse: arrayPerformanceResponse,
	}
	arrayMetricsChan <- responseBundle
}

// fetchObjectCounts is a helper function that makes requests for various object counts and adds
// a bundled response to the channel
func (collector *Collector) fetchObjectCounts(itemCountChan chan ObjectCountResponseBundle) {
	timer := timing.NewStageTimer("flasharray.Collector.fetchObjectCounts", log.Fields{"display_name": collector.DisplayName})
	defer timer.Finish()

	hostCount, err := collector.Client.GetHostCount()
	if err != nil {
		collector.logIncompleteData(err, "GetHostCount")
	}
	snapshotCount, err := collector.Client.GetVolumeSnapshotCount()
	if err != nil {
		collector.logIncompleteData(err, "GetSnapshotCount")
	}
	volumeCount, err := collector.Client.GetVolumeCount()
	if err != nil {
		collector.logIncompleteData(err, "GetHostCount")
	}
	volumePendingEradicationCount, err := collector.Client.GetVolumePendingEradicationCount()
	if err != nil {
		collector.logIncompleteData(err, "GetVolumePendingEradicationCount")
	}

	responseBundle := ObjectCountResponseBundle{
		HostCount:                     hostCount,
		SnapshotCount:                 snapshotCount,
		VolumeCount:                   volumeCount,
		VolumePendingEradicationCount: volumePendingEradicationCount,
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

// ConvertAlertsResponse converts an alert responses into the desired resource
func convertAlertsResponse(response *AlertResponse, arrayID string, arrayName string, arrayDisplayName string, arrayHostname string, flagged bool) *metrics.Alert {
	// FlashArray time formatted in "2006-01-02T15:04:05Z"
	openedTime, err := time.Parse("2006-01-02T15:04:05Z", response.Opened)
	if err != nil {
		// Default to current time
		openedTime = time.Now()
	}
	// Set the state depending if the "closed" field is populated with a non-null value
	state := "closed"
	if response.Closed == "" {
		state = "open"
	}
	alert := &metrics.Alert{
		AlertID:          response.ID,
		ArrayHostname:    arrayHostname,
		ArrayID:          arrayID,
		ArrayName:        arrayName,
		ArrayDisplayName: arrayDisplayName,
		Code:             response.Code,
		Component:        response.ComponentName,
		Created:          openedTime.UTC().Unix(),
		Description:      response.Details,
		Flagged:          flagged,
		Severity:         response.CurrentSeverity,
		State:            state,
		Summary:          response.Event,
	}
	alert.PopulateSeverityIndex()
	return alert
}

// ConvertArrayCapacityMetricsResponse converts an array capacity metric response into the desired resource
func convertArrayCapacityMetricsResponse(response *ArrayCapacityMetricsResponse) *metrics.ArrayCapacityMetric {
	return &metrics.ArrayCapacityMetric{
		DataReduction:  response.DataReduction,
		PercentFull:    float64(response.TotalSpace) / float64(response.Capacity),
		SharedSpace:    response.SharedSpace,
		SnapshotSpace:  response.Snapshots,
		SystemSpace:    response.SystemSpace,
		TotalReduction: response.TotalReduction,
		TotalSpace:     response.Capacity,
		UsedSpace:      response.TotalSpace,
		VolumeSpace:    response.VolumeSpace,
	}
}

// ConvertArrayPerformanceMetricsResponse converts an array performance metric response into the desired resource
func convertArrayPerformanceMetricsResponse(response *ArrayPerformanceMetricsResponse) *metrics.ArrayPerformanceMetric {
	return &metrics.ArrayPerformanceMetric{
		BytesPerRead:   response.BytesPerRead,
		BytesPerWrite:  response.BytesPerWrite,
		BytesPerOp:     response.BytesPerOp,
		OtherIOPS:      0, // Does not apply to FlashArray
		OtherLatency:   0, // Does not apply to FlashArray
		QueueDepth:     response.QueueDepth,
		ReadBandwidth:  response.OutputPerSec,
		ReadIOPS:       response.ReadsPerSec,
		ReadLatency:    response.ReadLatency,
		WriteBandwidth: response.InputPerSec,
		WriteIOPS:      response.WritesPerSec,
		WriteLatency:   response.WriteLatency,
	}
}

// ConvertVolumeCapacityMetricsResponse converts a volume capacity metric response into the desired resource
func convertVolumeCapacityMetricsResponse(response *VolumeCapacityMetricsResponse, snapshotCount uint32) *metrics.VolumeCapacityMetric {
	return &metrics.VolumeCapacityMetric{
		DataReduction:  response.DataReduction,
		SnapshotCount:  snapshotCount,
		TotalReduction: response.TotalReduction,
		UsedSpace:      response.Size,
	}
}

// ConvertVolumePerformanceMetricsResponse converts a volume performance metric response into the desired resource
func convertVolumePerformanceMetricsResponse(response *VolumePerformanceMetricsResponse) *metrics.VolumePerformanceMetric {
	return &metrics.VolumePerformanceMetric{
		ReadBandwidth:  response.OutputPerSec,
		ReadIOPS:       response.ReadsPerSec,
		ReadLatency:    response.ReadLatency,
		WriteBandwidth: response.InputPerSec,
		WriteIOPS:      response.WritesPerSec,
		WriteLatency:   response.WriteLatency,
	}
}
