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
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources/metrics"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/timing"

	"github.com/olivere/elastic"
	log "github.com/sirupsen/logrus"
)

// Type guard: ensure this implements the interface
var _ metrics.Database = (*Client)(nil)

// AddArrayMetrics adds the given metrics to both the time-series and latest indices
func (c *Client) AddArrayMetrics(metrics []*metrics.ArrayMetric) error {
	if len(metrics) == 0 {
		log.Debug("No device metrics to push, skipping")
		return nil
	}

	indexName := getArrayMetricsIndexName(time.Now().UTC())
	ctx := context.Background()

	timer := timing.NewStageTimer("Client.AddArrayMetrics", log.Fields{})
	defer timer.Finish()

	err := c.EnsureConnected(ctx)
	if err != nil {
		return err
	}

	timer.Stage("push_metrics")

	for _, metric := range metrics {
		// Try pushing time-series data
		err = c.tryRepeatReturnErrorOnly(func() error {
			log.WithFields(log.Fields{
				// Log name and ID separately from the metric as well to make it easier to search
				"array_name": metric.DisplayName,
				"id":         metric.ArrayID,
				"metric":     metric,
			}).Trace("Beginning device time series metric push")
			_, err = c.esclient.Index().
				Index(indexName).
				Type(arraysTimeSeriesTypeName).
				BodyJson(metric).
				Do(ctx)
			if err == nil {
				log.WithFields(log.Fields{
					// Log name and ID separately from the metric as well to make it easier to search
					"array_name": metric.DisplayName,
					"id":         metric.ArrayID,
					"metric":     metric,
				}).Trace("Array time series metric push successful")
			} else {
				log.WithFields(log.Fields{
					"err": err,
					// Log name and ID separately from the metric as well to make it easier to search
					"array_name": metric.DisplayName,
					"id":         metric.ArrayID,
					"metric":     metric,
				}).Error("Error pushing array time series metrics")
			}
			return err
		})
		if err != nil {
			return err
		}
	}

	return nil
}

// AddVolumeMetrics adds the given volume metrics to both the time-series and latest indices
func (c *Client) AddVolumeMetrics(metrics []*metrics.VolumeMetric) error {
	if len(metrics) == 0 {
		log.Debug("No volume metrics to push, skipping")
		return nil
	}

	indexName := getVolumeMetricsIndexName(time.Now().UTC())
	ctx := context.Background()

	timer := timing.NewStageTimer("Client.AddVolumeMetrics", log.Fields{})
	defer timer.Finish()

	err := c.EnsureConnected(ctx)
	if err != nil {
		return err
	}

	timer.Stage("push_metrics")

	requests := []elastic.BulkableRequest{}
	arrayNameMap := map[string]struct{}{}
	arrayIDMap := map[string]struct{}{}

	for _, metric := range metrics {
		log.WithFields(log.Fields{
			"array_name":  metric.ArrayDisplayName,
			"array_id":    metric.ArrayID,
			"volume_name": metric.VolumeName,
			"metric":      metric,
		}).Trace("Adding bulk request for volume time series metric")
		requests = append(requests, elastic.NewBulkIndexRequest().Index(indexName).Type(volumesTimeSeriesTypeName).Doc(metric))
		arrayNameMap[metric.ArrayDisplayName] = struct{}{}
		arrayIDMap[metric.ArrayID] = struct{}{}
	}
	arrayNames := []string{}
	arrayIDs := []string{}
	for name := range arrayNameMap {
		arrayNames = append(arrayNames, name)
	}
	for id := range arrayIDMap {
		arrayIDs = append(arrayIDs, id)
	}
	// Try pushing time-series data
	err = c.tryRepeatReturnErrorOnly(func() error {
		log.WithFields(log.Fields{
			"array_names": arrayNames,
			"array_ids":   arrayIDs,
		}).Trace("Beginning bulk request for volume time series metrics")
		res, err := c.esclient.Bulk().Add(requests...).Do(ctx)
		if err != nil {
			log.WithFields(log.Fields{
				"err":         err,
				"array_names": arrayNames,
				"array_ids":   arrayIDs,
			}).Error("Error pushing volume time series metrics (overall error, not individual document)")
			return err
		}

		failed := res.Failed()
		for _, failure := range failed {
			log.WithFields(log.Fields{
				"type":   failure.Error.Type,
				"id":     failure.Id,
				"reason": failure.Error.Reason,
			}).Error("Volume failed to index in bulk request")
		}
		if len(failed) > 0 {
			log.WithFields(log.Fields{
				"array_names": arrayNames,
				"array_ids":   arrayIDs,
			}).Error("Not all volumes indexed successfully")
			return fmt.Errorf("Some volumes failed in bulk request")
		}

		log.WithFields(log.Fields{
			"array_names": arrayNames,
			"array_ids":   arrayIDs,
		}).Trace("Time series metrics pushed successfully")
		return err
	})

	return nil
}

// UpdateAlerts upserts the given alerts
func (c *Client) UpdateAlerts(alerts []*metrics.Alert) error {
	if len(alerts) == 0 {
		log.Debug("No alerts to update, skipping push")
		return nil
	}

	ctx := context.Background()

	timer := timing.NewStageTimer("Client.UpdateAlerts", log.Fields{})
	defer timer.Finish()

	err := c.EnsureConnected(ctx)
	if err != nil {
		return err
	}

	requests := []elastic.BulkableRequest{}
	arrayNameMap := map[string]struct{}{}
	arrayIDMap := map[string]struct{}{}

	timer.Stage("update_alerts")

	for _, alert := range alerts {
		// Any of these missing indicates something probably very wrong
		if len(alert.ArrayID) == 0 || alert.AlertID == 0 || alert.Created == int64(0) {
			log.WithField("alert", alert).Debug("Alert missing required field, skipping")
			continue
		}
		id := fmt.Sprintf("%s-alert-%d", alert.ArrayID, alert.AlertID)

		log.WithFields(log.Fields{
			"array_name":  alert.ArrayName,
			"array_id":    alert.ArrayID,
			"alert_id":    alert.AlertID,
			"alert":       alert,
			"document_id": id,
		}).Trace("Adding bulk request for alert")
		requests = append(requests, elastic.NewBulkUpdateRequest().
			Index(alertsIndexName).
			Type(alertsIndexTypeName).
			Id(id).
			Doc(alert).
			RetryOnConflict(1). // Retry once on conflict, more than that will just be unnecessary lag
			DocAsUpsert(true))
		arrayNameMap[alert.ArrayDisplayName] = struct{}{}
		arrayIDMap[alert.ArrayID] = struct{}{}
	}

	arrayNames := []string{}
	arrayIDs := []string{}
	for name := range arrayNameMap {
		arrayNames = append(arrayNames, name)
	}
	for id := range arrayIDMap {
		arrayIDs = append(arrayIDs, id)
	}

	// TODO: replace with more complex logic to only retry failed IDs
	err = c.tryRepeatReturnErrorOnly(func() error {
		log.WithFields(log.Fields{
			"array_names": arrayNames,
			"array_ids":   arrayIDs,
		}).Trace("Beginning bulk request for alerts")
		res, err := c.esclient.Bulk().Add(requests...).Do(ctx)
		if err != nil {
			log.WithFields(log.Fields{
				"array_names": arrayNames,
				"array_ids":   arrayIDs,
			}).Error("Error in bulk request for alerts (overall error, not single document)")
			return err
		}
		failed := res.Failed()
		for _, failure := range failed {
			log.WithFields(log.Fields{
				"type":   failure.Error.Type,
				"id":     failure.Id,
				"reason": failure.Error.Reason,
			}).Error("Alert failed to upsert in bulk request")
		}
		if len(failed) > 0 {
			log.WithFields(log.Fields{
				"array_names": arrayNames,
				"array_ids":   arrayIDs,
			}).Error("Not all alerts upserted successfully")
			return fmt.Errorf("Some alerts failed in bulk upsert")
		}
		log.WithFields(log.Fields{
			"array_names": arrayNames,
			"array_ids":   arrayIDs,
		}).Trace("Alerts pushed successfully")
		return nil
	})
	return err
}

// CleanArrayMetrics deletes all indices that are older than the given age in days and marks any older than today as read-only
func (c *Client) CleanArrayMetrics(maxAgeInDays int) error {
	log.WithFields(log.Fields{
		"max_age_in_days": maxAgeInDays,
	}).Trace("Beginning device metrics cleaning")

	timer := timing.NewStageTimer("Client.CleanArrayMetrics", log.Fields{})
	defer timer.Finish()

	indices, err := c.getArrayMetricsIndices(context.Background())
	if err != nil {
		log.WithError(err).Error("Error getting device metrics indices")
		return err
	}
	toDelete := []string{}
	toReadOnly := []string{}

	timer.Stage("process_index_names")

	for _, index := range indices {
		date, err := getTimeFromArrayMetricsIndexName(index)
		if err != nil {
			log.WithField("index", index).Warn("Index has invalid date format, skipping; it will be retained")
			continue
		}
		now := time.Now().UTC()
		nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		ageInHours := nowDate.Sub(date).Hours()
		// 24 hours times the max age for deletion
		if ageInHours > float64(24*maxAgeInDays) {
			log.WithFields(log.Fields{
				"index":         index,
				"age_hours":     ageInHours,
				"max_age_hours": maxAgeInDays * 24,
			}).Info("Index is past retention date, deleting")
			toDelete = append(toDelete, index)
		} else if ageInHours > 24 {
			log.WithFields(log.Fields{
				"index":         index,
				"age_hours":     ageInHours,
				"max_age_hours": maxAgeInDays * 24,
			}).Info("Index is older than 1 day, marking as read-only")
			toReadOnly = append(toReadOnly, index)
		} else {
			log.WithFields(log.Fields{
				"index":         index,
				"age_hours":     ageInHours,
				"max_age_hours": maxAgeInDays * 24,
			}).Info("Index is fresh, not touching")
		}
	}

	timer.Stage("delete_indices")

	if len(toDelete) > 0 {
		err = c.DeleteIndices(context.Background(), toDelete)
		if err != nil {
			log.WithError(err).Error("Error deleting old indices")
			return err
		}
	}

	timer.Stage("mark_indices_read_only")

	if len(toReadOnly) > 0 {
		err = c.tryRepeatReturnErrorOnly(func() error {
			_, err := c.esclient.IndexPutSettings(toReadOnly...).BodyJson(map[string]interface{}{
				"index": map[string]interface{}{
					"blocks": map[string]interface{}{
						"read_only_allow_delete": true,
					},
				},
			}).Do(context.Background())
			return err
		})
		if err != nil {
			log.WithError(err).Error("Error marking index as read-only, but continuing (non-fatal)")
		}
	}
	log.WithFields(log.Fields{
		"max_age_in_days": maxAgeInDays,
	}).Trace("Device metrics cleaning finished")
	return nil
}

// CleanVolumeMetrics deletes all volume indices that are older than the given age in days and marks any older than today as read-only
func (c *Client) CleanVolumeMetrics(maxAgeInDays int) error {
	log.WithFields(log.Fields{
		"max_age_in_days": maxAgeInDays,
	}).Trace("Beginning volume metrics cleaning")

	timer := timing.NewStageTimer("Client.CleanVolumeMetrics", log.Fields{})
	defer timer.Finish()

	indices, err := c.getVolumeMetricsIndices(context.Background())
	if err != nil {
		log.WithError(err).Error("Error getting volume metrics indices")
		return err
	}
	toDelete := []string{}
	toReadOnly := []string{}

	timer.Stage("process_index_names")

	for _, index := range indices {
		date, err := getTimeFromVolumeMetricsIndexName(index)
		if err != nil {
			log.WithField("index", index).Warn("Index has invalid date format, skipping; it will be retained")
			continue
		}
		now := time.Now().UTC()
		nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		ageInHours := nowDate.Sub(date).Hours()
		// 24 hours times the max age for deletion (plus 1 for days)
		if ageInHours > float64(24*maxAgeInDays) {
			log.WithFields(log.Fields{
				"index":         index,
				"age_hours":     ageInHours,
				"max_age_hours": maxAgeInDays * 24,
			}).Info("Index is past retention date, deleting")
			toDelete = append(toDelete, index)
		} else if ageInHours > 24 {
			log.WithFields(log.Fields{
				"index":         index,
				"age_hours":     ageInHours,
				"max_age_hours": maxAgeInDays * 24,
			}).Info("Index is older than 1 day, marking as read-only")
			toReadOnly = append(toReadOnly, index)
		} else {
			log.WithFields(log.Fields{
				"index":         index,
				"age_hours":     ageInHours,
				"max_age_hours": maxAgeInDays * 24,
			}).Info("Index is fresh, not touching")
		}
	}

	timer.Stage("delete_indices")

	if len(toDelete) > 0 {
		log.WithFields(log.Fields{
			"to_delete": toDelete,
		}).Trace("Beginning volume metrics index deletion")
		err = c.DeleteIndices(context.Background(), toDelete)
		if err != nil {
			log.WithFields(log.Fields{
				"err":       err,
				"to_delete": toDelete,
			}).Error("Error deleting old indices")
			return err
		}
		log.WithFields(log.Fields{
			"to_delete": toDelete,
		}).Trace("Volume metrics index deletion successful")
	}

	timer.Stage("mark_indices_read_only")

	if len(toReadOnly) > 0 {
		err = c.tryRepeatReturnErrorOnly(func() error {
			log.WithFields(log.Fields{
				"to_read_only": toReadOnly,
			}).Trace("Beginning volume metrics index read-only settings change")
			_, err := c.esclient.IndexPutSettings(toReadOnly...).BodyJson(map[string]interface{}{
				"index": map[string]interface{}{
					"blocks": map[string]interface{}{
						"read_only_allow_delete": true,
					},
				},
			}).Do(context.Background())
			if err != nil {
				log.WithFields(log.Fields{
					"to_read_only": toReadOnly,
				}).Error("Error setting volume metrics index read-only setting")
				return err
			}
			log.WithFields(log.Fields{
				"to_read_only": toReadOnly,
			}).Trace("Volume metrics index read-only settings change successful")

			return nil
		})
		if err != nil {
			log.WithError(err).Error("Error marking indices as read-only, but continuing (non-fatal)")
		}
	}

	log.WithFields(log.Fields{
		"max_age_in_days": maxAgeInDays,
	}).Trace("Volume metrics cleaning finished")
	return nil
}

// CleanAlerts deletes all alerts older than the given age in days
func (c *Client) CleanAlerts(maxAgeInDays int) error {
	log.WithFields(log.Fields{
		"max_age_in_days": maxAgeInDays,
		"index":           alertsIndexName,
	}).Trace("Beginning alerts cleaning")

	timer := timing.NewStageTimer("Client.CleanAlerts", log.Fields{})
	defer timer.Finish()

	ageQuery := elastic.NewRangeQuery("Created")
	ageQuery.Lt(fmt.Sprintf("now-%dd/d", maxAgeInDays))
	err := c.DeleteByQuery(context.Background(), alertsIndexName, ageQuery)
	if err != nil {
		log.WithFields(log.Fields{
			"err":   err,
			"index": alertsIndexName,
		}).Error("Error cleaning alerts")
		return err
	}
	log.WithFields(log.Fields{
		"max_age_in_days": maxAgeInDays,
		"index":           alertsIndexName,
	}).Trace("Alerts index cleaning successful")

	log.WithFields(log.Fields{
		"index": alertsIndexName,
	}).Info("Triggering force merge on alerts index")

	timer.Stage("force_merge")

	// NOTE: not doing multiple retries here because this is such a costly operation, if it
	// errors out repeatedly trying it could cripple performance
	_, err = c.esclient.Forcemerge(alertsIndexName).IgnoreUnavailable(true).Do(context.Background())
	if err != nil {
		log.WithFields(log.Fields{
			"err":   err,
			"index": alertsIndexName,
		}).Error("Error force-merging alerts index")
	}
	log.WithFields(log.Fields{
		"index": alertsIndexName,
	}).Trace("Alerts index force-merge successful")
	return err
}

// CleanErrorLogs deletes all error logs that are older than the given age in days and marks any older than today as read-only
func (c *Client) CleanErrorLogs(maxAgeInDays int) error {
	log.WithFields(log.Fields{
		"max_age_in_days": maxAgeInDays,
	}).Trace("Beginning error log cleaning")

	timer := timing.NewStageTimer("Client.CleanErrorLogs", log.Fields{})
	defer timer.Finish()

	indices, err := c.getErrorLogIndices(context.Background())
	if err != nil {
		log.WithError(err).Error("Error getting error log indices")
		return err
	}
	toDelete := []string{}
	toReadOnly := []string{}

	timer.Stage("process_index_names")

	for _, index := range indices {
		date, err := getTimeFromErrorLogIndexName(index)
		if err != nil {
			log.WithField("index", index).Warn("Index has invalid date format, skipping; it will be retained")
			continue
		}
		now := time.Now().UTC()
		nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		ageInHours := nowDate.Sub(date).Hours()
		// 24 hours times the max age for deletion (plus 1 for days)
		if ageInHours > float64(24*maxAgeInDays) {
			log.WithFields(log.Fields{
				"index":         index,
				"age_hours":     ageInHours,
				"max_age_hours": maxAgeInDays * 24,
			}).Info("Index is past retention date, deleting")
			toDelete = append(toDelete, index)
		} else if ageInHours > 24 {
			log.WithFields(log.Fields{
				"index":         index,
				"age_hours":     ageInHours,
				"max_age_hours": maxAgeInDays * 24,
			}).Info("Index is older than 1 day, marking as read-only")
			toReadOnly = append(toReadOnly, index)
		} else {
			log.WithFields(log.Fields{
				"index":         index,
				"age_hours":     ageInHours,
				"max_age_hours": maxAgeInDays * 24,
			}).Info("Index is fresh, not touching")
		}
	}

	timer.Stage("delete_indices")

	if len(toDelete) > 0 {
		log.WithFields(log.Fields{
			"to_delete": toDelete,
		}).Trace("Beginning error log index deletion")
		err = c.DeleteIndices(context.Background(), toDelete)
		if err != nil {
			log.WithFields(log.Fields{
				"err":       err,
				"to_delete": toDelete,
			}).Error("Error deleting old indices")
			return err
		}
		log.WithFields(log.Fields{
			"to_delete": toDelete,
		}).Trace("Error log index deletion successful")
	}

	timer.Stage("mark_indices_read_only")

	if len(toReadOnly) > 0 {
		err = c.tryRepeatReturnErrorOnly(func() error {
			log.WithFields(log.Fields{
				"to_read_only": toReadOnly,
			}).Trace("Beginning error log index read-only settings change")
			_, err := c.esclient.IndexPutSettings(toReadOnly...).BodyJson(map[string]interface{}{
				"index": map[string]interface{}{
					"blocks": map[string]interface{}{
						"read_only_allow_delete": true,
					},
				},
			}).Do(context.Background())
			if err != nil {
				log.WithFields(log.Fields{
					"to_read_only": toReadOnly,
				}).Error("Error setting error log index read-only setting")
				return err
			}
			log.WithFields(log.Fields{
				"to_read_only": toReadOnly,
			}).Trace("Error log index read-only settings change successful")

			return nil
		})
		if err != nil {
			log.WithError(err).Error("Error marking indices as read-only, but continuing (non-fatal)")
		}
	}

	log.WithFields(log.Fields{
		"max_age_in_days": maxAgeInDays,
	}).Trace("Error log cleaning finished")
	return nil
}

// CleanTimerLogs deletes all timer logs that are older than the given age in days and marks any older than today as read-only
func (c *Client) CleanTimerLogs(maxAgeInDays int) error {
	log.WithFields(log.Fields{
		"max_age_in_days": maxAgeInDays,
	}).Trace("Beginning timer log cleaning")

	timer := timing.NewStageTimer("Client.CleanTimerLogs", log.Fields{})
	defer timer.Finish()

	indices, err := c.getTimerLogIndices(context.Background())
	if err != nil {
		log.WithError(err).Error("Error getting timer log indices")
		return err
	}
	toDelete := []string{}
	toReadOnly := []string{}

	timer.Stage("process_index_names")

	for _, index := range indices {
		date, err := getTimeFromTimerLogIndexName(index)
		if err != nil {
			log.WithField("index", index).Warn("Index has invalid date format, skipping; it will be retained")
			continue
		}
		now := time.Now().UTC()
		nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		ageInHours := nowDate.Sub(date).Hours()
		// 24 hours times the max age for deletion (plus 1 for days)
		if ageInHours > float64(24*maxAgeInDays) {
			log.WithFields(log.Fields{
				"index":         index,
				"age_hours":     ageInHours,
				"max_age_hours": maxAgeInDays * 24,
			}).Info("Index is past retention date, deleting")
			toDelete = append(toDelete, index)
		} else if ageInHours > 24 {
			log.WithFields(log.Fields{
				"index":         index,
				"age_hours":     ageInHours,
				"max_age_hours": maxAgeInDays * 24,
			}).Info("Index is older than 1 day, marking as read-only")
			toReadOnly = append(toReadOnly, index)
		} else {
			log.WithFields(log.Fields{
				"index":         index,
				"age_hours":     ageInHours,
				"max_age_hours": maxAgeInDays * 24,
			}).Info("Index is fresh, not touching")
		}
	}

	timer.Stage("delete_indices")

	if len(toDelete) > 0 {
		log.WithFields(log.Fields{
			"to_delete": toDelete,
		}).Trace("Beginning timer log index deletion")
		err = c.DeleteIndices(context.Background(), toDelete)
		if err != nil {
			log.WithFields(log.Fields{
				"err":       err,
				"to_delete": toDelete,
			}).Error("Error deleting old indices")
			return err
		}
		log.WithFields(log.Fields{
			"to_delete": toDelete,
		}).Trace("Timer log index deletion successful")
	}

	timer.Stage("mark_indices_read_only")

	if len(toReadOnly) > 0 {
		err = c.tryRepeatReturnErrorOnly(func() error {
			log.WithFields(log.Fields{
				"to_read_only": toReadOnly,
			}).Trace("Beginning timer log index read-only settings change")
			_, err := c.esclient.IndexPutSettings(toReadOnly...).BodyJson(map[string]interface{}{
				"index": map[string]interface{}{
					"blocks": map[string]interface{}{
						"read_only_allow_delete": true,
					},
				},
			}).Do(context.Background())
			if err != nil {
				log.WithFields(log.Fields{
					"to_read_only": toReadOnly,
				}).Error("Error setting timer log index read-only setting")
				return err
			}
			log.WithFields(log.Fields{
				"to_read_only": toReadOnly,
			}).Trace("Timer log index read-only settings change successful")

			return nil
		})
		if err != nil {
			log.WithError(err).Error("Error marking indices as read-only, but continuing (non-fatal)")
		}
	}

	log.WithFields(log.Fields{
		"max_age_in_days": maxAgeInDays,
	}).Trace("Timer log cleaning finished")
	return nil
}
