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
	"fmt"
	"testing"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/mock"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	testArrayEndpoint2 = ""
	testArrayToken2    = ""
)

func TestFlashArrayCollectorInvalidEndpoint(t *testing.T) {
	_, err := NewCollector("000000000000000000000000", "test-array", "101.241.128.13", testArrayToken2, nil)
	assert.Error(t, err)
}

func TestFlashArrayCollectorInvalidToken(t *testing.T) {
	t.Skip("Waiting for mock interceptor")

	client, err := NewCollector("000000000000000000000000", "test-array", testArrayEndpoint2, "nah", nil)
	assert.NoError(t, err)

	_, err = client.GetArrayName()
	assert.Error(t, err)
}

func TestFlashArrayCollector(t *testing.T) {
	t.Skip("Waiting for mock interceptor")

	logrus.SetLevel(logrus.TraceLevel)

	metaInterface := &mock.ArrayMetadataImpl{}
	metaInterface.On("GetTags", "000000000000000000000000").Return(map[string]string{
		"tag1": "value1",
	}, nil)

	collector, err := NewCollector("000000000000000000000000", "test-array", testArrayEndpoint2, testArrayToken2, metaInterface)
	assert.NoError(t, err)

	var response interface{}

	arrayDataResponse, err := collector.GetAllArrayData()
	assert.NoError(t, err)
	assert.NotNil(t, arrayDataResponse)
	assert.Len(t, arrayDataResponse.ArrayMetric.Tags, 1)
	assert.Equal(t, "value1", arrayDataResponse.ArrayMetric.Tags["tag1"])

	volumeDataResponse, err := collector.GetAllVolumeData(30)
	assert.NoError(t, err)
	assert.NotNil(t, volumeDataResponse)
	if len(volumeDataResponse.VolumeMetricsTimeSeries) > 0 {
		// Test the first volume to see if it has the right tags
		firstVolume := volumeDataResponse.VolumeMetricsTimeSeries[0]
		assert.Len(t, firstVolume.ArrayTags, 1)
		assert.Equal(t, "value1", firstVolume.ArrayTags["tag1"])
	}

	response = collector.GetArrayID()
	assert.Equal(t, "000000000000000000000000", response)

	response, err = collector.GetArrayModel()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = collector.GetArrayName()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = collector.GetArrayVersion()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	tagResponse, err := collector.GetArrayTags()
	assert.NoError(t, err)
	assert.NotNil(t, tagResponse)
	assert.Len(t, tagResponse, 1)
	assert.Equal(t, "value1", tagResponse["tag1"])

	response = collector.GetArrayType()
	assert.Equal(t, common.FlashArray, response)

	response = collector.GetDisplayName()
	assert.NotNil(t, "cinder-fa1", response)
}

func TestFlashArrayCollectorTagFetchError(t *testing.T) {
	t.Skip("Waiting for mock interceptor")

	logrus.SetLevel(logrus.TraceLevel)

	metaInterface := &mock.ArrayMetadataImpl{}
	metaInterface.On("GetTags", "000000000000000000000000").Return(map[string]string{}, fmt.Errorf("Some error"))

	collector, err := NewCollector("000000000000000000000000", "test-array", testArrayEndpoint2, testArrayToken2, metaInterface)
	assert.NoError(t, err)

	var response interface{}

	arrayDataResponse, err := collector.GetAllArrayData()
	assert.NoError(t, err)
	assert.NotNil(t, arrayDataResponse)
	assert.Len(t, arrayDataResponse.ArrayMetric.Tags, 0)

	volumeDataResponse, err := collector.GetAllVolumeData(30)
	assert.NoError(t, err)
	assert.NotNil(t, volumeDataResponse)
	if len(volumeDataResponse.VolumeMetricsTimeSeries) > 0 {
		// Test the first volume to see if it has the right tags
		firstVolume := volumeDataResponse.VolumeMetricsTimeSeries[0]
		assert.Len(t, firstVolume.ArrayTags, 0)
	}

	response = collector.GetArrayID()
	assert.Equal(t, "000000000000000000000000", response)

	response, err = collector.GetArrayModel()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = collector.GetArrayName()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	response, err = collector.GetArrayVersion()
	assert.NoError(t, err)
	assert.NotNil(t, response)

	_, err = collector.GetArrayTags()
	assert.Error(t, err)

	response = collector.GetArrayType()
	assert.Equal(t, common.FlashArray, response)

	response = collector.GetDisplayName()
	assert.NotNil(t, "cinder-fa1", response)
}
