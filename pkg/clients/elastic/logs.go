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
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	errorLogIndexPrefix = "log-error-"
	timerLogIndexPrefix = "log-timer-"
)

var (
	errorLogTemplate = map[string]interface{}{
		"index_patterns": []string{
			getErrorLogIndexWildcard(),
		},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					"lowercase_analyzer": map[string]interface{}{
						"filter":    []string{"lowercase"},
						"tokenizer": "keyword",
					},
				},
				"normalizer": map[string]interface{}{
					"lowercase_normalizer": map[string]interface{}{
						"type":        "custom",
						"char_filter": []interface{}{},
						"filter":      []string{"lowercase"},
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"_doc": map[string]interface{}{
				"properties": map[string]interface{}{
					"frame_file": map[string]interface{}{
						"type": "keyword",
					},
					"frame_function": map[string]interface{}{
						"type": "keyword",
					},
					"frame_line": map[string]interface{}{
						"type": "long",
					},
					"level": map[string]interface{}{
						"type": "keyword",
					},
					"message": map[string]interface{}{
						"type": "text",
					},
					"source": map[string]interface{}{
						"type": "keyword",
					},
					"timestamp": map[string]interface{}{
						"type":   "date",
						"format": "epoch_second",
					},
				},
			},
		},
	}

	timerLogTemplate = map[string]interface{}{
		"index_patterns": []string{
			getTimerLogIndexWildcard(),
		},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					"lowercase_analyzer": map[string]interface{}{
						"filter":    []string{"lowercase"},
						"tokenizer": "keyword",
					},
				},
				"normalizer": map[string]interface{}{
					"lowercase_normalizer": map[string]interface{}{
						"type":        "custom",
						"char_filter": []interface{}{},
						"filter":      []string{"lowercase"},
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"_doc": map[string]interface{}{
				"properties": map[string]interface{}{
					"process_name": map[string]interface{}{
						"type": "keyword",
					},
					"source": map[string]interface{}{
						"type": "keyword",
					},
					"timestamp": map[string]interface{}{
						"type":   "date",
						"format": "epoch_second",
					},
					"total_runtime": map[string]interface{}{
						"type": "long",
					},
				},
			},
		},
	}
)

// GetErrorLogIndexName returns the dynamic index name for error logs
func GetErrorLogIndexName() string {
	return fmt.Sprintf("%s%s", errorLogIndexPrefix, time.Now().Format("2006-01-02"))
}

// GetTimerLogIndexName returns then dynamic index name for timer logs
func GetTimerLogIndexName() string {
	return fmt.Sprintf("%s%s", timerLogIndexPrefix, time.Now().Format("2006-01-02"))
}

// CreateErrorLogTemplate creates the template for the error logs
func (c *Client) CreateErrorLogTemplate(ctx context.Context) error {
	return c.createTemplate(ctx, fmt.Sprintf("%stemplate", errorLogIndexPrefix), errorLogTemplate)
}

// CreateTimerLogTemplate creates the template for the error logs
func (c *Client) CreateTimerLogTemplate(ctx context.Context) error {
	return c.createTemplate(ctx, fmt.Sprintf("%stemplate", timerLogIndexPrefix), timerLogTemplate)
}

func (c *Client) getLogIndices(ctx context.Context, wildcard string) ([]string, error) {
	return c.tryRepeatReturnStringSliceError(func() ([]string, error) {
		indices, err := c.esclient.CatIndices().Columns("index").Index(wildcard).Do(ctx)
		if err != nil {
			return nil, err
		}
		var foundIndices []string
		for _, index := range indices {
			foundIndices = append(foundIndices, index.Index)
		}
		return foundIndices, nil
	})
}

func (c *Client) getErrorLogIndices(ctx context.Context) ([]string, error) {
	return c.getLogIndices(ctx, getErrorLogIndexWildcard())
}

func (c *Client) getTimerLogIndices(ctx context.Context) ([]string, error) {
	return c.getLogIndices(ctx, getTimerLogIndexWildcard())
}

func getErrorLogIndexWildcard() string {
	return fmt.Sprintf("%s*", errorLogIndexPrefix)
}

func getTimerLogIndexWildcard() string {
	return fmt.Sprintf("%s*", timerLogIndexPrefix)
}

func getTimeFromLogIndexName(indexName string, prefix string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", strings.TrimPrefix(indexName, prefix))
	if err != nil {
		return time.Now(), err
	}
	return parsed.UTC(), nil
}

func getTimeFromErrorLogIndexName(indexName string) (time.Time, error) {
	return getTimeFromLogIndexName(indexName, errorLogIndexPrefix)
}

func getTimeFromTimerLogIndexName(indexName string) (time.Time, error) {
	return getTimeFromLogIndexName(indexName, timerLogIndexPrefix)
}
