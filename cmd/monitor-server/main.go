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
	"os"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/apiserver"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/array"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/elastic"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/hooks"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/jobs"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/workerpool"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/logger"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/version"
	log "github.com/sirupsen/logrus"
)

const (
	sourceName = "monitor-server"
)

func main() {
	log.SetLevel(log.TraceLevel)
	log.SetFormatter(logger.CensorAPITokensFromFormatter(&log.TextFormatter{}))
	log.SetReportCaller(true)

	log.WithFields(log.Fields{
		"version": version.Get(),
	}).Info("Starting Monitor service")

	err := parseMonitorServerEnvironmentVariables()
	if err != nil {
		log.WithError(err).Fatal("Error loading server environment variables, exiting...")
		os.Exit(1)
		return
	}

	apiServerConn := apiserver.NewConnection("http://pure1-unplugged-api-server")
	discoveryService := apiServerConn
	metadataConn := apiServerConn
	deviceFactory := array.NewRESTFactory(apiServerConn)

	databaseService, err := elastic.InitializeClient(monitorServerEnv.ElasticHost, 0, time.Second*5)
	if err != nil {
		log.WithError(err).Fatal("Error initializing elastic client, exiting...")
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

	metricsCollectionTicker := time.NewTicker(time.Duration(monitorServerEnv.MonitorPeriod) * time.Second)

	workerPool := workerpool.CreateThreadPool(monitorServerEnv.WorkerPoolThreads, monitorServerEnv.WorkerPoolBufferLength)

	for {
		select {
		case <-metricsCollectionTicker.C:
			createMonitorJobs(discoveryService, deviceFactory, metadataConn, workerPool)
			break
		}
	}
}

func createMonitorJobs(discoveryService resources.ArrayDiscovery, deviceFactory resources.CollectorFactory, metadataConnection resources.ArrayMetadata, pool workerpool.Pool) {
	devices, err := discoveryService.GetArrays()
	if err != nil {
		log.WithError(err).Error("Error getting devices from discovery service, skipping this iteration")
		return
	}
	for _, device := range devices {
		log.WithFields(device.GetLogFields(true)).Trace("Enqueueing monitor check job for device")
		pool.Enqueue(&jobs.MonitorCheckJob{
			DeviceInfo:    device,
			DeviceFactory: deviceFactory,
			Metadata:      metadataConnection,
		}, time.Duration(monitorServerEnv.MonitorPeriod)*time.Second)
	}
}
