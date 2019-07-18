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

package resources

import (
	"testing"

	"github.com/olivere/elastic"

	"github.com/stretchr/testify/assert"
)

func TestQueryEmpty(t *testing.T) {
	query := ArrayQuery{}

	assert.True(t, query.IsEmpty())
}

func TestQueryEmptyWithID(t *testing.T) {
	query := ArrayQuery{Ids: []string{"asdf"}}

	assert.False(t, query.IsEmpty())
}

func TestQueryEmptyWithName(t *testing.T) {
	query := ArrayQuery{Names: []string{"asdf"}}

	assert.False(t, query.IsEmpty())
}

func TestQueryEmptyWithModel(t *testing.T) {
	query := ArrayQuery{Models: []string{"asdf"}}

	assert.False(t, query.IsEmpty())
}

func TestQueryEmptyWithVersion(t *testing.T) {
	query := ArrayQuery{Versions: []string{"asdf"}}

	assert.False(t, query.IsEmpty())
}

func TestRemoveNilQueriesEmpty(t *testing.T) {
	result := removeNilQueries()
	assert.Empty(t, result)
}

func TestRemoveNilQueriesNoEmpty(t *testing.T) {
	result := removeNilQueries(elastic.NewMatchAllQuery(), elastic.NewMatchNoneQuery())
	assert.Len(t, result, 2)
}

func TestRemoveNilQueriesWithEmpty(t *testing.T) {
	result := removeNilQueries(elastic.NewMatchAllQuery(), nil, elastic.NewMatchNoneQuery())
	assert.Len(t, result, 2)
}

func TestGetSortParameterEmpty(t *testing.T) {
	query := ArrayQuery{}
	sort, err := query.GetSortParameter()
	assert.NoError(t, err)
	assert.Empty(t, sort)
}

func TestGetSortParameterName(t *testing.T) {
	query := ArrayQuery{Sort: "name"}
	sort, err := query.GetSortParameter()
	assert.NoError(t, err)
	assert.Equal(t, "Name.keyword", sort)
}

func TestGetSortParameterModel(t *testing.T) {
	query := ArrayQuery{Sort: "model"}
	sort, err := query.GetSortParameter()
	assert.NoError(t, err)
	assert.Equal(t, "Model.keyword", sort)
}

func TestGetSortParameterVersion(t *testing.T) {
	query := ArrayQuery{Sort: "version"}
	sort, err := query.GetSortParameter()
	assert.NoError(t, err)
	assert.Equal(t, "Version.keyword", sort)
}

func TestGetSortParameterInvalid(t *testing.T) {
	query := ArrayQuery{Sort: "asdf"}
	_, err := query.GetSortParameter()
	assert.Error(t, err)
}
