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

package elastic

import (
	"time"

	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
)

type errorReturnOnlyFunction func() error

type boolErrorReturnFunction func() (bool, error)

type stringSliceErrorReturnFunction func() ([]string, error)

// Client is a wrapper around the core Elastic Client struct which provides some
// additional methods. It also implements the metrics.Database interface and will,
// in the future, implement the API server-related resources.ArrayDatabase interface as well.
type Client struct {
	esclient *elastic.Client
	host     string
	// Set to zero for infinite, set to one or higher to limit
	maxAttempts uint
	retryTime   time.Duration

	errorLog *log.Logger
	infoLog  *log.Logger
}
