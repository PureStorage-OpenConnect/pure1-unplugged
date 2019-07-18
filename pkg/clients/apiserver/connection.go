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

package apiserver

import (
	"fmt"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/http"
	"github.com/go-resty/resty"

	log "github.com/sirupsen/logrus"
)

// Type guards: ensure this implements the interfaces
var _ resources.ArrayDiscovery = (*APIServer)(nil)
var _ resources.ArrayMetadata = (*APIServer)(nil)

// NewConnection establishes a connection with the API server
// at the given backend URL. Note that this returns a pointer to an API
// server, not to a specific interface, as it implements multiple interfaces.
func NewConnection(backendURL string) *APIServer {
	return &APIServer{
		serverEndpoint: backendURL,
	}
}

// GetArrays is an implementation of the ArrayDiscovery interface
func (a *APIServer) GetArrays() ([]*resources.ArrayRegistrationInfo, error) {
	log.WithField("endpoint", a.serverEndpoint).Trace("Starting API server device list GET")
	uncastResponse, err := http.RestyGet(bulkDeviceResponse{}, resty.R(), fmt.Sprintf("%s/arrays", a.serverEndpoint))
	if err != nil {
		return nil, err
	}
	if response, ok := uncastResponse.(*bulkDeviceResponse); ok {
		return response.Items, nil
	}
	return nil, fmt.Errorf("Error casting response to bulkDeviceResponse")
}

// Patch patches the given device with the given body
func (a *APIServer) Patch(arrayID string, body *resources.ArrayPatchInfo) error {
	resp, err := resty.R().SetQueryParam("ids", arrayID).SetBody(*body).Patch(fmt.Sprintf("%s/arrays", a.serverEndpoint))
	if err != nil {
		return err
	}
	if resp.IsError() {
		return fmt.Errorf("Error %s", resp.Status())
	}

	return nil
}

// GetTags fetches the tags for the given device from the API server
func (a *APIServer) GetTags(arrayID string) (map[string]string, error) {
	log.WithField("endpoint", a.serverEndpoint).Trace("Starting API server device tags GET")
	uncastResponse, err := http.RestyGet(bulkTagsResponse{}, resty.R().SetQueryParam("ids", arrayID), fmt.Sprintf("%s/arrays/tags", a.serverEndpoint))
	if err != nil {
		return nil, err
	}
	if response, ok := uncastResponse.(*bulkTagsResponse); ok {
		if len(response.Items) == 0 {
			return nil, fmt.Errorf("Response contained no arrays")
		}
		for _, array := range response.Items {
			if array.ID != arrayID {
				continue
			}

			parsedTags := map[string]string{}
			for _, tag := range array.Tags {
				if tag.Key == "" {
					continue
				}
				parsedTags[tag.Key] = tag.Value
			}

			return parsedTags, nil
		}

		return nil, fmt.Errorf("Response did not contain an array with the matching ID")
	}
	return nil, fmt.Errorf("Error casting response to bulkTagsResponse")
}
