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
	"fmt"
	"testing"
	"time"

	clientmock "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/mock"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	emptyQuery = resources.GenerateEmptyQuery()
)

func TestGetArrays(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}

	handler := MetadataConnection{DAO: &mockImpl, Tokens: &tokenStorage}

	lastSeenTime := time.Unix(1000, 0).UTC()

	arrays := []*resources.Array{
		&resources.Array{InternalID: "aaaa", Name: "test_dev1", Status: "Connected", Tags: []map[string]string{
			map[string]string{
				"key":       "test_key",
				"value":     "test_value",
				"namespace": "test_ns",
			},
		}, Lastseen: lastSeenTime},
		&resources.Array{InternalID: "aaab", Name: "test_dev2", Status: "Connected", Tags: []map[string]string{
			map[string]string{
				"key":       "test_key",
				"value":     "test_value",
				"namespace": "test_ns",
			},
		}, Lastseen: lastSeenTime},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(arrays, nil)
	tokenStorage.On("GetToken", "aaaa").Return("test-token", nil)
	tokenStorage.On("GetToken", "aaab").Return("test-token2", nil)

	res, err := handler.GetArrays(emptyQuery)
	assert.NoError(t, err)
	assert.Equal(t, "aaaa", res.Response[0]["id"]) // Should contain ID
	assert.Equal(t, "aaab", res.Response[1]["id"])
	assert.Equal(t, "test_dev1", res.Response[0]["name"]) // Should contain name
	assert.Equal(t, "test_dev2", res.Response[1]["name"])
	assert.Equal(t, "Connected", res.Response[0]["status"]) // Should contain status
	assert.Equal(t, "Connected", res.Response[1]["status"])
	assert.Equal(t, lastSeenTime, res.Response[0]["_as_of"]) // Should contain _as_of
	assert.Equal(t, lastSeenTime, res.Response[1]["_as_of"])
	assert.Equal(t, "test-token", res.Response[0]["api_token"]) // Should contain api_token
	assert.Equal(t, "test-token2", res.Response[1]["api_token"])
	assert.NotContains(t, res.Response[0], "tags") // Should not contain tags
	assert.NotContains(t, res.Response[1], "tags")
}

func TestGetArraysError(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	mockImpl.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, fmt.Errorf("Some error"))

	_, err := handler.GetArrays(emptyQuery)
	assert.Error(t, err)
}

func TestGetArrayStatuses(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	lastSeenTime := time.Unix(1000, 0).UTC()

	arrays := []*resources.Array{
		&resources.Array{InternalID: "aaaa", Name: "test_dev1", Status: "Connected", Tags: []map[string]string{
			map[string]string{
				"key":       "test_key",
				"value":     "test_value",
				"namespace": "test_ns",
			},
		}, Lastseen: lastSeenTime},
		&resources.Array{InternalID: "aaab", Name: "test_dev2", Status: "Connected", Tags: []map[string]string{
			map[string]string{
				"key":       "test_key",
				"value":     "test_value",
				"namespace": "test_ns",
			},
		}, Lastseen: lastSeenTime},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(arrays, nil)

	res, err := handler.GetArrayStatuses(emptyQuery)
	assert.NoError(t, err)
	assert.Equal(t, "aaaa", res.Response[0]["id"]) // Should contain ID
	assert.Equal(t, "aaab", res.Response[1]["id"])
	assert.NotContains(t, res.Response[0], "name") // Should not contain name (or anything aside from status, ID, and _as_of)
	assert.NotContains(t, res.Response[1], "name")
	assert.Equal(t, "Connected", res.Response[0]["status"]) // Should contain status
	assert.Equal(t, "Connected", res.Response[1]["status"])
	assert.Equal(t, lastSeenTime, res.Response[0]["_as_of"]) // Should contain _as_of
	assert.Equal(t, lastSeenTime, res.Response[1]["_as_of"])
	assert.NotContains(t, res.Response[0], "tags") // Should not contain tags
	assert.NotContains(t, res.Response[1], "tags")
}

func TestGetArrayStatusesError(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	mockImpl.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, fmt.Errorf("Some error"))

	_, err := handler.GetArrayStatuses(emptyQuery)
	assert.Error(t, err)
}

func TestGetArrayTags(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	lastSeenTime := time.Unix(1000, 0).UTC()

	tags := []map[string]string{
		map[string]string{
			"key":       "test_key",
			"value":     "test_value",
			"namespace": "test_ns",
		},
	}

	arrays := []*resources.Array{
		&resources.Array{InternalID: "aaaa", Name: "test_dev1", Status: "Connected", Tags: tags, Lastseen: lastSeenTime},
		&resources.Array{InternalID: "aaab", Name: "test_dev2", Status: "Connected", Tags: tags, Lastseen: lastSeenTime},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(arrays, nil)

	res, err := handler.GetArrayTags(emptyQuery)
	assert.NoError(t, err)
	assert.Equal(t, "aaaa", res.Response[0]["id"]) // Should contain ID
	assert.Equal(t, "aaab", res.Response[1]["id"])
	assert.NotContains(t, res.Response[0], "name") // Should not contain name (or anything aside from tags, ID, and _as_of)
	assert.NotContains(t, res.Response[1], "name")
	assert.NotContains(t, res.Response[0], "status") // Should not contain status
	assert.NotContains(t, res.Response[1], "status")
	assert.Equal(t, lastSeenTime, res.Response[0]["_as_of"]) // Should contain _as_of
	assert.Equal(t, lastSeenTime, res.Response[1]["_as_of"])
	assert.Equal(t, tags, res.Response[0]["tags"]) // Should contain tags
	assert.Equal(t, tags, res.Response[1]["tags"])
}

func TestGetArrayTagsNil(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	lastSeenTime := time.Unix(1000, 0).UTC()

	arrays := []*resources.Array{
		&resources.Array{InternalID: "aaaa", Name: "test_dev1", Status: "Connected", Tags: nil, Lastseen: lastSeenTime},
		&resources.Array{InternalID: "aaab", Name: "test_dev2", Status: "Connected", Tags: nil, Lastseen: lastSeenTime},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(arrays, nil)

	res, err := handler.GetArrayTags(emptyQuery)
	assert.NoError(t, err)
	assert.Equal(t, "aaaa", res.Response[0]["id"]) // Should contain ID
	assert.Equal(t, "aaab", res.Response[1]["id"])
	assert.NotContains(t, res.Response[0], "name") // Should not contain name (or anything aside from tags, ID, and _as_of)
	assert.NotContains(t, res.Response[1], "name")
	assert.NotContains(t, res.Response[0], "status") // Should not contain status
	assert.NotContains(t, res.Response[1], "status")
	assert.Equal(t, lastSeenTime, res.Response[0]["_as_of"]) // Should contain _as_of
	assert.Equal(t, lastSeenTime, res.Response[1]["_as_of"])
	assert.Equal(t, []map[string]string{}, res.Response[0]["tags"]) // Should contain tags
	assert.Equal(t, []map[string]string{}, res.Response[1]["tags"])
}

func TestGetArrayTagsError(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	mockImpl.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, fmt.Errorf("Some error"))

	_, err := handler.GetArrayTags(emptyQuery)
	assert.Error(t, err)
}

func TestPostArray(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}

	handler := MetadataConnection{DAO: &mockImpl, Tokens: &tokenStorage}

	mockImpl.On("InsertArray", mock.AnythingOfType("*resources.Array")).Return(nil)
	tokenStorage.On("SaveToken", mock.AnythingOfType("string"), "asdf").Return(nil)

	res, err := handler.PostArray(map[string]interface{}{
		"name":          "test_dev1",
		"mgmt_endpoint": "192.168.99.100",
		"device_type":   common.FlashArray,
		"api_token":     "asdf",
	})
	assert.NoError(t, err)
	assert.Equal(t, "test_dev1", res["name"])
	assert.Equal(t, "192.168.99.100", res["mgmt_endpoint"])
	assert.Equal(t, common.FlashArray, res["device_type"])
	assert.Equal(t, "asdf", res["api_token"])
	assert.Equal(t, time.Time{}, res["_as_of"])
	assert.NotEqual(t, time.Time{}, res["_last_updated"])
}

func TestPostArrayBadParse(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	mockImpl.On("InsertArray", mock.AnythingOfType("map[string]interface{}")).Return(nil)

	_, err := handler.PostArray(map[string]interface{}{
		"id":            "asdf", // This shouldn't ever be set in a real request (although it could happen)
		"name":          "test_dev1",
		"mgmt_endpoint": "192.168.99.100",
		"device_type":   common.FlashArray,
		"api_token":     "asdf",
	})
	assert.Error(t, err)
}

func TestPostArrayMissingField(t *testing.T) {
	handler := MetadataConnection{}

	_, err := handler.PostArray(map[string]interface{}{
		"name":          "test_dev1",
		"mgmt_endpoint": "192.168.99.100",
		"device_type":   "  ",
		"api_token":     "asdf",
	})
	assert.Error(t, err)
}

func TestPostArrayInsertError(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}

	handler := MetadataConnection{DAO: &mockImpl, Tokens: &tokenStorage}

	tokenStorage.On("SaveToken", mock.AnythingOfType("string"), "asdf").Return(nil)
	mockImpl.On("InsertArray", mock.AnythingOfType("*resources.Array")).Return(fmt.Errorf("Some error"))

	_, err := handler.PostArray(map[string]interface{}{
		"name":          "test_dev1",
		"mgmt_endpoint": "192.168.99.100",
		"device_type":   common.FlashArray,
		"api_token":     "asdf",
	})
	assert.Error(t, err)
}

func TestPatchArray(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}

	handler := MetadataConnection{DAO: &mockImpl, Tokens: &tokenStorage}

	fetchedArrays := []*resources.Array{
		&resources.Array{
			InternalID:   "aaaa",
			Name:         "test_dev1",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
		},
	}
	patchedArrays := []*resources.Array{
		&resources.Array{
			InternalID:   "aaaa",
			Name:         "NEWNAME",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
		},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(fetchedArrays, nil)

	mockImpl.On("PatchArray", mock.AnythingOfType("*resources.Array")).Return(patchedArrays[0], nil)
	tokenStorage.On("SaveToken", "aaaa", "asdf").Return(nil)
	tokenStorage.On("GetToken", "aaaa").Return("asdf", nil)

	res, err := handler.PatchArrays(emptyQuery, map[string]interface{}{
		"name": "NEWNAME",
	})

	assert.NoError(t, err)
	assert.Equal(t, "NEWNAME", res.Response[0]["name"])
	assert.Equal(t, "192.168.99.100", res.Response[0]["mgmt_endpoint"])
	assert.Equal(t, common.FlashArray, res.Response[0]["device_type"])
	assert.Equal(t, "asdf", res.Response[0]["api_token"])
}

func TestPatchArrayFindError(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	mockImpl.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, fmt.Errorf("Some error"))

	_, err := handler.PatchArrays(emptyQuery, map[string]interface{}{
		"name": "NEWNAME",
	})

	assert.Error(t, err)
}

func TestPatchArrayErrorPatching(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}

	handler := MetadataConnection{DAO: &mockImpl, Tokens: &tokenStorage}

	fetchedArrays := []*resources.Array{
		&resources.Array{
			Name:         "test_dev1",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
		},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(fetchedArrays, nil)

	_, err := handler.PatchArrays(emptyQuery, map[string]interface{}{
		"api_token": "  ",
	})

	assert.Error(t, err)
}

func TestPatchArrayPushError(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}

	handler := MetadataConnection{DAO: &mockImpl, Tokens: &tokenStorage}

	fetchedArrays := []*resources.Array{
		&resources.Array{
			InternalID:   "aaaa",
			Name:         "test_dev1",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
		},
		&resources.Array{
			InternalID:   "aaab",
			Name:         "test_dev2",
			MgmtEndPoint: "192.168.99.101",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
		},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(fetchedArrays, nil)
	mockImpl.On("PatchArray", mock.AnythingOfType("*resources.Array")).Return(&resources.Array{}, fmt.Errorf("Some error"))
	tokenStorage.On("SaveToken", "aaaa", "asdf").Return(nil)

	_, err := handler.PatchArrays(emptyQuery, map[string]interface{}{
		"name": "NEWNAME",
	})

	assert.Error(t, err)
}

func TestPatchArrayTokenSaveError(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}

	handler := MetadataConnection{DAO: &mockImpl, Tokens: &tokenStorage}

	fetchedArrays := []*resources.Array{
		&resources.Array{
			InternalID:   "aaaa",
			Name:         "test_dev1",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
		},
	}
	patchedArrays := []*resources.Array{
		&resources.Array{
			InternalID:   "aaaa",
			Name:         "NEWNAME",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
		},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(fetchedArrays, nil)

	mockImpl.On("PatchArray", mock.AnythingOfType("*resources.Array")).Return(patchedArrays[0], nil)
	tokenStorage.On("SaveToken", "aaaa", "asdf").Return(fmt.Errorf("Some error"))

	_, err := handler.PatchArrays(emptyQuery, map[string]interface{}{
		"name": "NEWNAME",
	})

	assert.Error(t, err)
}

func TestPatchArrayTokenGetError(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}

	handler := MetadataConnection{DAO: &mockImpl, Tokens: &tokenStorage}

	fetchedArrays := []*resources.Array{
		&resources.Array{
			InternalID:   "aaaa",
			Name:         "test_dev1",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
		},
	}
	patchedArrays := []*resources.Array{
		&resources.Array{
			InternalID:   "aaaa",
			Name:         "NEWNAME",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
		},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(fetchedArrays, nil)

	mockImpl.On("PatchArray", mock.AnythingOfType("*resources.Array")).Return(patchedArrays[0], nil)
	tokenStorage.On("SaveToken", "aaaa", "asdf").Return(nil)
	tokenStorage.On("GetToken", "aaaa").Return("", fmt.Errorf("Some error"))

	_, err := handler.PatchArrays(emptyQuery, map[string]interface{}{
		"name": "NEWNAME",
	})

	assert.Error(t, err)
}

func TestPatchArrayTags(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	tagPatch := []map[string]string{
		map[string]string{
			"key":       "some_key",
			"value":     "some_value",
			"namespace": "some_namespace",
		},
	}

	fetchedArrays := []*resources.Array{
		&resources.Array{
			Name:         "test_dev1",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
		},
	}
	patchedArrays := []*resources.Array{
		&resources.Array{
			Name:         "test_dev1",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
			Tags:         tagPatch,
		},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(fetchedArrays, nil)
	mockImpl.On("PatchArrayTags", mock.AnythingOfType("*resources.Array")).Return(patchedArrays[0], nil)

	res, err := handler.PatchArrayTags(emptyQuery, tagPatch)
	assert.NoError(t, err)
	assert.Equal(t, tagPatch, res.Response[0]["tags"])
}

func TestPatchArrayTagsFindError(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	tagPatch := []map[string]string{
		map[string]string{
			"key":       "some_key",
			"value":     "some_value",
			"namespace": "some_namespace",
		},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, fmt.Errorf("Some error"))

	_, err := handler.PatchArrayTags(emptyQuery, tagPatch)
	assert.Error(t, err)
}

func TestPatchArrayTagsPatchError(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	tagPatch := []map[string]string{
		map[string]string{
			"key": "some_key",
			// Missing value
			"namespace": "some_namespace",
		},
	}

	fetchedArrays := []*resources.Array{
		&resources.Array{
			Name:         "test_dev1",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
		},
		&resources.Array{
			Name:         "test_dev2",
			MgmtEndPoint: "192.168.99.101",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
		},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(fetchedArrays, nil)

	_, err := handler.PatchArrayTags(emptyQuery, tagPatch)
	assert.Error(t, err)
}

func TestPatchArrayTagsPushError(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	tagPatch := []map[string]string{
		map[string]string{
			"key":       "some_key",
			"value":     "some_value",
			"namespace": "some_namespace",
		},
	}

	fetchedArrays := []*resources.Array{
		&resources.Array{
			Name:         "test_dev1",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
		},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(fetchedArrays, nil)
	mockImpl.On("PatchArrayTags", mock.AnythingOfType("*resources.Array")).Return(&resources.Array{}, fmt.Errorf("Some error"))

	_, err := handler.PatchArrayTags(emptyQuery, tagPatch)
	assert.Error(t, err)
}

func TestDeleteArrays(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}

	handler := MetadataConnection{DAO: &mockImpl, Tokens: &tokenStorage}

	mockImpl.On("DeleteArray", &emptyQuery).Return([]string{"dev1"}, nil)
	tokenStorage.On("DeleteToken", "dev1").Return(nil)

	count, err := handler.DeleteArrays(emptyQuery)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestDeleteArrayTags(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	fetchedArrays := []*resources.Array{
		&resources.Array{
			Name:         "test_dev1",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
			Tags: []map[string]string{
				map[string]string{
					"key":       "test_key",
					"value":     "test_value",
					"namespace": "test_ns",
				},
			},
		},
	}
	patchedArrays := []*resources.Array{
		&resources.Array{
			Name:         "test_dev1",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
			Tags:         []map[string]string{},
		},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(fetchedArrays, nil)
	mockImpl.On("PatchArrayTags", mock.AnythingOfType("*resources.Array")).Return(patchedArrays[0], nil)

	res, err := handler.DeleteArrayTags(emptyQuery, []string{"test_key"})
	assert.NoError(t, err)
	assert.Empty(t, res.Response[0]["tags"])
}

func TestDeleteArrayTagsFindError(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	mockImpl.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, fmt.Errorf("Some error"))

	_, err := handler.DeleteArrayTags(emptyQuery, []string{"test_key"})
	assert.Error(t, err)
}

func TestDeleteArrayTagsPatchError(t *testing.T) {
	mockImpl := clientmock.ArrayDatabaseImpl{}

	handler := MetadataConnection{DAO: &mockImpl}

	fetchedArrays := []*resources.Array{
		&resources.Array{
			Name:         "test_dev1",
			MgmtEndPoint: "192.168.99.100",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
			Tags: []map[string]string{
				map[string]string{
					"key":       "test_key",
					"value":     "test_value",
					"namespace": "test_ns",
				},
			},
		},
		&resources.Array{
			Name:         "test_dev2",
			MgmtEndPoint: "192.168.99.101",
			DeviceType:   common.FlashArray,
			APIToken:     "asdf",
			Tags: []map[string]string{
				map[string]string{
					"key":       "test_key",
					"value":     "test_value",
					"namespace": "test_ns",
				},
			},
		},
	}

	mockImpl.On("FindArrays", &emptyQuery).Return(fetchedArrays, nil)
	mockImpl.On("PatchArrayTags", mock.AnythingOfType("*resources.Array")).Return(&resources.Array{}, fmt.Errorf("Some error"))

	_, err := handler.DeleteArrayTags(emptyQuery, []string{"test_key"})
	assert.Error(t, err)
}
