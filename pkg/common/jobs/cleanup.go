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

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/timing"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/workerpool"

	log "github.com/sirupsen/logrus"
)

// Type guard: ensure this implements the interface
var _ workerpool.Job = (*MetricCleanupJob)(nil)

// Description gets a string description of this job
func (m *MetricCleanupJob) Description() string {
	return fmt.Sprintf("Device metrics cleanup job")
}

// Execute cleans up the old metrics in the given database
func (m *MetricCleanupJob) Execute() {
	if m.TargetDatabase == nil {
		log.Error("Tried to cleanup metrics in nil database, stopping")
		return
	}

	timer := timing.NewStageTimer("MetricCleanupJob.Execute", log.Fields{})
	defer timer.Finish()

	log.Trace("Starting to cleanup device metrics")
	err := m.TargetDatabase.CleanArrayMetrics(m.MaxAgeInDays)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error cleaning device metrics, stopping")
		return
	}
	log.Trace("Finished cleaning up device metrics, starting device volume metrics cleanup")
	timer.Stage("volume_cleanup")

	err = m.TargetDatabase.CleanVolumeMetrics(m.MaxAgeInDays)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error cleaning device volume metrics, stopping")
		return
	}
	log.Trace("Completed device metrics cleanup job")
}

// Type guard: ensure this implements the interface
var _ workerpool.Job = (*AlertCleanupJob)(nil)

// Description gets a string description of this job
func (m *AlertCleanupJob) Description() string {
	return fmt.Sprintf("Device alerts cleanup job")
}

// Execute cleans up the old alerts in the given database
func (m *AlertCleanupJob) Execute() {
	if m.TargetDatabase == nil {
		log.Error("Tried to cleanup alerts in nil database, stopping")
		return
	}

	log.Trace("Starting to cleanup device alerts")
	timer := timing.NewStageTimer("AlertCleanupJob.Execute", log.Fields{})
	defer timer.Finish()

	err := m.TargetDatabase.CleanAlerts(m.MaxAgeInDays)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error cleaning device alerts, stopping")
		return
	}
	log.Trace("Completed device alerts cleanup job")
}

// Type guard: ensure this implements the interface
var _ workerpool.Job = (*ErrorLogCleanupJob)(nil)

// Description gets a string description of this job
func (m *ErrorLogCleanupJob) Description() string {
	return fmt.Sprintf("Error log cleanup job")
}

// Execute cleans up the old alerts in the given database
func (m *ErrorLogCleanupJob) Execute() {
	if m.TargetDatabase == nil {
		log.Error("Tried to cleanup error logs in nil database, stopping")
		return
	}

	log.Trace("Starting to cleanup error log")
	timer := timing.NewStageTimer("ErrorLogCleanupJob.Execute", log.Fields{})
	defer timer.Finish()

	err := m.TargetDatabase.CleanErrorLogs(m.MaxAgeInDays)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error cleaning error log, stopping")
		return
	}
	log.Trace("Completed error log cleanup job")
}

// Type guard: ensure this implements the interface
var _ workerpool.Job = (*TimerLogCleanupJob)(nil)

// Description gets a string description of this job
func (m *TimerLogCleanupJob) Description() string {
	return fmt.Sprintf("Timer log cleanup job")
}

// Execute cleans up the old alerts in the given database
func (m *TimerLogCleanupJob) Execute() {
	if m.TargetDatabase == nil {
		log.Error("Tried to cleanup timer logs in nil database, stopping")
		return
	}

	log.Trace("Starting to cleanup timer log")
	timer := timing.NewStageTimer("TimerLogCleanupJob.Execute", log.Fields{})
	defer timer.Finish()

	err := m.TargetDatabase.CleanTimerLogs(m.MaxAgeInDays)
	if err != nil {
		log.WithFields(log.Fields{
			"err": err,
		}).Error("Error cleaning timer log, stopping")
		return
	}
	log.Trace("Completed timer log cleanup job")
}
