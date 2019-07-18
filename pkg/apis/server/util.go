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
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/http/errors"
	log "github.com/sirupsen/logrus"
)

func parseRequestQueryParams(r *http.Request) (resources.ArrayQuery, error) {
	var ids []string
	if len(r.FormValue("ids")) > 0 {
		idsSplit := strings.Split(r.FormValue("ids"), ",")
		// Validate that all IDs are actually valid IDs
		for _, id := range idsSplit {
			err := resources.ValidateHexObjectID(id)
			if err != nil {
				return resources.ArrayQuery{}, errors.MakeBadRequestHTTPErr(err)
			}
		}
		ids = idsSplit
	} else {
		ids = []string{}
	}

	var names []string
	if len(r.FormValue("names")) > 0 {
		names = strings.Split(r.FormValue("names"), ",")
	} else {
		names = []string{}
	}

	var models []string
	if len(r.FormValue("models")) > 0 {
		models = strings.Split(r.FormValue("models"), ",")
	} else {
		models = []string{}
	}

	var versions []string
	if len(r.FormValue("versions")) > 0 {
		versions = strings.Split(r.FormValue("versions"), ",")
	} else {
		versions = []string{}
	}

	limit := 0
	if len(r.FormValue("limit")) > 0 {
		parsedLimit, err := strconv.ParseInt(r.FormValue("limit"), 10, 64)
		if err != nil {
			return resources.ArrayQuery{}, errors.MakeBadRequestHTTPErr(err)
		}
		if parsedLimit < 1 {
			return resources.ArrayQuery{}, errors.MakeBadRequestHTTPErr(fmt.Errorf("Limit must be >= 1"))
		}
		limit = int(parsedLimit)
	}

	offset := 0
	if len(r.FormValue("offset")) > 0 {
		parsedOffset, err := strconv.ParseInt(r.FormValue("offset"), 10, 64)
		if err != nil {
			return resources.ArrayQuery{}, errors.MakeBadRequestHTTPErr(err)
		}
		if parsedOffset < 0 {
			return resources.ArrayQuery{}, errors.MakeBadRequestHTTPErr(fmt.Errorf("Offset must be >= 0"))
		}
		offset = int(parsedOffset)
	}

	sortField := ""
	sortDesc := false
	if len(r.FormValue("sort")) > 0 {
		sortStr := r.FormValue("sort")
		if sortStr[len(sortStr)-1] == '-' {
			sortDesc = true
			sortField = sortStr[:len(sortStr)-1]
		} else {
			sortDesc = false
			sortField = sortStr
		}
	}

	return resources.ArrayQuery{
		Ids:            ids,
		Names:          names,
		Models:         models,
		Versions:       versions,
		Limit:          limit,
		Offset:         offset,
		Sort:           sortField,
		SortDescending: sortDesc,
	}, nil
}

func respond(w http.ResponseWriter, code int, body interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(code)
	if writeErr := json.NewEncoder(w).Encode(body); writeErr != nil {
		log.WithFields(log.Fields{
			"writeError": writeErr,
			"body":       body,
		}).Error("An error occurred while writing the original content")
		_, writeErr = w.Write([]byte("An error occurred while writing the original content"))
		if writeErr != nil {
			log.WithError(writeErr).Error("An error occurred while writing the error about the original content writing")
			// Not really a better way to handle this: we can't overwrite the response code since we
			// already started to write content earlier
		}
	}
}

func respondWithErrorCode(w http.ResponseWriter, err error, code int) {
	respond(w, code, errors.JSONErr{Code: code, Text: err.Error()})
}

func respondWithSuccess(w http.ResponseWriter, res interface{}) {
	respond(w, http.StatusOK, res)
}

// handleError is a way to generically respond to errors
// fired by a handler: if it's an HTTPErr, it will respond
// using the given code: if not, it'll treat it as an
// internal server error
func handleError(w http.ResponseWriter, err error) {
	if httpErr, ok := err.(*errors.HTTPErr); ok {
		respondWithErrorCode(w, httpErr.Inner, httpErr.Code)
	} else {
		respondWithErrorCode(w, err, http.StatusInternalServerError)
	}
}
