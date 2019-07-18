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

package flasharray

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	testArrayEndpoint = ""
	testArrayToken    = ""
)

func TestFlashArrayClientWrongEndpoint(t *testing.T) {
	_, err := NewClient("test-client", "10.14.75.103", testArrayToken)
	assert.Error(t, err)
}

func TestFlashArrayClientInvalidEndpoint(t *testing.T) {
	_, err := NewClient("test-client", "https://aaaaaa.com", testArrayToken)
	assert.Error(t, err)
}

func TestFlashArrayClientInvalidToken(t *testing.T) {
	t.Skip("Waiting for mock interceptor")

	client, err := NewClient("test-client", testArrayEndpoint, "nope")
	assert.NoError(t, err)

	_, err = client.GetArrayInfo()
	assert.Error(t, err)
}

func TestFlashArrayClient(t *testing.T) {
	t.Skip("Waiting for mock interceptor")

	logrus.SetLevel(logrus.TraceLevel)

	client, err := NewClient("test-client", testArrayEndpoint, testArrayToken)
	assert.NoError(t, err)

	var response interface{}

	response, err = client.GetAlertsClosed()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = client.GetAlertsFlagged()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = client.GetAlertsOpen()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = client.GetArrayCapacityMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = client.GetArrayInfo()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = client.GetArrayPerformanceMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = client.GetHostCount()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = client.GetModel()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = client.GetVolumeCapacityMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = client.GetVolumeCount()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = client.GetVolumePendingEradicationCount()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = client.GetVolumePerformanceMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = client.GetVolumeSnapshotCount()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = client.GetVolumeSnapshots()
	assert.NoError(t, err)
	assert.NotNil(t, response)
}
