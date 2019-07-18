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

import "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"

// APIServer is a connection to the Pure1 Unplugged API server using a raw REST
// client.
type APIServer struct {
	serverEndpoint string
}

type bulkDeviceResponse struct {
	Items []*resources.ArrayRegistrationInfo `json:"response"`
}

type bulkTagsResponse struct {
	Items []*tagsResponse `json:"response"`
}

type tagsResponse struct {
	ID   string `json:"id"`
	Tags []*tag `json:"tags"`
}

type tag struct {
	Key       string `json:"key"`
	Namespace string `json:"namespace"`
	Value     string `json:"value"`
}
