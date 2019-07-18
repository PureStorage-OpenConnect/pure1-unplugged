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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/tokenstore"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/elastic"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/hooks"
	log "github.com/sirupsen/logrus"
)

const (
	sourceName = "auth-server"
)

func main() {
	log.SetLevel(log.TraceLevel)
	log.SetReportCaller(true)

	err := server.ParseAuthEnv()
	if err != nil {
		log.WithError(err).Fatal("Error loading auth server environment variables, exiting...")
		os.Exit(1)
		return
	}

	databaseService, err := elastic.InitializeClient(server.AuthServerEnvConf.ElasticHost, 0, time.Second*5)
	if err != nil {
		log.WithError(err).Fatal("Error initializing elastic client, exiting...")
		os.Exit(1)
		return
	}

	errorHook, err := hooks.NewErrorLogHook(sourceName, []log.Level{log.WarnLevel, log.ErrorLevel, log.FatalLevel}, databaseService)
	if err != nil {
		log.WithError(err).Fatal("Error creating ErrorLogHook, exiting...")
		os.Exit(1)
		return
	}

	log.AddHook(errorHook)

	// Generate HMAC secret
	bytes := make([]byte, 32)
	_, err = rand.Read(bytes)

	if err != nil {
		panic(fmt.Sprintf("Error generating HMAC secret: %v", err))
	}
	tokenstore.HmacSecret = base64.URLEncoding.EncodeToString(bytes)

	err = server.Cmd().Execute()
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Errorf("Failed to execute commands of Auth Server!")
	}
}
