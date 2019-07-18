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
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources/metrics"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/timing"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/workerpool"
	log "github.com/sirupsen/logrus"
)

// Type guards: ensure these implement the interface
var _ workerpool.Job = (*ArrayMetricPushJob)(nil)
var _ workerpool.Job = (*ArrayVolumeMetricPushJob)(nil)
var _ workerpool.Job = (*ArrayAlertPushJob)(nil)

// Description gets a string description of this job
func (a *ArrayMetricPushJob) Description() string {
	return "Array metric push job"
}

// Execute pushes the given array metric to the given database
func (a *ArrayMetricPushJob) Execute() {
	if a.Metric == nil {
		log.Error("Tried to push nil array metric, stopping")
		return
	}

	if a.TargetDatabase == nil {
		log.WithField("metric", *a.Metric).Error("Tried to push metric to nil database, stopping (nowhere to put data)")
		return
	}

	timer := timing.NewStageTimer("ArrayMetricPushJob.Execute", log.Fields{
		"array_id":   a.Metric.ArrayID,
		"array_name": a.Metric.ArrayName,
	})
	defer timer.Finish()

	log.WithFields(log.Fields{
		"array_id":   a.Metric.ArrayID,
		"array_name": a.Metric.ArrayName,
	}).Trace("Starting to push array metrics")
	err := a.TargetDatabase.AddArrayMetrics([]*metrics.ArrayMetric{a.Metric})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"array_id":   a.Metric.ArrayID,
			"array_name": a.Metric.ArrayName,
		}).Error("Error pushing array metrics to database")
		return
	}

	log.WithFields(log.Fields{
		"array_id":   a.Metric.ArrayID,
		"array_name": a.Metric.ArrayName,
	}).Trace("Successfully pushed array metrics")
}

// Description gets a string description of this job
func (a *ArrayVolumeMetricPushJob) Description() string {
	return "Array volume metric push job"
}

// Execute pushes the given volume metrics to the given database
func (a *ArrayVolumeMetricPushJob) Execute() {
	if a.Metrics == nil {
		log.Trace("Tried to push nil volume metrics array, stopping")
		return
	}

	if a.TargetDatabase == nil {
		log.WithField("metrics", a.Metrics).Error("Tried to push volume metrics to nil database, stopping (nowhere to put data)")
		return
	}

	timer := timing.NewStageTimer("ArrayVolumeMetricPushJob.Execute", log.Fields{})
	defer timer.Finish()

	log.Trace("Starting to push volume metrics")
	err := a.TargetDatabase.AddVolumeMetrics(a.Metrics)
	if err != nil {
		log.WithError(err).Error("Error pushing volume metrics to database")
		return
	}

	log.Trace("Successfully pushed volume metrics")
}

// Description gets a string description of this job
func (a *ArrayAlertPushJob) Description() string {
	return "Array alert push job"
}

// Execute pushes the given alerts to the given database
func (a *ArrayAlertPushJob) Execute() {
	if a.Alerts == nil {
		return
	}

	if a.TargetDatabase == nil {
		log.WithField("alerts", a.Alerts).Error("Tried to push alerts to nil database, stopping (nowhere to put data)")
		return
	}

	timer := timing.NewStageTimer("ArrayAlertPushJob.Execute", log.Fields{})
	defer timer.Finish()

	log.Trace("Starting to push alerts")
	err := a.TargetDatabase.UpdateAlerts(a.Alerts)
	if err != nil {
		log.WithError(err).Error("Error pushing alerts to database")
		return
	}

	log.Trace("Successfully pushed alerts")
}
