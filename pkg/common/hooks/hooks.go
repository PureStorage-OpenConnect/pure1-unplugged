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
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/timing"

	log "github.com/sirupsen/logrus"
)

// NewErrorLogHook crates a hook to store logs with the matching levels in elastic
func NewErrorLogHook(source string, logLevels []log.Level, client *elastic.Client) (*ElasticHook, error) {
	ctx, cancel := context.WithCancel(context.TODO())

	err := client.CreateErrorLogTemplate(ctx)
	if err != nil {
		log.WithError(err).Error("ErrorLogHook failed to verify index template")
		cancel()
		return nil, err
	}

	return &ElasticHook{
		client:    client,
		ctx:       ctx,
		ctxCancel: cancel,
		fire:      errorLogHookFireAsync,
		index:     elastic.GetErrorLogIndexName,
		levels:    logLevels,
		source:    source,
	}, nil
}

// NewStageTimerHook creates a hook to store logs from the StageTimer in elastic
func NewStageTimerHook(source string, client *elastic.Client) (*ElasticHook, error) {
	ctx, cancel := context.WithCancel(context.TODO())

	err := client.CreateTimerLogTemplate(ctx)
	if err != nil {
		log.WithError(err).Error("StageTimerHook failed to verify index template")
		cancel()
		return nil, err
	}

	return &ElasticHook{
		client:    client,
		ctx:       ctx,
		ctxCancel: cancel,
		fire:      stageTimerFireFuncAsync,
		index:     elastic.GetTimerLogIndexName,
		levels:    []log.Level{log.DebugLevel},
		source:    source,
	}, nil
}

// Cancel is required for the implementation
func (hook *ElasticHook) Cancel() {
	hook.ctxCancel()
}

// Fire is required for the implementation
func (hook *ElasticHook) Fire(entry *log.Entry) error {
	return hook.fire(entry, hook)
}

// Levels is required for the implementation
func (hook *ElasticHook) Levels() []log.Level {
	return hook.levels
}

func errorLogHookFireAsync(entry *log.Entry, hook *ElasticHook) error {
	go errorLogHookFire(entry, hook)
	return nil
}

func errorLogHookFire(entry *log.Entry, hook *ElasticHook) error {
	_, err := hook.client.Client().
		Index().
		Index(hook.index()).
		Type("_doc").
		BodyJson(processErrorLogEntry(entry, hook)).
		Do(hook.ctx)
	return err
}

func processErrorLogEntry(entry *log.Entry, hook *ElasticHook) map[string]interface{} {
	data := entry.Data
	data["frame_file"] = entry.Caller.File
	data["frame_function"] = entry.Caller.Function
	data["frame_line"] = entry.Caller.Line
	data["level"] = entry.Level.String()
	data["message"] = entry.Message
	data["source"] = hook.source
	data["timestamp"] = entry.Time.UTC().Unix()
	return data
}

func processStageTimerEntry(entry *log.Entry, hook *ElasticHook) map[string]interface{} {
	data := entry.Data
	data["timestamp"] = entry.Time.UTC().Unix()
	data["source"] = hook.source
	return data
}

func stageTimerFireFuncAsync(entry *log.Entry, hook *ElasticHook) error {
	go stageTimerFireFunc(entry, hook)
	return nil
}

func stageTimerFireFunc(entry *log.Entry, hook *ElasticHook) error {
	// Ignore all logs that are not from the stage timer
	if entry.Message != timing.DebugMessage {
		return nil
	}

	_, err := hook.client.Client().
		Index().
		Index(hook.index()).
		Type("_doc").
		BodyJson(processStageTimerEntry(entry, hook)).
		Do(hook.ctx)
	return err
}
