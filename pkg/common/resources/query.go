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
	"fmt"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/http/errors"
	"github.com/olivere/elastic"
)

// IsEmpty returns whether this query has any filter parameters set or not
// (it doesn't care about sorting or pagination)
func (q *ArrayQuery) IsEmpty() bool {
	return len(q.Ids) == 0 && len(q.Names) == 0 && len(q.Models) == 0 && len(q.Versions) == 0
}

func removeNilQueries(queries ...elastic.Query) []elastic.Query {
	newQueries := []elastic.Query{}

	for _, query := range queries {
		if query != nil {
			newQueries = append(newQueries, query)
		}
	}

	return newQueries
}

func generateTermQueryObject(elasticQuery string, terms ...string) elastic.Query {
	if len(terms) == 0 {
		return nil
	}
	queries := []elastic.Query{}
	for _, term := range terms {
		queries = append(queries, elastic.NewWildcardQuery(elasticQuery, term))
	}
	if len(queries) == 1 {
		return queries[0]
	}
	return elastic.NewBoolQuery().Should(queries...)
}

func generateIdsQueryObject(ids ...string) elastic.Query {
	if len(ids) == 0 {
		return nil
	}
	return elastic.NewIdsQuery().Ids(ids...)
}

// GetSortParameter converts the parameter in the query
// to an Elastic sort query
func (q *ArrayQuery) GetSortParameter() (string, error) {
	switch q.Sort {
	case "":
		return "", nil
	case "name":
		return "Name.keyword", nil
	case "model":
		return "Model.keyword", nil
	case "version":
		return "Version.keyword", nil
	default:
		return "", errors.MakeBadRequestHTTPErr(fmt.Errorf("Invalid sort parameter"))
	}
}

// GenerateElasticQueryObject converts this query into an elastic.Query
// object suitable for API calls
func (q *ArrayQuery) GenerateElasticQueryObject() elastic.Query {
	if q.IsEmpty() {
		qu := elastic.NewMatchAllQuery()
		return qu
	}
	queries := []elastic.Query{}
	queries = append(queries, generateIdsQueryObject(q.Ids...)) // Presumably faster than a term query, since ID is a higher-level elastic concept
	queries = append(queries, generateTermQueryObject("Name.keyword", q.Names...))
	queries = append(queries, generateTermQueryObject("Model.keyword", q.Models...))
	queries = append(queries, generateTermQueryObject("Version.keyword", q.Versions...))
	queries = removeNilQueries(queries...)

	outerQuery := elastic.NewBoolQuery().Must(queries...)

	return outerQuery
}

// GenerateEmptyQuery generates an empty query,
// like if it was returned from parseRequestQueryParams
// with no query parameters.
func GenerateEmptyQuery() ArrayQuery {
	return ArrayQuery{
		Ids:            []string{},
		Names:          []string{},
		Models:         []string{},
		Versions:       []string{},
		Sort:           "",
		SortDescending: false,
		Offset:         0,
		Limit:          0,
	}
}
