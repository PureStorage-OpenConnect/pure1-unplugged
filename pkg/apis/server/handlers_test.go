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
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/http/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	clientmock "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/clients/mock"
)

const (
	emptyBulkResponse = "{\"response\":[]}\n"
)

var (
	emptyQuery = resources.GenerateEmptyQuery()
)

// assertJSONError checks that:
// 1. This matches the JSONErr format (and is valid JSON)
// 2. The JSONErr contains the proper error code
// 3. The JSONErr text is not nil and not empty/all whitespace
func assertJSONError(t *testing.T, body *bytes.Buffer, code int) {
	var parsed errors.JSONErr
	err := json.Unmarshal(body.Bytes(), &parsed)
	assert.NoError(t, err)
	assert.Equal(t, code, parsed.Code)
	assert.NotNil(t, parsed.Text)
	assert.NotEmpty(t, strings.TrimSpace(parsed.Text))
}

// assertError checks that:
// 1. The response code matches the given code
// 2. The body passes assertJSONError's conditions
func assertError(t *testing.T, response httptest.ResponseRecorder, code int) {
	assert.Equal(t, code, response.Code)
	assertJSONError(t, response.Body, code)
}

// assertMapContainsArrayKeys checks that:
// 1. All keys of a array result are contained
// 2. All keys are strings (including the two date/times, those are parsed at a higher level)
// 3. ID is not empty
func assertMapContainsArrayKeys(t *testing.T, body map[string]interface{}) {
	keys := []string{"id", "name", "status", "mgmt_endpoint", "device_type", "api_token", "model", "version", "_as_of", "_last_updated"}
	assert.Equal(t, len(keys), len(body))
	for _, key := range keys {
		assert.Contains(t, body, key)
		assert.IsType(t, "", body[key])
	}

	assert.NotEmpty(t, strings.TrimSpace(body["id"].(string)))
}

// assertMapContainsStatusKeys checks that:
// 1. All keys of a array status result are contained
// 2. All keys are strings
// 3. ID is not empty
func assertMapContainsStatusKeys(t *testing.T, body map[string]interface{}) {
	keys := []string{"id", "status", "_as_of"}
	assert.Equal(t, len(keys), len(body))
	for _, key := range keys {
		assert.Contains(t, body, key)
		assert.IsType(t, "", body[key])
	}

	assert.NotEmpty(t, strings.TrimSpace(body["id"].(string)))
}

// assertMapContainsTagsKeys checks that:
// 1. All keys of a array tags result are contained
// 2. All keys are strings with the exception of tags
// 3. ID is not empty
func assertMapContainsTagsKeys(t *testing.T, body map[string]interface{}) {
	keys := []string{"id", "tags", "_as_of"}
	assert.Equal(t, len(keys), len(body))
	for _, key := range keys {
		assert.Contains(t, body, key)
		if key == "tags" {
			assert.IsType(t, []interface{}{}, body[key])
		} else {
			assert.IsType(t, "", body[key])
		}
	}

	assert.NotEmpty(t, strings.TrimSpace(body["id"].(string)))
}

// assertBulkResponseMapContainsArrayKeys checks that:
// 1. The given body contains the key "response"
// 2. Every map inside of "response" fits the conditions of assertMapContainsArrayKeys
func assertBulkResponseMapContainsArrayKeys(t *testing.T, body map[string]interface{}) {
	assert.Contains(t, body, "response")
	response := body["response"].([]interface{})
	for _, item := range response {
		mapped := item.(map[string]interface{})
		assertMapContainsArrayKeys(t, mapped)
	}
}

// assertBulkResponseMapContainsStatusKeys checks that:
// 1. The given body contains the key "response"
// 2. Every map inside of "response" fits the conditions of assertMapContainsStatusKeys
func assertBulkResponseMapContainsStatusKeys(t *testing.T, body map[string]interface{}) {
	assert.Contains(t, body, "response")
	response := body["response"].([]interface{})
	for _, item := range response {
		mapped := item.(map[string]interface{})
		assertMapContainsStatusKeys(t, mapped)
	}
}

// assertBulkResponseMapContainsTagsKeys checks that:
// 1. The given body contains the key "response"
// 2. Every map inside of "response" fits the conditions of assertMapContainsTagsKeys
func assertBulkResponseMapContainsTagsKeys(t *testing.T, body map[string]interface{}) {
	assert.Contains(t, body, "response")
	response := body["response"].([]interface{})
	for _, item := range response {
		mapped := item.(map[string]interface{})
		assertMapContainsTagsKeys(t, mapped)
	}
}

// parseBody check that:
// 1. The response code is 200 OK
// 2. The body is valid JSON
// and returns the parsed JSON
func parseBody(t *testing.T, response httptest.ResponseRecorder) map[string]interface{} {
	assert.Equal(t, http.StatusOK, response.Code)
	var parsed map[string]interface{}
	err := json.Unmarshal(response.Body.Bytes(), &parsed)
	assert.NoError(t, err)
	return parsed
}

func TestGetArray(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("GET", "/api-server/arrays", nil)

	getArrays(&recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, emptyBulkResponse, recorder.Body.String())
}

func TestGetArrayQuery(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}
	connection.DAO = &mockDAO
	connection.Tokens = &tokenStorage

	query := resources.GenerateEmptyQuery()
	query.Ids = []string{"000000000000000000000000"}
	mockDAO.On("FindArrays", &query).Return([]*resources.Array{
		&resources.Array{InternalID: "000000000000000000000000"},
	}, nil)
	tokenStorage.On("GetToken", "000000000000000000000000").Return("test-token", nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("GET", "/api-server/arrays?ids=000000000000000000000000", nil)

	getArrays(&recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	body := parseBody(t, recorder)
	assertBulkResponseMapContainsArrayKeys(t, body)

	response := body["response"].([]interface{})
	assert.Len(t, response, 1)

	array1 := response[0].(map[string]interface{})
	assert.Equal(t, "000000000000000000000000", array1["id"])
	assert.Equal(t, "test-token", array1["api_token"])
}

func TestGetArrayBadQuery(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("GET", "/api-server/arrays?ids=a", nil)

	getArrays(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestGetArrayInternalError(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, errors.MakeInternalHTTPErr(fmt.Errorf("Some error")))

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("GET", "/api-server/arrays", nil)

	getArrays(&recorder, req)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestPostArrayValidFlashArray(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}
	connection.DAO = &mockDAO
	connection.Tokens = &tokenStorage

	name := "test_array1"
	mgmtEndpoint := "192.168.99.100"
	deviceType := common.FlashArray
	apiToken := "asdf"

	mockDAO.On("InsertArray", mock.Anything).Return(nil)
	tokenStorage.On("SaveToken", mock.AnythingOfType("string"), apiToken).Return(nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("POST", "/api-server/arrays", strings.NewReader(fmt.Sprintf(`{
	"name": "%s",
	"mgmt_endpoint": "%s",
	"device_type": "%s",
	"api_token": "%s"
}`, name, mgmtEndpoint, deviceType, apiToken)))

	postArray(&recorder, req)
	body := parseBody(t, recorder)
	assertMapContainsArrayKeys(t, body)
	assert.Equal(t, name, body["name"])
	assert.Equal(t, mgmtEndpoint, body["mgmt_endpoint"])
	assert.Equal(t, deviceType, body["device_type"])
	assert.Equal(t, apiToken, body["api_token"])
}

func TestPostArrayInvalidJSON(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("POST", "/api-server/arrays", strings.NewReader("a"))

	postArray(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestPostArrayInvalidFlashArray(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	mockDAO.On("InsertArray", mock.Anything).Return(nil)

	name := "test_array1"
	mgmtEndpoint := "192.168.99.100"
	deviceType := common.FlashArray

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("POST", "/api-server/arrays", strings.NewReader(fmt.Sprintf(`{
	"name": "%s",
	"mgmt_endpoint": "%s",
	"device_type": "%s"
}`, name, mgmtEndpoint, deviceType))) // Missing API token (for example)

	postArray(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestPostArrayInternalError(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}
	connection.DAO = &mockDAO
	connection.Tokens = &tokenStorage

	name := "test_array1"
	mgmtEndpoint := "192.168.99.100"
	deviceType := common.FlashArray
	apiToken := "asdf"

	mockDAO.On("InsertArray", mock.Anything).Return(errors.MakeInternalHTTPErr(fmt.Errorf("Some error")))
	tokenStorage.On("SaveToken", mock.AnythingOfType("string"), apiToken).Return(nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("POST", "/api-server/arrays", strings.NewReader(fmt.Sprintf(`{
	"name": "%s",
	"mgmt_endpoint": "%s",
	"device_type": "%s",
	"api_token": "%s"
}`, name, mgmtEndpoint, deviceType, apiToken)))

	postArray(&recorder, req)
	assertError(t, recorder, http.StatusInternalServerError)
}

func TestPatchArrayValidFlashArray(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}
	connection.DAO = &mockDAO
	connection.Tokens = &tokenStorage

	id := "000000000000000000000000"
	originalName := "test_array1"
	newName := "new_array1"
	mgmtEndpoint := "192.168.99.100"
	deviceType := common.FlashArray
	apiToken := "asdf"

	originalArray := resources.Array{
		InternalID:   id,
		Name:         originalName,
		MgmtEndPoint: mgmtEndpoint,
		DeviceType:   deviceType,
		APIToken:     apiToken,
	}

	patchedArray := resources.Array{
		InternalID:   id,
		Name:         newName,
		MgmtEndPoint: mgmtEndpoint,
		DeviceType:   deviceType,
		APIToken:     apiToken,
	}

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{&originalArray}, nil)
	mockDAO.On("PatchArray", mock.Anything).Return(&patchedArray, nil)
	tokenStorage.On("SaveToken", id, apiToken).Return(nil) // Technically we don't need to save, but it can't hurt
	tokenStorage.On("GetToken", id).Return(apiToken, nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays", strings.NewReader(fmt.Sprintf(`{
	"name": "%s"
}`, newName)))

	patchArray(&recorder, req)
	body := parseBody(t, recorder)
	assertBulkResponseMapContainsArrayKeys(t, body)
	firstResult := body["response"].([]interface{})[0].(map[string]interface{})
	assert.Equal(t, newName, firstResult["name"])
	assert.Equal(t, mgmtEndpoint, firstResult["mgmt_endpoint"])
	assert.Equal(t, deviceType, firstResult["device_type"])
	assert.Equal(t, apiToken, firstResult["api_token"])
}

func TestPatchArrayInvalidQuery(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	newName := "new_array1"

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays?ids=a", strings.NewReader(fmt.Sprintf(`{
	"name": "%s"
}`, newName)))

	patchArray(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestPatchArrayInvalidJSON(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays", strings.NewReader("a"))

	patchArray(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestPatchArrayInternalFindError(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	newName := "new_array1"

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, errors.MakeInternalHTTPErr(fmt.Errorf("Some error")))

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays", strings.NewReader(fmt.Sprintf(`{
	"name": "%s"
}`, newName)))

	patchArray(&recorder, req)
	assertError(t, recorder, http.StatusInternalServerError)
}

func TestPatchArrayInternalPatchError(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}
	connection.DAO = &mockDAO
	connection.Tokens = &tokenStorage

	id := "000000000000000000000000"
	originalName := "test_array1"
	newName := "new_array1"
	mgmtEndpoint := "192.168.99.100"
	deviceType := common.FlashArray
	apiToken := "asdf"

	originalArray := resources.Array{
		InternalID:   id,
		Name:         originalName,
		MgmtEndPoint: mgmtEndpoint,
		DeviceType:   deviceType,
		APIToken:     apiToken,
	}

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{&originalArray}, nil)
	mockDAO.On("PatchArray", mock.Anything).Return(&resources.Array{}, errors.MakeInternalHTTPErr(fmt.Errorf("Some error")))
	tokenStorage.On("SaveToken", id, apiToken).Return(nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays", strings.NewReader(fmt.Sprintf(`{
	"name": "%s"
}`, newName)))

	patchArray(&recorder, req)
	assertError(t, recorder, http.StatusInternalServerError)
}

func TestDeleteArray(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	tokenStorage := clientmock.APITokenStorageImpl{}
	connection.DAO = &mockDAO
	connection.Tokens = &tokenStorage

	query := resources.GenerateEmptyQuery()
	query.Ids = []string{"000000000000000000000000"}

	tokenStorage.On("DeleteToken", "000000000000000000000000").Return(nil)

	mockDAO.On("DeleteArray", &query).Return(query.Ids, nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("DELETE", "/api-server/arrays?ids=000000000000000000000000", nil)

	deleteArray(&recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.NotEmpty(t, strings.TrimSpace(recorder.Body.String()))
}

func TestDeleteArrayInvalidQuery(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("DELETE", "/api-server/arrays?ids=a", nil)

	deleteArray(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestDeleteArrayEmptyQuery(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("DELETE", "/api-server/arrays", nil)

	deleteArray(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestDeleteArrayInternalError(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	query := resources.GenerateEmptyQuery()
	query.Names = []string{"asdf"}

	mockDAO.On("DeleteArray", &query).Return([]string{}, errors.MakeInternalHTTPErr(fmt.Errorf("Some error")))

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("DELETE", "/api-server/arrays?names=asdf", nil)

	deleteArray(&recorder, req)
	assertError(t, recorder, http.StatusInternalServerError)
}

func TestGetArrayStatus(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("GET", "/api-server/arrays/status", nil)

	getArrayStatus(&recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, emptyBulkResponse, recorder.Body.String())
}

func TestGetArrayStatusQuery(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	query := resources.GenerateEmptyQuery()
	query.Ids = []string{"000000000000000000000000"}
	mockDAO.On("FindArrays", &query).Return([]*resources.Array{
		&resources.Array{InternalID: "000000000000000000000000"},
	}, nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("GET", "/api-server/arrays/status?ids=000000000000000000000000", nil)

	getArrayStatus(&recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	body := parseBody(t, recorder)
	assertBulkResponseMapContainsStatusKeys(t, body)
}

func TestGetArrayStatusBadQuery(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("GET", "/api-server/arrays/status?ids=a", nil)

	getArrayStatus(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestGetArrayStatusInternalError(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, errors.MakeInternalHTTPErr(fmt.Errorf("Some error")))

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("GET", "/api-server/arrays/status", nil)

	getArrayStatus(&recorder, req)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestGetArrayTags(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("GET", "/api-server/arrays/tags", nil)

	getArrayTags(&recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, emptyBulkResponse, recorder.Body.String())
}

func TestGetArrayTagsQuery(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	query := resources.GenerateEmptyQuery()
	query.Ids = []string{"000000000000000000000000"}
	mockDAO.On("FindArrays", &query).Return([]*resources.Array{
		&resources.Array{InternalID: "000000000000000000000000"},
	}, nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("GET", "/api-server/arrays/tags?ids=000000000000000000000000", nil)

	getArrayTags(&recorder, req)
	assert.Equal(t, http.StatusOK, recorder.Code)
	body := parseBody(t, recorder)
	assertBulkResponseMapContainsTagsKeys(t, body)
}

func TestGetArrayTagsBadQuery(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("GET", "/api-server/arrays/tags?ids=a", nil)

	getArrayTags(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestGetArrayTagsInternalError(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, errors.MakeInternalHTTPErr(fmt.Errorf("Some error")))

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("GET", "/api-server/arrays/tags", nil)

	getArrayTags(&recorder, req)
	assert.Equal(t, http.StatusInternalServerError, recorder.Code)
}

func TestPatchArrayTags(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	tagKey := "somekey"
	tagValue := "somevalue"
	tagNS := "somenamespace"

	originalArray := resources.Array{
		Name:         "test-array1",
		MgmtEndPoint: "192.168.99.100",
		DeviceType:   common.FlashArray,
		APIToken:     "asdf",
		Tags:         []map[string]string{},
	}

	patchedArray := resources.Array{
		Name:         "test-array1",
		InternalID:   "000000000000000000000000",
		MgmtEndPoint: "192.168.99.100",
		DeviceType:   common.FlashArray,
		APIToken:     "asdf",
		Tags: []map[string]string{
			map[string]string{
				"key":       tagKey,
				"value":     tagValue,
				"namespace": tagNS,
			},
		},
	}

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{&originalArray}, nil)
	mockDAO.On("PatchArrayTags", mock.Anything).Return(&patchedArray, nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays/tags", strings.NewReader(fmt.Sprintf(`{
		"tags": [
			{
				"key": "%s",
				"value": "%s",
				"namespace": "%s"
			}
		]
}`, tagKey, tagValue, tagNS)))

	patchArrayTags(&recorder, req)
	body := parseBody(t, recorder)
	assertBulkResponseMapContainsTagsKeys(t, body)
}

func TestPatchArrayTagsQuery(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	tagKey := "somekey"
	tagValue := "somevalue"
	tagNS := "somenamespace"

	originalArray := resources.Array{
		Name:         "test-array1",
		MgmtEndPoint: "192.168.99.100",
		DeviceType:   common.FlashArray,
		APIToken:     "asdf",
		Tags:         []map[string]string{},
	}

	patchedArray := resources.Array{
		Name:         "test-array1",
		InternalID:   "000000000000000000000000",
		MgmtEndPoint: "192.168.99.100",
		DeviceType:   common.FlashArray,
		APIToken:     "asdf",
		Tags: []map[string]string{
			map[string]string{
				"key":       tagKey,
				"value":     tagValue,
				"namespace": tagNS,
			},
		},
	}

	query := resources.GenerateEmptyQuery()
	query.Ids = []string{"000000000000000000000000"}

	mockDAO.On("FindArrays", &query).Return([]*resources.Array{&originalArray}, nil)
	mockDAO.On("PatchArrayTags", mock.Anything).Return(&patchedArray, nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays/tags?ids=000000000000000000000000", strings.NewReader(fmt.Sprintf(`{
		"tags": [
			{
				"key": "%s",
				"value": "%s",
				"namespace": "%s"
			}
		]
}`, tagKey, tagValue, tagNS)))

	patchArrayTags(&recorder, req)
	body := parseBody(t, recorder)
	assertBulkResponseMapContainsTagsKeys(t, body)
}

func TestPatchArrayTagsInvalidQuery(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	tagKey := "somekey"
	tagValue := "somevalue"
	tagNS := "somenamespace"

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays/tags?ids=a", strings.NewReader(fmt.Sprintf(`{
		"tags": [
			{
				"key": "%s",
				"value": "%s",
				"namespace": "%s"
			}
		]
}`, tagKey, tagValue, tagNS)))

	patchArrayTags(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestPatchArrayTagsInvalidJSON(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays/tags", strings.NewReader(`a`))

	patchArrayTags(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestPatchArrayTagsMissingTags(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays/tags", strings.NewReader("{}"))

	patchArrayTags(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestPatchArrayTagsNotMap(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays/tags", strings.NewReader(`{
		"tags": [
			"notamap"
		]
}`))

	patchArrayTags(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestPatchArrayTagsNotStringValue(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays/tags", strings.NewReader(`{
		"tags": [
			{
				"key": 0
			}
		]
}`))

	patchArrayTags(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestPatchArrayTagsFindError(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	tagKey := "somekey"
	tagValue := "somevalue"
	tagNS := "somenamespace"

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, errors.MakeInternalHTTPErr(fmt.Errorf("Some error")))

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays/tags", strings.NewReader(fmt.Sprintf(`{
		"tags": [
			{
				"key": "%s",
				"value": "%s",
				"namespace": "%s"
			}
		]
}`, tagKey, tagValue, tagNS)))

	patchArrayTags(&recorder, req)
	assertError(t, recorder, http.StatusInternalServerError)
}

func TestPatchArrayTagsPatchError(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	tagKey := "somekey"
	tagValue := "somevalue"
	tagNS := "somenamespace"

	originalArray := resources.Array{
		Name:         "test-array1",
		MgmtEndPoint: "192.168.99.100",
		DeviceType:   common.FlashArray,
		APIToken:     "asdf",
		Tags:         []map[string]string{},
	}

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{&originalArray}, nil)
	mockDAO.On("PatchArrayTags", mock.Anything).Return(&resources.Array{}, errors.MakeInternalHTTPErr(fmt.Errorf("Some error")))

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("PATCH", "/api-server/arrays/tags", strings.NewReader(fmt.Sprintf(`{
		"tags": [
			{
				"key": "%s",
				"value": "%s",
				"namespace": "%s"
			}
		]
}`, tagKey, tagValue, tagNS)))

	patchArrayTags(&recorder, req)
	assertError(t, recorder, http.StatusInternalServerError)
}

func TestDeleteArrayTags(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	tagKey := "somekey"
	tagValue := "somevalue"
	tagNS := "somenamespace"

	originalArray := resources.Array{
		Name:         "test-array1",
		MgmtEndPoint: "192.168.99.100",
		DeviceType:   common.FlashArray,
		APIToken:     "asdf",
		Tags: []map[string]string{
			map[string]string{
				"key":       tagKey,
				"value":     tagValue,
				"namespace": tagNS,
			},
		},
	}

	patchedArray := resources.Array{
		Name:         "test-array1",
		InternalID:   "000000000000000000000000",
		MgmtEndPoint: "192.168.99.100",
		DeviceType:   common.FlashArray,
		APIToken:     "asdf",
		Tags:         []map[string]string{},
	}

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{&originalArray}, nil)
	mockDAO.On("PatchArrayTags", mock.Anything).Return(&patchedArray, nil)

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api-server/arrays/tags?tags=%s", tagKey), nil)

	deleteArrayTags(&recorder, req)
	body := parseBody(t, recorder)
	assertBulkResponseMapContainsTagsKeys(t, body)
}

func TestDeleteArrayTagsInvalidQuery(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	tagKey := "somekey"

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api-server/arrays/tags?ids=a&tags=%s", tagKey), nil)

	deleteArrayTags(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestDeleteArrayTagsMissingTags(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("DELETE", "/api-server/arrays/tags", nil)

	deleteArrayTags(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestDeleteArrayTagsEmptyTags(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("DELETE", "/api-server/arrays/tags?tags=", nil)

	deleteArrayTags(&recorder, req)
	assertError(t, recorder, http.StatusBadRequest)
}

func TestDeleteArrayTagsFindError(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	tagKey := "somekey"

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{}, errors.MakeInternalHTTPErr(fmt.Errorf("Some error")))

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api-server/arrays/tags?tags=%s", tagKey), nil)

	deleteArrayTags(&recorder, req)
	assertError(t, recorder, http.StatusInternalServerError)
}

func TestDeleteArrayTagsPatchError(t *testing.T) {
	mockDAO := clientmock.ArrayDatabaseImpl{}
	connection.DAO = &mockDAO

	tagKey := "somekey"
	tagValue := "somevalue"
	tagNS := "somenamespace"

	originalArray := resources.Array{
		Name:         "test-array1",
		MgmtEndPoint: "192.168.99.100",
		DeviceType:   common.FlashArray,
		APIToken:     "asdf",
		Tags: []map[string]string{
			map[string]string{
				"key":       tagKey,
				"value":     tagValue,
				"namespace": tagNS,
			},
		},
	}

	mockDAO.On("FindArrays", &emptyQuery).Return([]*resources.Array{&originalArray}, nil)
	mockDAO.On("PatchArrayTags", mock.Anything).Return(&resources.Array{}, errors.MakeInternalHTTPErr(fmt.Errorf("Some error")))

	recorder := httptest.ResponseRecorder{Body: &bytes.Buffer{}}
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/api-server/arrays/tags?tags=%s", tagKey), nil)

	deleteArrayTags(&recorder, req)
	assertError(t, recorder, http.StatusInternalServerError)
}
