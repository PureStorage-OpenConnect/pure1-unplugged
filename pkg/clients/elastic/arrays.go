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
	"encoding/json"
	"reflect"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/http/errors"
	log "github.com/sirupsen/logrus"
)

// Type guard: ensure this implements the interface
var _ resources.ArrayDatabase = (*Client)(nil)

// FindArrays searches with the given query to find arrays. Note that this leaves
// the API token blank, and thus it will need to be filled in by the calling method.
func (c *Client) FindArrays(query *resources.ArrayQuery) ([]*resources.Array, error) {
	ctx := context.Background()

	err := c.EnsureConnected(ctx)
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}

	searchService := c.esclient.Search(arraysIndexName).Type("_doc").Query(query.GenerateElasticQueryObject()).From(query.Offset).IgnoreUnavailable(true).AllowNoIndices(true)
	if query.Limit > 0 {
		searchService.Size(query.Limit)
	} else {
		// Default to 1000 results if not specified (should be bigger than any "practical" fleet size to ensure returning
		// all results)
		searchService.Size(1000)
	}
	if len(query.Sort) > 0 {
		sortParam, err := query.GetSortParameter()
		if err != nil {
			return nil, err
		}
		searchService.Sort(sortParam, !query.SortDescending)
	}

	res, err := searchService.Do(ctx)

	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}

	results := []*resources.Array{}

	for _, res := range res.Each(reflect.TypeOf(&resources.Array{})) {
		device := res.(*resources.Array)
		results = append(results, device)
	}

	return results, nil
}

// InsertArray inserts the given storage device into Elastic.
func (c *Client) InsertArray(device *resources.Array) error {
	ctx := context.Background()

	err := c.EnsureConnected(ctx)
	if err != nil {
		return errors.MakeInternalHTTPErr(err)
	}

	originalArray := *device
	copied := originalArray // Create a copy so we don't modify the original

	copied.APIToken = "" // Clear the API token so we don't store it in Elastic
	log.WithField("device_id", copied.InternalID).Trace("Beginning to push device to Elastic")
	_, err = c.esclient.Index().Index(arraysIndexName).Type("_doc").Id(copied.InternalID).BodyJson(&copied).Do(ctx)
	if err != nil {
		return errors.MakeInternalHTTPErr(err)
	}
	log.WithField("device_id", device.InternalID).Trace("Device pushed to Elastic successfully")
	return nil
}

// PatchArray patches the given device's fields (except for tags).
func (c *Client) PatchArray(device *resources.Array) (*resources.Array, error) {
	ctx := context.Background()

	err := c.EnsureConnected(ctx)
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}

	// Note: I'm not sure if we need originalArray. I figured to be safe I'd include it, worst case it's a bit more memory used
	originalArray := *device
	copied := originalArray // Create a copy so we don't modify the original

	copied.Tags = nil // Remove tags since this endpoint shouldn't be able to modify them
	copied.APIToken = ""

	log.WithField("device_id", copied.InternalID).Trace("Beginning device update to Elastic")
	// We're fine if the index doesn't exist, the patch would fail anyways because there's nothing to patch
	res, err := c.esclient.Update().Index(arraysIndexName).Type("_doc").Id(copied.InternalID).Doc(&copied).FetchSource(true).Do(ctx)
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}
	log.WithField("device_id", copied.InternalID).Trace("Device update to Elastic successful")

	marshalledBytes, err := res.GetResult.Source.MarshalJSON()
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}
	parsed := &resources.Array{}
	err = json.Unmarshal(marshalledBytes, &parsed)
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}

	return parsed, nil
}

// PatchArrayTags patches the tags on the given device.
func (c *Client) PatchArrayTags(device *resources.Array) (*resources.Array, error) {
	ctx := context.Background()

	err := c.EnsureConnected(ctx)
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}
	justTags := map[string]interface{}{
		"Tags":        device.Tags,
		"LastUpdated": device.Lastupdated.Format(resources.DateTimeFormat),
	}

	// We're fine if the index doesn't exist, the patch would fail anyways because there's nothing to patch
	res, err := c.esclient.Update().Index(arraysIndexName).Type("_doc").Id(device.InternalID).Doc(&justTags).FetchSource(true).Do(ctx)
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}

	marshalledBytes, err := res.GetResult.Source.MarshalJSON()
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}
	parsed := &resources.Array{}
	err = json.Unmarshal(marshalledBytes, &parsed)
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}

	return parsed, nil
}

// DeleteArray deletes the devices matching the given query from Elastic.
func (c *Client) DeleteArray(query *resources.ArrayQuery) ([]string, error) {
	ctx := context.Background()

	err := c.EnsureConnected(ctx)
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}

	// Get all the devices that match this query first, so we can delete the tokens by ID
	res, err := c.esclient.Search(arraysIndexName).Type("_doc").Query(query.GenerateElasticQueryObject()).IgnoreUnavailable(true).AllowNoIndices(true).Do(ctx)
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}

	ids := []string{}

	for _, res := range res.Each(reflect.TypeOf(&resources.Array{})) {
		array := res.(*resources.Array)
		ids = append(ids, array.InternalID)
	}

	elasticQuery := query.GenerateElasticQueryObject()

	_, err = c.esclient.DeleteByQuery(arraysIndexName).Type("_doc").Query(elasticQuery).Do(ctx)
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}
	return ids, err
}
