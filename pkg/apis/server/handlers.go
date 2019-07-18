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
	"net/http"
	"strings"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/db"

	purehttp "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/http"
)

var (
	connection db.MetadataConnection
)

func getArrays(w http.ResponseWriter, r *http.Request) {
	query, err := parseRequestQueryParams(r)
	if err != nil {
		handleError(w, err)
		return
	}

	results, err := connection.GetArrays(query)
	if err != nil {
		handleError(w, err)
		return
	}

	respondWithSuccess(w, results)
}

func postArray(w http.ResponseWriter, r *http.Request) {
	mapped, err := purehttp.ParseBodyToMap(r)
	if err != nil {
		handleError(w, err)
		return
	}

	err = purehttp.EnsureKeysAreFilled(mapped, "api_token", "device_type", "mgmt_endpoint", "name")
	if err != nil {
		respondWithErrorCode(w, err, http.StatusBadRequest)
		return
	}

	result, err := connection.PostArray(mapped)
	if err != nil {
		handleError(w, err)
		return
	}
	respondWithSuccess(w, result)
}

func patchArray(w http.ResponseWriter, r *http.Request) {
	query, err := parseRequestQueryParams(r)
	if err != nil {
		handleError(w, err)
		return
	}

	mapped, err := purehttp.ParseBodyToMap(r)
	if err != nil {
		handleError(w, err)
		return
	}

	res, err := connection.PatchArrays(query, mapped)

	if err != nil {
		handleError(w, err)
		return
	}
	respondWithSuccess(w, res)
}

func deleteArray(w http.ResponseWriter, r *http.Request) {
	query, err := parseRequestQueryParams(r)
	if err != nil {
		handleError(w, err)
		return
	}
	if query.IsEmpty() {
		respondWithErrorCode(w, fmt.Errorf("At least one query parameter must be specified (ids, names, models, versions)"), http.StatusBadRequest)
		return
	}

	count, err := connection.DeleteArrays(query)
	if err != nil {
		handleError(w, err)
		return
	}

	response := map[string]interface{}{
		"deletedCount": count,
	}
	respondWithSuccess(w, response)
}

func getArrayStatus(w http.ResponseWriter, r *http.Request) {
	query, err := parseRequestQueryParams(r)
	if err != nil {
		handleError(w, err)
		return
	}

	results, err := connection.GetArrayStatuses(query)
	if err != nil {
		handleError(w, err)
		return
	}

	respondWithSuccess(w, results)
}

func getArrayTags(w http.ResponseWriter, r *http.Request) {
	query, err := parseRequestQueryParams(r)
	if err != nil {
		handleError(w, err)
		return
	}

	results, err := connection.GetArrayTags(query)
	if err != nil {
		handleError(w, err)
		return
	}

	respondWithSuccess(w, results)
}

func patchArrayTags(w http.ResponseWriter, r *http.Request) {
	query, err := parseRequestQueryParams(r)
	if err != nil {
		handleError(w, err)
		return
	}

	mapped, err := purehttp.ParseBodyToMap(r)
	if err != nil {
		handleError(w, err)
		return
	}

	if _, ok := mapped["tags"]; !ok {
		respondWithErrorCode(w, fmt.Errorf("Key tags is not present"), http.StatusBadRequest)
		return
	}

	// Convert tags map to string->string array
	originalTags := mapped["tags"].([]interface{})
	convertedTags := []map[string]string{}
	for _, tag := range originalTags {
		tagMap, ok := tag.(map[string]interface{})
		if !ok {
			respondWithErrorCode(w, fmt.Errorf("Array item is not a map"), http.StatusBadRequest)
			return
		}

		convertedTag := map[string]string{}

		for key, value := range tagMap {
			// Do the interface{} -> string conversion
			stringValue, ok := value.(string)
			if !ok {
				respondWithErrorCode(w, fmt.Errorf("Error converting tag value into string. Offending key-value pair is '%s: %v'", key, value), http.StatusBadRequest)
				return
			}
			convertedTag[key] = stringValue
		}

		convertedTags = append(convertedTags, convertedTag)
	}

	res, err := connection.PatchArrayTags(query, convertedTags)

	if err != nil {
		handleError(w, err)
		return
	}
	respondWithSuccess(w, res)
}

func deleteArrayTags(w http.ResponseWriter, r *http.Request) {
	query, err := parseRequestQueryParams(r)
	if err != nil {
		handleError(w, err)
		return
	}

	if len(r.FormValue("tags")) == 0 {
		respondWithErrorCode(w, fmt.Errorf("Key tags must not be empty"), http.StatusBadRequest)
		return
	}

	tags := strings.Split(r.FormValue("tags"), ",")

	res, err := connection.DeleteArrayTags(query, tags)

	if err != nil {
		handleError(w, err)
		return
	}

	respondWithSuccess(w, res)
}
