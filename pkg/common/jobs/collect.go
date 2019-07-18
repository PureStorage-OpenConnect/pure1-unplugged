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

package jobs

import (
	"fmt"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/timing"
	log "github.com/sirupsen/logrus"
)

// Shared helper between all device-operating jobs to get a useful summary string
func getDeviceSummary(a *resources.ArrayRegistrationInfo) string {
	if a == nil {
		return "<nil device>"
	}

	return fmt.Sprintf("ID %s (display name %s)", a.ID, a.Name)
}

// Type guard: ensure this implements the interface
//var _ workerpool.Job = (*ArrayMetricCollectJob)(nil)

// Description gets a string description of this job
func (m *ArrayMetricCollectJob) Description() string {
	return fmt.Sprintf("Array metric collection job for array %s", getDeviceSummary(m.TargetArray))
}

// Execute cleans up the old metrics in the given database
func (m *ArrayMetricCollectJob) Execute() {
	if m.TargetArray == nil {
		log.Error("Tried to fetch metrics for nil array, stopping")
		return
	}

	arrayID := m.TargetArray.ID
	arrayName := m.TargetArray.Name

	if m.TargetDatabase == nil {
		log.WithFields(log.Fields{
			"array_id":   arrayID,
			"array_name": arrayName,
		}).Error("Tried to fetch metrics, but database was nil, stopping (nowhere to put data)")
		return
	}

	if m.TargetPool == nil {
		log.WithFields(log.Fields{
			"array_id":   arrayID,
			"array_name": arrayName,
		}).Error("Tried to fetch metrics, but worker pool was nil, stopping (nowhere to put data push jobs)")
		return
	}

	timer := timing.NewStageTimer("ArrayMetricCollectJob.Execute", log.Fields{
		"array_id":   arrayID,
		"array_name": arrayName,
	})
	defer timer.Finish()

	log.WithField("array", *m.TargetArray).Trace("Instantiating connection for array")
	connection, err := m.CollectorFactory.InitializeCollector(m.TargetArray)
	if err != nil {
		log.WithError(err).Error("Error instantiating connection for array, stopping")
		return
	}

	timer.Stage("collecting")

	arrayMetrics, err := connection.GetAllArrayData()
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"array_id":   arrayID,
			"array_name": arrayName,
		}).Error("Error collecting array metrics")
		return
	}

	// Dispatch pushing jobs
	metricPushJob := &ArrayMetricPushJob{
		Metric:         arrayMetrics.ArrayMetric,
		TargetDatabase: m.TargetDatabase,
	}
	alertPushJob := &ArrayAlertPushJob{
		Alerts:         arrayMetrics.Alerts,
		TargetDatabase: m.TargetDatabase,
	}

	m.TargetPool.Enqueue(metricPushJob, 60*time.Second)
	m.TargetPool.Enqueue(alertPushJob, 60*time.Second)
}

// Description gets a string description of this job
func (m *ArrayVolumeMetricCollectJob) Description() string {
	return fmt.Sprintf("Array volume collection job for array %s", getDeviceSummary(m.TargetArray))
}

// Execute cleans up the old metrics in the given database
func (m *ArrayVolumeMetricCollectJob) Execute() {
	if m.TargetArray == nil {
		log.Error("Tried to fetch metrics for nil array, stopping")
		return
	}

	arrayID := m.TargetArray.ID
	arrayName := m.TargetArray.Name

	if m.TargetDatabase == nil {
		log.WithFields(log.Fields{
			"array_id":   arrayID,
			"array_name": arrayName,
		}).Error("Tried to fetch metrics, but database was nil, stopping (nowhere to put data)")
		return
	}

	if m.TargetPool == nil {
		log.WithFields(log.Fields{
			"array_id":   arrayID,
			"array_name": arrayName,
		}).Error("Tried to fetch metrics, but worker pool was nil, stopping (nowhere to put data push jobs)")
		return
	}

	timer := timing.NewStageTimer("ArrayMetricCollectJob.Execute", log.Fields{
		"array_id":   arrayID,
		"array_name": arrayName,
	})
	defer timer.Finish()

	log.WithField("array", *m.TargetArray).Trace("Instantiating connection for array")
	connection, err := m.CollectorFactory.InitializeCollector(m.TargetArray)
	if err != nil {
		log.WithError(err).Error("Error instantiating connection for array, stopping")
		return
	}

	timer.Stage("collecting")

	metrics, err := connection.GetAllVolumeData(m.TimeWindow)
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"array_id":   arrayID,
			"array_name": arrayName,
		}).Error("Error collecting volume metrics")
		return
	}

	// Dispatch pushing job
	volumePushJob := &ArrayVolumeMetricPushJob{
		Metrics:        metrics.VolumeMetricsTimeSeries,
		TargetDatabase: m.TargetDatabase,
	}

	m.TargetPool.Enqueue(volumePushJob, 60*time.Second)
}
