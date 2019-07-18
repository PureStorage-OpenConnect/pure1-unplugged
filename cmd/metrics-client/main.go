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

package main

import (
	"context"
	"os"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/apiserver"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/array"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/elastic"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/hooks"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/jobs"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources/metrics"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/workerpool"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/logger"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/version"

	log "github.com/sirupsen/logrus"
)

const (
	sourceName = "metrics-client"
)

func main() {
	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)
	// Censor API tokens from the output
	log.SetFormatter(logger.CensorAPITokensFromFormatter(&log.TextFormatter{}))
	log.WithFields(log.Fields{
		"version": version.Get(),
	}).Info("Staring Metrics client")

	err := parseMetricsEnvironmentVariables()
	if err != nil {
		log.WithError(err).Fatal("Error parsing metrics client environment variables, exiting...")
		os.Exit(1)
		return
	}

	discoveryService := apiserver.NewConnection("http://pure1-unplugged-api-server")
	collectorFactory := array.NewRESTFactory(discoveryService)
	databaseService, err := elastic.InitializeClient(metricsClientEnvConf.Host, 0, time.Second*5)
	if err != nil {
		log.WithError(err).Fatal("Error initializing elastic connection, exiting...")
		os.Exit(1)
		return
	}

	ctx := context.Background()

	err = databaseService.CreateArrayMetricsTemplate(ctx)
	if err != nil {
		log.WithError(err).Fatal("Error initializing array metrics template")
		os.Exit(1)
		return
	}

	err = databaseService.CreateVolumeMetricsTemplate(ctx)
	if err != nil {
		log.WithError(err).Fatal("Error initializing volume metrics template")
		os.Exit(1)
		return
	}

	err = databaseService.CreateAlertsTemplate(ctx)
	if err != nil {
		log.WithError(err).Fatal("Error initializing alerts template")
		os.Exit(1)
		return
	}

	timerHook, err := hooks.NewStageTimerHook(sourceName, databaseService)
	if err != nil {
		log.WithError(err).Fatal("Error creating StageTimerHook, exiting...")
		os.Exit(1)
		return
	}
	errorHook, err := hooks.NewErrorLogHook(sourceName, []log.Level{log.WarnLevel, log.ErrorLevel, log.FatalLevel}, databaseService)
	if err != nil {
		log.WithError(err).Fatal("Error creating ErrorLogHook, exiting...")
		os.Exit(1)
		return
	}
	log.AddHook(timerHook)
	log.AddHook(errorHook)

	arrayMetricsCollectionFrequency := time.Duration(metricsClientEnvConf.ArrayMetricCollectionPeriod) * time.Second
	faVolumeMetricsCollectionFrequency := time.Duration(metricsClientEnvConf.FAVolumeMetricCollectionPeriod) * time.Second
	fbVolumeMetricsCollectionFrequency := time.Duration(metricsClientEnvConf.FBVolumeMetricCollectionPeriod) * time.Second

	arrayMetricsCollectionTicker := time.NewTicker(arrayMetricsCollectionFrequency)
	faVolumeMetricsCollectionTicker := time.NewTicker(faVolumeMetricsCollectionFrequency)
	fbVolumeMetricsCollectionTicker := time.NewTicker(fbVolumeMetricsCollectionFrequency)
	dataRetentionTicker := time.NewTicker(time.Duration(metricsClientEnvConf.MetricsRetentionCheckPeriod) * time.Hour)

	workerPool := workerpool.CreateThreadPool(metricsClientEnvConf.WorkerPoolThreads, metricsClientEnvConf.WorkerPoolBufferLength)

	for {
		select {
		case <-arrayMetricsCollectionTicker.C:
			createArrayMetricsJobs(&workerPool, discoveryService, databaseService, collectorFactory, arrayMetricsCollectionFrequency)
			break
		case <-faVolumeMetricsCollectionTicker.C:
			createVolumeMetricsJobs(&workerPool, discoveryService, databaseService, collectorFactory, faVolumeMetricsCollectionFrequency, common.FlashArray)
			break
		case <-fbVolumeMetricsCollectionTicker.C:
			createVolumeMetricsJobs(&workerPool, discoveryService, databaseService, collectorFactory, fbVolumeMetricsCollectionFrequency, common.FlashBlade)
			break
		case <-dataRetentionTicker.C:
			createDataRetentionJobs(&workerPool, databaseService)
			break
		}
	}
}

func createArrayMetricsJobs(workerPool *workerpool.Pool, discoveryService resources.ArrayDiscovery, databaseService metrics.Database, collectorFactory resources.CollectorFactory, collectionPeriod time.Duration) {
	if discoveryService == nil {
		log.Error("Discovery service is nil, stopping")
		return
	}
	if databaseService == nil {
		log.Error("Database service is nil, stopping")
		return
	}

	log.Trace("Starting to fetch arrays from discovery service")
	arrays, err := discoveryService.GetArrays()

	if err != nil {
		log.WithError(err).Error("Error fetching array list, skipping this iteration")
		return
	}
	log.WithField("arrays", arrays).Trace("Fetched array list")

	for _, arrayStruct := range arrays {
		log.WithField("array", arrayStruct).Trace("Enqueueing array metrics collect job for array")
		workerPool.Enqueue(&jobs.ArrayMetricCollectJob{TargetArray: arrayStruct, CollectorFactory: collectorFactory, TargetDatabase: databaseService, TargetPool: workerPool}, collectionPeriod)
		log.WithField("array", arrayStruct).Trace("Finished array enqueueing metrics collect job for array")
	}
	log.Trace("Array loop completed")
}

func createVolumeMetricsJobs(workerPool *workerpool.Pool, discoveryService resources.ArrayDiscovery, databaseService metrics.Database, collectorFactory resources.CollectorFactory, collectionPeriod time.Duration, arrayType string) {
	if discoveryService == nil {
		log.Error("Discovery service is nil, stopping")
		return
	}
	if databaseService == nil {
		log.Error("Database service is nil, stopping")
		return
	}

	log.Trace("Starting to fetch arrays from discovery service")
	arrays, err := discoveryService.GetArrays()

	if err != nil {
		log.WithError(err).Error("Error fetching array list, skipping this iteration")
		return
	}
	log.WithField("arrays", arrays).Trace("Fetched array list")

	for _, arrayStruct := range arrays {
		if arrayStruct.DeviceType == arrayType {
			log.WithField("array", arrayStruct).Trace("Enqueueing volume metrics collect job for array")
			workerPool.Enqueue(&jobs.ArrayVolumeMetricCollectJob{TargetArray: arrayStruct, CollectorFactory: collectorFactory, TargetDatabase: databaseService, TargetPool: workerPool, TimeWindow: int64(collectionPeriod.Seconds())}, collectionPeriod)
			log.WithField("array", arrayStruct).Trace("Finished enqueueing volume metrics collect job for array")
		}
	}
	log.Trace("Array loop completed")
}

func createDataRetentionJobs(workerPool *workerpool.Pool, databaseService metrics.Database) {
	log.Info("Beginning data retention enforcement")
	workerPool.Enqueue(&jobs.MetricCleanupJob{TargetDatabase: databaseService, MaxAgeInDays: metricsClientEnvConf.MetricsRetentionPeriod}, time.Hour) // Give it an hour to run, so it almost certainly will
	log.Trace("Metrics cleanup job enqueued, enqueueing alerts cleanup job")
	workerPool.Enqueue(&jobs.AlertCleanupJob{TargetDatabase: databaseService, MaxAgeInDays: metricsClientEnvConf.AlertsRetentionPeriod}, time.Hour) // Give it an hour to run, so it almost certainly will
	log.Trace("Alerts cleanup job enqueued")
	workerPool.Enqueue(&jobs.ErrorLogCleanupJob{TargetDatabase: databaseService, MaxAgeInDays: metricsClientEnvConf.ErrorLogRetentionPeriod}, time.Hour)
	log.Trace("Error log cleanup job enqueued")
	workerPool.Enqueue(&jobs.TimerLogCleanupJob{TargetDatabase: databaseService, MaxAgeInDays: metricsClientEnvConf.StageTimerRetentionPeriod}, time.Hour)
	log.Trace("Stage timer log cleanup job enqueued")
}
