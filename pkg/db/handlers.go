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

package db

import (
	"strings"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/http/errors"
	"gopkg.in/mgo.v2/bson"

	log "github.com/sirupsen/logrus"
)

func (h *MetadataConnection) populateAPIToken(array *resources.Array) error {
	token, err := h.Tokens.GetToken(array.InternalID)
	if err != nil {
		return err
	}
	array.APIToken = token
	return nil
}

// GetArrays fetches all the arrays that match the given query
func (h *MetadataConnection) GetArrays(query resources.ArrayQuery) (BulkResponse, error) {
	results, err := h.DAO.FindArrays(&query)
	if err != nil {
		return BulkResponse{}, err
	}

	arrayMaps := []map[string]interface{}{}
	for _, array := range results {
		h.populateAPIToken(array)
		arrayMaps = append(arrayMaps, array.ConvertToArrayMap())
	}

	return BulkResponse{Response: arrayMaps}, nil
}

// GetArrayStatuses fetches all the array statuses that match the given query
func (h *MetadataConnection) GetArrayStatuses(query resources.ArrayQuery) (BulkResponse, error) {
	results, err := h.DAO.FindArrays(&query)
	if err != nil {
		return BulkResponse{}, err
	}

	statusMaps := []map[string]interface{}{}
	for _, array := range results {
		statusMaps = append(statusMaps, array.ConvertToStatusMap())
	}

	return BulkResponse{Response: statusMaps}, nil
}

// GetArrayTags fetches all the array tags that match the given query
func (h *MetadataConnection) GetArrayTags(query resources.ArrayQuery) (BulkResponse, error) {
	results, err := h.DAO.FindArrays(&query)
	if err != nil {
		return BulkResponse{}, err
	}

	tagMaps := []map[string]interface{}{}
	for _, array := range results {
		tagMaps = append(tagMaps, array.ConvertToTagsMap())
	}

	return BulkResponse{Response: tagMaps}, nil
}

// PostArray registers a new array to the given database
func (h *MetadataConnection) PostArray(m map[string]interface{}) (map[string]interface{}, error) {
	parsed, err := resources.ParseArrayFromREST(m)
	if err != nil {
		return nil, err
	}

	parsed.InternalID = bson.NewObjectId().Hex()
	parsed.Lastseen = time.Time{} // It's never been seen, by definition: don't let someone fake it
	parsed.Lastupdated = time.Now().UTC()
	parsed.Status = "Connecting"

	err = parsed.HasRequiredPostFields()
	if err != nil {
		return nil, errors.MakeBadRequestHTTPErr(err)
	}

	err = h.Tokens.SaveToken(parsed.InternalID, parsed.APIToken)
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}

	err = h.DAO.InsertArray(&parsed)
	if err != nil {
		return nil, err
	}
	return parsed.ConvertToArrayMap(), nil
}

// PatchArrays updates all arrays matching the given query
func (h *MetadataConnection) PatchArrays(query resources.ArrayQuery, m map[string]interface{}) (BulkResponse, error) {
	arrays, err := h.DAO.FindArrays(&query)
	if err != nil {
		return BulkResponse{}, err
	}

	patchedArrays := []*resources.Array{}

	// Apply the patch locally, checking for errors as we do
	for _, array := range arrays {
		err = array.ApplyPatch(m)
		if err != nil {
			return BulkResponse{}, errors.MakeBadRequestHTTPErr(err)
		}
		// Copy the patched array over to the new list
		patchedArrays = append(patchedArrays, array)
	}

	responses := []map[string]interface{}{}

	// Push the patched arrays to the backing database
	for _, array := range patchedArrays {
		if len(strings.TrimSpace(array.APIToken)) != 0 {
			// Technically this may sometimes save the API token when it doesn't need to, but that's fine: good security
			// and it's not that slow
			err = h.Tokens.SaveToken(array.InternalID, array.APIToken)
			if err != nil {
				return BulkResponse{}, errors.MakeInternalHTTPErr(err)
			}
		}

		newArray, err := h.DAO.PatchArray(array)
		if err != nil {
			return BulkResponse{}, err
		}

		fetchedToken, err := h.Tokens.GetToken(newArray.InternalID)
		if err != nil {
			return BulkResponse{}, err
		}

		newArray.APIToken = fetchedToken
		responses = append(responses, newArray.ConvertToArrayMap())
	}

	return BulkResponse{Response: responses}, nil
}

// PatchArrayTags updates the tags of all arrays matching the given query
func (h *MetadataConnection) PatchArrayTags(query resources.ArrayQuery, m []map[string]string) (BulkResponse, error) {
	arrays, err := h.DAO.FindArrays(&query)
	if err != nil {
		return BulkResponse{}, err
	}

	patchedArrays := []*resources.Array{}

	// Apply the patch locally, checking for errors as we do
	for _, array := range arrays {
		err = array.ApplyTagPatch(m)
		if err != nil {
			return BulkResponse{}, errors.MakeBadRequestHTTPErr(err)
		}
		patchedArrays = append(patchedArrays, array)
	}

	responses := []map[string]interface{}{}

	// Push the patched arrays to the backing database
	for _, array := range patchedArrays {
		newArray, err := h.DAO.PatchArrayTags(array)
		if err != nil {
			return BulkResponse{}, err
		}
		responses = append(responses, newArray.ConvertToTagsMap())
	}

	return BulkResponse{Response: responses}, nil
}

// DeleteArrays deletes all arrays that match the given query, and returns the count of arrays deleted
func (h *MetadataConnection) DeleteArrays(query resources.ArrayQuery) (int, error) {
	ids, err := h.DAO.DeleteArray(&query)
	if err != nil {
		return 0, err
	}

	var lastTokenErr error

	for _, id := range ids {
		err = h.Tokens.DeleteToken(id)
		if err != nil {
			lastTokenErr = err
			log.WithError(err).WithFields(log.Fields{
				"array_id": id,
			}).Error("Error deleting token for device: continuing to delete the rest of the tokens, but this call to DeleteArrays will fail at the end")
		}
	}

	if lastTokenErr != nil {
		return 0, lastTokenErr
	}

	return len(ids), nil
}

// DeleteArrayTags deletes the given tags from all arrays matching the given query
func (h *MetadataConnection) DeleteArrayTags(query resources.ArrayQuery, tags []string) (BulkResponse, error) {
	arrays, err := h.DAO.FindArrays(&query)
	if err != nil {
		return BulkResponse{}, err
	}

	patchedArrays := []*resources.Array{}

	// Apply the patch locally, checking for errors as we do
	for _, array := range arrays {
		array.DeleteTags(tags)
		patchedArrays = append(patchedArrays, array)
	}

	responses := []map[string]interface{}{}

	// Push the patched arrays to the backing database
	for _, array := range patchedArrays {
		newArray, err := h.DAO.PatchArrayTags(array)
		if err != nil {
			return BulkResponse{}, err
		}
		responses = append(responses, newArray.ConvertToTagsMap())
	}

	return BulkResponse{Response: responses}, nil
}
