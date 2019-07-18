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

package hooks

import (
	"context"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/elastic"

	"github.com/sirupsen/logrus"
)

type fireFunc func(entry *logrus.Entry, hook *ElasticHook) error

// ElasticHook is a custom logrus hook for inserting certain logs into elastic
type ElasticHook struct {
	client    *elastic.Client
	ctx       context.Context
	ctxCancel context.CancelFunc
	fire      fireFunc
	index     func() string
	levels    []logrus.Level
	source    string
}
