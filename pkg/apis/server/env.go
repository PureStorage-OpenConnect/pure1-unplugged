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

package server

import (
	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"
)

var (
	// APIServerEnv is the loaded environment variables for the API server
	APIServerEnv *apiServerEnvironmentVariables
)

// ApiServerEnvironmentVariables represent any environment variables that can be parsed
// for the monitor server
type apiServerEnvironmentVariables struct {
	ElasticHost string `env:"ELASTIC_CLIENT_HOST" envDefault:"localhost:9200"`
}

// ParseAPIServerEnvironmentVariables loads the environment variables into APIServerEnv
func ParseAPIServerEnvironmentVariables() error {
	log.WithField("config", APIServerEnv).Info("Initializing api server environment variables")
	APIServerEnv = new(apiServerEnvironmentVariables)
	err := env.Parse(APIServerEnv)
	if err != nil {
		return err
	}
	log.WithField("config", APIServerEnv).Info("Done initializing api server environment variables")
	return nil
}
