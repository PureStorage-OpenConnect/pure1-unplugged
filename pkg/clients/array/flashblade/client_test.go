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

package flashblade

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	TestArrayEndpoint = ""
	TestArrayToken    = ""
)

func TestFlashBladeClientWrongEndpoint(t *testing.T) {
	_, err := NewClient("test-client", "101.103.45.103", TestArrayToken)
	assert.Error(t, err)
}

func TestFlashBladeClientInvalidEndpoint(t *testing.T) {
	_, err := NewClient("test-client", "https://aaaaaa.com", TestArrayToken)
	assert.Error(t, err)
}

func TestFlashBladeClientInvalidToken(t *testing.T) {
	t.Skip("Waiting for mock interceptor")

	client, err := NewClient("test-client", TestArrayEndpoint, "nope")
	assert.NoError(t, err)

	_, err = client.GetArrayInfo()
	assert.Error(t, err)
}

func TestFlashBladeClient(t *testing.T) {
	t.Skip("Waiting for mock interceptor")

	logrus.SetLevel(logrus.TraceLevel)

	client, err := NewClient("test-client", TestArrayEndpoint, TestArrayToken)
	assert.NoError(t, err)

	var response interface{}

	response, err = client.GetAlerts()
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

	response, err = client.GetFileSystemCapacityMetrics()
	assert.NoError(t, err)
	assert.NotNil(t, response)
	numFileSystems := uint32(len(response.([]*FileSystemCapacityMetricsResponse)))

	response, err = client.GetFileSystemCount()
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, numFileSystems, response)

	response, err = client.GetFileSystemPerformanceMetrics(120)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	// We expect 2 points per file system
	assert.Equal(t, numFileSystems*4, uint32(len(response.([]*FileSystemPerformanceMetricsResponse))))

	response, err = client.GetFileSystemSnapshotCount()
	assert.NoError(t, err)
	assert.NotNil(t, response)
	numSnapshots := response

	response, err = client.GetFileSystemSnapshots()
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, numSnapshots, uint32(len(response.([]*FileSystemSnapshotResponse))))
}
