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
	"fmt"

	"github.com/caarlos0/env"
	log "github.com/sirupsen/logrus"
)

// AuthServerEnvConf stores the loaded environment variables
var AuthServerEnvConf *envAuthServerConfig

type envAuthServerConfig struct {
	ElasticHost        string `env:"ELASTIC_HOST" envDefault:"localhost:9200"`
	ClientID           string `env:"OAUTH2_CLIENT_ID" envDefault:"pure1-unplugged"`
	ClientSecret       string `env:"OAUTH2_CLIENT_SECRET" envDefault:"ZXhhbXBsZS1hcHAtc2VjcmV0"`
	RedirectURL        string `env:"AUTH_SERVER_CALLBACK_URL" envDefault:"http://127.0.0.1:5555/callback"`
	IssuerURL          string `env:"OPENID_CONNECT_ISSUER_URL" envDefault:"http://127.0.0.1:5556/dex"`
	Listen             string `env:"AUTH_SERVER_LISTEN_AT" envDefault:"http://127.0.0.1:5555"`
	TLSCert            string `env:"AUTH_SERVER_TLS_CERT" envDefault:""`
	TLSKey             string `env:"AUTH_SERVER_TLS_KEY" envDefault:""`
	TLSDialTimeout     int    `env:"TLS_DIAL_TIMEOUT" envDefault:"30"`
	TLSTimeout         int    `env:"TLS_TIMEOUT" envDefault:"10"`
	TLSContinueTimeout int    `env:"TLS_CONTINUE_TIMEOUT" envDefault:"1"`
	Debug              bool   `env:"AUTH_SERVER_DEBUG" envDefault:"false"`
}

// ParseAuthEnv parses the environment variables for the auth server
func ParseAuthEnv() error {
	log.Debugf("Initializing auth server config, conf = '%+v'", fmt.Sprintf("%+v", AuthServerEnvConf))
	AuthServerEnvConf = new(envAuthServerConfig)
	err := env.Parse(AuthServerEnvConf)
	if err != nil {
		return err
	}
	log.Debugf("Done initializing auth server config, conf = '%+v'", fmt.Sprintf("%+v", AuthServerEnvConf))
	return nil
}
