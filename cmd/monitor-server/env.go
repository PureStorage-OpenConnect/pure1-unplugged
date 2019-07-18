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
	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"
)

var (
	monitorServerEnv *MonitorServerEnvironmentVariables
)

// MonitorServerEnvironmentVariables represent any environment variables that can be parsed
// for the monitor server
type MonitorServerEnvironmentVariables struct {
	ElasticHost            string `env:"ELASTIC_HOST" envDefault:"localhost:9200"`
	MonitorPeriod          int    `env:"MONITOR_PERIOD" envDefault:"15"`
	WorkerPoolThreads      int    `env:"WORKER_THREADS" envDefault:"10"` // Reasonable defaults for most workloads
	WorkerPoolBufferLength int    `env:"WORKER_BUFFER_LENGTH" envDefault:"25"`
}

func parseMonitorServerEnvironmentVariables() error {
	log.WithField("config", monitorServerEnv).Info("Initializing monitor server environment variables")
	monitorServerEnv = new(MonitorServerEnvironmentVariables)
	err := env.Parse(monitorServerEnv)
	if err != nil {
		return err
	}
	log.WithField("config", monitorServerEnv).Info("Done initializing monitor server environment variables")
	return nil
}
