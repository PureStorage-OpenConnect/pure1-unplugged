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
	metricsClientEnvConf *EnvMetricsClientConfig
)

// EnvMetricsClientConfig represents a configuration object for managing Elastic and
// Metrics that is populated from system environment variables.
type EnvMetricsClientConfig struct {
	NameSpace                      string `env:"METRICS_CLIENT_NAME_SPACE" envDefault:"pure1-unplugged"`
	MetricsIndexPrefix             string `env:"ELASTIC_METRICS_INDEX_PREFIX" envDefault:"pure1-unplugged-metrics"`
	MetricsTypeName                string `env:"ELASTIC_METRICS_TYPE_NAME" envDefault:"metrics"`
	VolumeMetricsIndexPrefix       string `env:"ELASTIC_VOLUME_METRICS_INDEX_PREFIX" envDefault:"pure1-unplugged-volumes"`
	VolumeMetricsTypeName          string `env:"ELASTIC_VOLUME_METRICS_TYPE_NAME" envDefault:"metrics"`
	MetricsRetentionCheckPeriod    int    `env:"ELASTIC_METRICS_RETENTION_CHECK_PERIOD" envDefault:"24"`
	MetricsRetentionPeriod         int    `env:"ELASTIC_METRICS_RETENTION_PERIOD" envDefault:"31"`
	AlertsRetentionPeriod          int    `env:"ELASTIC_ALERTS_RETENTION_PERIOD" envDefault:"365"`
	ErrorLogRetentionPeriod        int    `env:"ELASTIC_ERROR_LOG_RETENTION_PERIOD" envDefault:"1"`
	StageTimerRetentionPeriod      int    `env:"ELASTIC_STAGE_TIMER_RETENTION_PERIOD" envDefault:"1"`
	AlertsIndexName                string `env:"ELASTIC_ALERT_INDEX_NAME" envDefault:"pure1-unplugged-alerts"`
	AlertsTypeName                 string `env:"ELASTIC_ALERT_TYPE_NAME" envDefault:"alerts"`
	Host                           string `env:"ELASTIC_HOST" envDefault:"localhost:9200"`
	ArrayMetricCollectionPeriod    int    `env:"ELASTIC_ARRAY_METRIC_COLLECTION_PERIOD" envDefault:"30"`
	FAVolumeMetricCollectionPeriod int    `env:"ELASTIC_FA_VOLUME_METRIC_COLLECTION_PERIOD" envDefault:"30"`
	FBVolumeMetricCollectionPeriod int    `env:"ELASTIC_FB_VOLUME_METRIC_COLLECTION_PERIOD" envDefault:"300"` // Cannot collect as frequently as FA
	WorkerPoolThreads              int    `env:"WORKER_THREADS" envDefault:"50"`                              // Reasonable defaults for most workloads
	WorkerPoolBufferLength         int    `env:"WORKER_BUFFER_LENGTH" envDefault:"200"`
}

func parseMetricsEnvironmentVariables() error {
	log.WithField("config", metricsClientEnvConf).Debug("Initializing metrics client environment variables")
	metricsClientEnvConf = new(EnvMetricsClientConfig)
	err := env.Parse(metricsClientEnvConf)
	if err != nil {
		return err
	}
	log.WithField("config", metricsClientEnvConf).Debug("Done initializing metrics client environment variables")
	return nil
}
