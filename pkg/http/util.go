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

package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/http/errors"
	"github.com/go-resty/resty"
)

// RestyGet performs a GET request and places the parsed JSON result in
func RestyGet(result interface{}, request *resty.Request, url string) (interface{}, error) {
	resp, err := request.SetResult(result).Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("Error %s", resp.Status())
	}
	return resp.Result(), nil
}

// ReadBody reads all contents of a given request body, returning
// an Internal Server Error if it fails
func ReadBody(r *http.Request) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}

	if err = r.Body.Close(); err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}
	return body, nil
}

// ReadResponseBody reads all contents of a given response body
func ReadResponseBody(r *http.Response) ([]byte, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if err = r.Body.Close(); err != nil {
		return nil, err
	}
	return body, nil
}

// ParseBodyToMap parses a JSON body to map[string]interface{}, returning
// either an Internal Server Error or Bad Request Error on failure (depending on the issue)
func ParseBodyToMap(r *http.Request) (map[string]interface{}, error) {
	body, err := ReadBody(r)
	if err != nil {
		return nil, errors.MakeInternalHTTPErr(err)
	}

	var mapped map[string]interface{}
	err = json.Unmarshal(body, &mapped)
	if err != nil {
		// Invalid JSON is an invalid request
		return mapped, errors.MakeBadRequestHTTPErr(err)
	}
	return mapped, nil
}

// ParseResponseBodyToMap parses a JSON body to map[string]interface{}, returning
// either an Internal Server Error or Bad Request Error on failure (depending on the issue)
func ParseResponseBodyToMap(r *http.Response) (map[string]interface{}, error) {
	body, err := ReadResponseBody(r)
	if err != nil {
		return nil, err
	}

	var mapped map[string]interface{}
	err = json.Unmarshal(body, &mapped)
	if err != nil {
		return mapped, err
	}
	return mapped, nil
}

// EnsureKeysAreFilled ensures that the given object (usually parsed from JSON)
// has the given keys and that given keys are not empty (all blank space)
func EnsureKeysAreFilled(m map[string]interface{}, keys ...string) error {
	for _, key := range keys {
		if _, ok := m[key]; !ok {
			return fmt.Errorf("Key %s is not present", key)
		}
		if len(strings.TrimSpace(m[key].(string))) == 0 {
			return fmt.Errorf("Key %s must not be empty", key)
		}
	}
	return nil
}
