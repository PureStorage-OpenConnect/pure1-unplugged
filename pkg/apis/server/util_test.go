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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/http/errors"
	"github.com/stretchr/testify/assert"
)

func TestParseRequestQueryParamsEmpty(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server", nil)

	query, err := parseRequestQueryParams(req)
	assert.NoError(t, err)
	assert.True(t, query.IsEmpty())
}

func TestParseRequestQuerySingleId(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?ids=000000000000000000000000", nil)

	query, err := parseRequestQueryParams(req)
	assert.NoError(t, err)
	assert.Len(t, query.Ids, 1)
	assert.Equal(t, query.Ids[0], "000000000000000000000000")
}

func TestParseRequestQueryMultipleIds(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?ids=000000000000000000000000,111111111111111111111111", nil)

	query, err := parseRequestQueryParams(req)
	assert.NoError(t, err)
	assert.Len(t, query.Ids, 2)
	assert.Equal(t, query.Ids[0], "000000000000000000000000")
	assert.Equal(t, query.Ids[1], "111111111111111111111111")
}

func TestParseRequestQueryInvalidId(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?ids=a", nil)

	_, err := parseRequestQueryParams(req)
	errors.AssertIsHTTPErrOfCode(t, err, http.StatusBadRequest)
}

func TestParseRequestQuerySingleName(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?names=asdf", nil)

	query, err := parseRequestQueryParams(req)
	assert.NoError(t, err)
	assert.Len(t, query.Names, 1)
	assert.Equal(t, "asdf", query.Names[0])
}

func TestParseRequestQueryMultipleNames(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?names=asdf,fdsa", nil)

	query, err := parseRequestQueryParams(req)
	assert.NoError(t, err)
	assert.Len(t, query.Names, 2)
	assert.Equal(t, "asdf", query.Names[0])
	assert.Equal(t, "fdsa", query.Names[1])
}

func TestParseRequestQuerySingleModel(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?models=asdf", nil)

	query, err := parseRequestQueryParams(req)
	assert.NoError(t, err)
	assert.Len(t, query.Models, 1)
	assert.Equal(t, "asdf", query.Models[0])
}

func TestParseRequestQueryMultipleModels(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?models=asdf,fdsa", nil)

	query, err := parseRequestQueryParams(req)
	assert.NoError(t, err)
	assert.Len(t, query.Models, 2)
	assert.Equal(t, "asdf", query.Models[0])
	assert.Equal(t, "fdsa", query.Models[1])
}

func TestParseRequestQuerySingleVersion(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?versions=asdf", nil)

	query, err := parseRequestQueryParams(req)
	assert.NoError(t, err)
	assert.Len(t, query.Versions, 1)
	assert.Equal(t, "asdf", query.Versions[0])
}

func TestParseRequestQueryMultipleVersions(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?versions=asdf,fdsa", nil)

	query, err := parseRequestQueryParams(req)
	assert.NoError(t, err)
	assert.Len(t, query.Versions, 2)
	assert.Equal(t, "asdf", query.Versions[0])
	assert.Equal(t, "fdsa", query.Versions[1])
}

func TestParseRequestQueryLimit(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?limit=1", nil)

	query, err := parseRequestQueryParams(req)
	assert.NoError(t, err)
	assert.Equal(t, 1, query.Limit)
}

func TestParseRequestQueryInvalidLimit(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?limit=a", nil)

	_, err := parseRequestQueryParams(req)
	errors.AssertIsHTTPErrOfCode(t, err, http.StatusBadRequest)
}

func TestParseRequestQueryZeroLimit(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?limit=0", nil)

	_, err := parseRequestQueryParams(req)
	errors.AssertIsHTTPErrOfCode(t, err, http.StatusBadRequest)
}

func TestParseRequestQueryNegativeLimit(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?limit=-1", nil)

	_, err := parseRequestQueryParams(req)
	errors.AssertIsHTTPErrOfCode(t, err, http.StatusBadRequest)
}

func TestParseRequestQueryOffset(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?offset=1", nil)

	query, err := parseRequestQueryParams(req)
	assert.NoError(t, err)
	assert.Equal(t, 1, query.Offset)
}

func TestParseRequestQueryInvalidOffset(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?offset=a", nil)

	_, err := parseRequestQueryParams(req)
	errors.AssertIsHTTPErrOfCode(t, err, http.StatusBadRequest)
}

func TestParseRequestQueryNegativeOffset(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?offset=-1", nil)

	_, err := parseRequestQueryParams(req)
	errors.AssertIsHTTPErrOfCode(t, err, http.StatusBadRequest)
}

func TestParseRequestQuerySortPositive(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?sort=asdf", nil)

	query, err := parseRequestQueryParams(req)
	assert.NoError(t, err)
	assert.Equal(t, "asdf", query.Sort)
	assert.False(t, query.SortDescending)
}

func TestParseRequestQuerySortNegative(t *testing.T) {
	req := httptest.NewRequest("POST", "/api-server?sort=asdf-", nil)

	query, err := parseRequestQueryParams(req)
	assert.NoError(t, err)
	assert.Equal(t, "asdf", query.Sort)
	assert.True(t, query.SortDescending)
}
