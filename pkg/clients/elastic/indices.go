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
	arraysIndexName         = "pure-arrays"
	arraysTimeSeriesPrefix  = "pure-arrays-metrics-"
	volumesTimeSeriesPrefix = "pure-volumes-metrics-"
	alertsIndexName         = "pure-alerts"

	arraysIndexTypeName = "_doc"
	// These below should eventually be changed to _doc: leaving as-is for now for compatibility
	arraysTimeSeriesTypeName  = "metrics"
	volumesTimeSeriesTypeName = "metrics"
	alertsIndexTypeName       = "alerts"
)

var (
	arraysTemplate = map[string]interface{}{
		"index_patterns": []string{
			arraysIndexName,
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
					"InternalID": map[string]interface{}{
						"type": "text",
					},
					"Name": map[string]interface{}{
						"type":     "text",
						"analyzer": "lowercase_analyzer",
						"fields": map[string]interface{}{
							"keyword": map[string]interface{}{
								"type":       "keyword",
								"normalizer": "lowercase_normalizer",
							},
						},
					},
					"MgmtEndpoint": map[string]interface{}{
						"type": "text",
					},
					"Status": map[string]interface{}{
						"type": "text",
					},
					"DeviceType": map[string]interface{}{
						"type": "text",
					},
					"Model": map[string]interface{}{
						"type":     "text",
						"analyzer": "lowercase_analyzer",
						"fields": map[string]interface{}{
							"keyword": map[string]interface{}{
								"type":       "keyword",
								"normalizer": "lowercase_normalizer",
							},
						},
					},
					"Version": map[string]interface{}{
						"type":     "text",
						"analyzer": "lowercase_analyzer",
						"fields": map[string]interface{}{
							"keyword": map[string]interface{}{
								"type":       "keyword",
								"normalizer": "lowercase_normalizer",
							},
						},
					},
					"AsOf": map[string]interface{}{
						"type":   "date",
						"format": "date_hour_minute_second_millis", // yyyy-MM-dd'T'HH:mm:ss.SSS
					},
					"LastUpdated": map[string]interface{}{
						"type":   "date",
						"format": "date_hour_minute_second_millis",
					},
					"Tags": map[string]interface{}{
						"type": "nested",
					},
				},
			},
		},
	}

	arraysTimeSeriesTemplate = map[string]interface{}{
		"index_patterns": []string{
			getArrayMetricsIndexWildcard(),
		},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			arraysTimeSeriesTypeName: map[string]interface{}{
				"properties": map[string]interface{}{
					"AlertMessageCount": map[string]interface{}{
						"type": "long",
					},
					"ArrayID": map[string]interface{}{
						"type": "keyword",
					},
					"ArrayName": map[string]interface{}{ // What the array thinks it's name is
						"type": "keyword",
					},
					"ArrayType": map[string]interface{}{
						"type": "keyword",
					},
					"BytesPerOp": map[string]interface{}{
						"type": "long",
					},
					"BytesPerRead": map[string]interface{}{
						"type": "long",
					},
					"BytesPerWrite": map[string]interface{}{
						"type": "long",
					},
					"CreatedAt": map[string]interface{}{
						"type":   "date",
						"format": "epoch_second",
					},
					"DataReduction": map[string]interface{}{
						"type": "double",
					},
					"DisplayName": map[string]interface{}{ // What the user entered
						"type": "keyword",
					},
					"FileSystemCount": map[string]interface{}{
						"type": "long",
					},
					"HostCount": map[string]interface{}{
						"type": "long",
					},
					"OtherIOPS": map[string]interface{}{
						"type": "long",
					},
					"OtherLatency": map[string]interface{}{
						"type": "long",
					},
					"PercentFull": map[string]interface{}{
						"type": "double",
					},
					"QueueDepth": map[string]interface{}{
						"type": "long",
					},
					"ReadBandwidth": map[string]interface{}{
						"type": "long",
					},
					"ReadIOPS": map[string]interface{}{
						"type": "long",
					},
					"ReadLatency": map[string]interface{}{
						"type": "long",
					},
					"SharedSpace": map[string]interface{}{
						"type": "long",
					},
					"SnapshotCount": map[string]interface{}{
						"type": "long",
					},
					"SnapshotSpace": map[string]interface{}{
						"type": "long",
					},
					"SystemSpace": map[string]interface{}{
						"type": "long",
					},
					"Tags": map[string]interface{}{
						"type":    "object",
						"dynamic": true,
						"enabled": true,
					},
					"TotalReduction": map[string]interface{}{
						"type": "double",
					},
					"TotalSpace": map[string]interface{}{
						"type": "long",
					},
					"UsedSpace": map[string]interface{}{
						"type": "long",
					},
					"VolumeCount": map[string]interface{}{
						"type": "long",
					},
					"VolumePendingEradicationCount": map[string]interface{}{
						"type": "long",
					},
					"VolumeSpace": map[string]interface{}{
						"type": "long",
					},
					"WriteBandwidth": map[string]interface{}{
						"type": "long",
					},
					"WriteIOPS": map[string]interface{}{
						"type": "long",
					},
					"WriteLatency": map[string]interface{}{
						"type": "long",
					},
				},
			},
		},
	}

	volumesTimeSeriesTemplate = map[string]interface{}{
		"index_patterns": []string{
			getVolumeMetricsIndexWildcard(),
		},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
		},
		"mappings": map[string]interface{}{
			volumesTimeSeriesTypeName: map[string]interface{}{
				"properties": map[string]interface{}{
					"ArrayDisplayName": map[string]interface{}{
						"type": "keyword",
					},
					"ArrayID": map[string]interface{}{
						"type": "keyword",
					},
					"ArrayName": map[string]interface{}{
						"type": "keyword",
					},
					"ArrayTags": map[string]interface{}{
						"type":    "object",
						"dynamic": true,
						"enabled": true,
					},
					"CreatedAt": map[string]interface{}{
						"type":   "date",
						"format": "epoch_second",
					},
					"DataReduction": map[string]interface{}{
						"type": "double",
					},
					"OtherIOPS": map[string]interface{}{
						"type": "double",
					},
					"OtherLatency": map[string]interface{}{
						"type": "double",
					},
					"ProvisionedSpace": map[string]interface{}{
						"type": "double",
					},
					"ReadBandwidth": map[string]interface{}{
						"type": "double",
					},
					"ReadIOPS": map[string]interface{}{
						"type": "double",
					},
					"ReadLatency": map[string]interface{}{
						"type": "double",
					},
					"SnapshotCount": map[string]interface{}{
						"type": "double",
					},
					"TotalReduction": map[string]interface{}{
						"type": "float",
					},
					"Type": map[string]interface{}{
						"type": "keyword",
					},
					"UsedSpace": map[string]interface{}{
						"type": "double",
					},
					"VolumeName": map[string]interface{}{
						"type": "keyword",
					},
					"WriteBandwidth": map[string]interface{}{
						"type": "double",
					},
					"WriteIOPS": map[string]interface{}{
						"type": "double",
					},
					"WriteLatency": map[string]interface{}{
						"type": "double",
					},
				},
			},
		},
	}

	alertsTemplate = map[string]interface{}{
		"index_patterns": []string{
			alertsIndexName,
		},
		"settings": map[string]interface{}{
			"number_of_shards":   1,
			"number_of_replicas": 0,
			"index": map[string]interface{}{
				"max_inner_result_window": 10000,
			},
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
			alertsIndexTypeName: map[string]interface{}{
				"properties": map[string]interface{}{
					"Action": map[string]interface{}{
						"type": "text",
					},
					"AlertID": map[string]interface{}{
						"type": "integer",
					},
					"ArrayID": map[string]interface{}{
						"type":       "keyword",
						"normalizer": "lowercase_normalizer",
					},
					"ArrayHostname": map[string]interface{}{
						"type":       "keyword",
						"normalizer": "lowercase_normalizer",
					},
					"ArrayDisplayName": map[string]interface{}{
						"type":       "keyword",
						"normalizer": "lowercase_normalizer",
					},
					"ArrayName": map[string]interface{}{
						"type":       "keyword",
						"normalizer": "lowercase_normalizer",
					},
					"Code": map[string]interface{}{
						"type": "integer",
					},
					"Component": map[string]interface{}{
						"type": "text",
					},
					"Created": map[string]interface{}{
						"type":   "date",
						"format": "epoch_second",
					},
					"Description": map[string]interface{}{
						"type": "text",
					},
					"Flagged": map[string]interface{}{
						"type": "boolean",
					},
					"Notified": map[string]interface{}{
						"type":   "date",
						"format": "epoch_second",
					},
					"Severity": map[string]interface{}{
						"type":      "text",
						"fielddata": true,
					},
					"SeverityIndex": map[string]interface{}{
						"type": "integer",
					},
					"State": map[string]interface{}{
						"type":       "keyword",
						"normalizer": "lowercase_normalizer",
					},
					"Summary": map[string]interface{}{
						"type": "text",
						"fields": map[string]interface{}{
							"raw": map[string]interface{}{
								"type":      "text",
								"analyzer":  "lowercase_analyzer",
								"fielddata": true,
							},
						},
					},
					"Updated": map[string]interface{}{
						"type":   "date",
						"format": "epoch_second",
					},
					"Variables": map[string]interface{}{
						"type": "object",
						// Make sure Elastic doesn't actually parse the inner variables document.
						// Reason (example use case): if two different alerts use different
						// variable types for the same key, this would error out normally
						// (since Elastic would have already made a dynamic mapping for it of the first type).
						// This sidesteps that by just never mapping it in the first place (which is fine: querying
						// by these variables is so sporadic probably nobody would ever use it and would add lots
						// of complications).
						"enabled": false,
					},
				},
			},
		},
	}
)

// CreateArrayTemplate creates the template for the array index
func (c *Client) CreateArrayTemplate(ctx context.Context) error {
	return c.createTemplate(ctx, fmt.Sprintf("%s-template", arraysIndexName), arraysTemplate)
}

// CreateArrayMetricsTemplate creates the template for the array metrics indices
func (c *Client) CreateArrayMetricsTemplate(ctx context.Context) error {
	return c.createTemplate(ctx, fmt.Sprintf("%stemplate", arraysTimeSeriesPrefix), arraysTimeSeriesTemplate)
}

// CreateVolumeMetricsTemplate creates the template for the volume metrics indices
func (c *Client) CreateVolumeMetricsTemplate(ctx context.Context) error {
	return c.createTemplate(ctx, fmt.Sprintf("%stemplate", volumesTimeSeriesPrefix), volumesTimeSeriesTemplate)
}

// CreateAlertsTemplate creates the template for the alert index
func (c *Client) CreateAlertsTemplate(ctx context.Context) error {
	return c.createTemplate(ctx, fmt.Sprintf("%s-template", alertsIndexName), alertsTemplate)
}

func (c *Client) getArrayMetricsIndices(ctx context.Context) ([]string, error) {
	return c.tryRepeatReturnStringSliceError(func() ([]string, error) {
		indices, err := c.esclient.CatIndices().Columns("index").Index(getArrayMetricsIndexWildcard()).Do(ctx)
		if err != nil {
			return nil, err
		}
		foundIndices := []string{}
		for _, index := range indices {
			foundIndices = append(foundIndices, index.Index)
		}
		return foundIndices, nil
	})
}

func (c *Client) getVolumeMetricsIndices(ctx context.Context) ([]string, error) {
	return c.tryRepeatReturnStringSliceError(func() ([]string, error) {
		indices, err := c.esclient.CatIndices().Columns("index").Index(getVolumeMetricsIndexWildcard()).Do(ctx)
		if err != nil {
			return nil, err
		}
		foundIndices := []string{}
		for _, index := range indices {
			foundIndices = append(foundIndices, index.Index)
		}
		return foundIndices, nil
	})
}

func getArrayMetricsIndexName(time time.Time) string {
	return fmt.Sprintf("%s%s", arraysTimeSeriesPrefix, time.UTC().Format("2006-01-02"))
}

func getArrayMetricsIndexWildcard() string {
	return fmt.Sprintf("%s*", arraysTimeSeriesPrefix)
}

func getVolumeMetricsIndexName(time time.Time) string {
	return fmt.Sprintf("%s%s", volumesTimeSeriesPrefix, time.UTC().Format("2006-01-02"))
}

func getVolumeMetricsIndexWildcard() string {
	return fmt.Sprintf("%s*", volumesTimeSeriesPrefix)
}

func getTimeFromArrayMetricsIndexName(indexName string) (time.Time, error) {
	// If this has the prefix, it'll go great. If it doesn't or it's malformed, the date parsing will fail
	// and return an error
	parsed, err := time.Parse("2006-01-02", strings.TrimPrefix(indexName, arraysTimeSeriesPrefix))
	if err != nil {
		return time.Now(), err
	}
	return parsed.UTC(), nil
}

func getTimeFromVolumeMetricsIndexName(indexName string) (time.Time, error) {
	// If this has the prefix, it'll go great. If it doesn't or it's malformed, the date parsing will fail
	// and return an error
	parsed, err := time.Parse("2006-01-02", strings.TrimPrefix(indexName, volumesTimeSeriesPrefix))
	if err != nil {
		return time.Now(), err
	}
	return parsed.UTC(), nil
}
