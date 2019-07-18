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
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources/metrics"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/workerpool"
)

// MonitorCheckJob is a job that a) attempts to connect to the given device, and b)
// updates the API server with the new status
type MonitorCheckJob struct {
	DeviceInfo    *resources.ArrayRegistrationInfo
	DeviceFactory resources.CollectorFactory
	Metadata      resources.ArrayMetadata
}

// ArrayMetricCollectJob is a Job used to fetch the metrics for a given array
// which then kicks off more jobs to push the metrics to the given database
// (Collects volume metrics only for FlashArray)
type ArrayMetricCollectJob struct {
	TargetArray      *resources.ArrayRegistrationInfo
	CollectorFactory resources.CollectorFactory
	TargetDatabase   metrics.Database
	TargetPool       *workerpool.Pool
}

// ArrayVolumeMetricCollectJob is a Job used to fetch the volume metrics for a given array
// which then kicks off another job to push the metrics to the given database
// (Used by FlashBlade only)
type ArrayVolumeMetricCollectJob struct {
	TargetArray      *resources.ArrayRegistrationInfo
	CollectorFactory resources.CollectorFactory
	TargetDatabase   metrics.Database
	TargetPool       *workerpool.Pool
	TimeWindow       int64
}

// ArrayMetricPushJob pushes the given metric to the given database
type ArrayMetricPushJob struct {
	TargetDatabase metrics.Database
	Metric         *metrics.ArrayMetric
}

// ArrayVolumeMetricPushJob pushes the given volume metrics to the given database
type ArrayVolumeMetricPushJob struct {
	TargetDatabase metrics.Database
	Metrics        []*metrics.VolumeMetric
}

// ArrayAlertPushJob pushes the given alerts to the given database
type ArrayAlertPushJob struct {
	TargetDatabase metrics.Database
	Alerts         []*metrics.Alert
}

// AlertCleanupJob is a Job used to cleanup old array alerts in the given database
type AlertCleanupJob struct {
	TargetDatabase metrics.Database
	MaxAgeInDays   int
}

// ErrorLogCleanupJob is a job used to clean up error logs in the given database
type ErrorLogCleanupJob struct {
	TargetDatabase metrics.Database
	MaxAgeInDays   int
}

// MetricCleanupJob is a Job used to cleanup old array metrics in the given database
type MetricCleanupJob struct {
	TargetDatabase metrics.Database
	MaxAgeInDays   int
}

// TimerLogCleanupJob is a job used to clean up stage timer logs in the given database
type TimerLogCleanupJob struct {
	TargetDatabase metrics.Database
	MaxAgeInDays   int
}
