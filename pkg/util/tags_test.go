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

package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	defaultNamespace = "pure1-unplugged"
)

func TestTagsListToMapEmpty(t *testing.T) {
	tagsList := []map[string]string{}
	tagsMap := TagsListToMap(map[string]string{}, &tagsList, defaultNamespace)

	assert.Empty(t, tagsMap)
}

func TestTagsListToMapSingle(t *testing.T) {
	tagsList := []map[string]string{
		{
			"namespace": "ns",
			"key":       "testkey",
			"value":     "testval",
		},
	}
	tagsMap := TagsListToMap(map[string]string{}, &tagsList, defaultNamespace)

	assert.Len(t, tagsMap, 1)
	assert.Equal(t, "testval", tagsMap["testkey,ns"])
}

func TestTagsListToMapSingleOverwrite(t *testing.T) {
	tagsList := []map[string]string{
		{
			"namespace": "ns",
			"key":       "testkey",
			"value":     "testval",
		},
	}
	tagsMap := TagsListToMap(map[string]string{
		"testkey,ns": "oldvalue",
	}, &tagsList, defaultNamespace)

	assert.Len(t, tagsMap, 1)
	assert.Equal(t, "testval", tagsMap["testkey,ns"])
}

func TestTagsListToMapSingleMissingNamespace(t *testing.T) {
	tagsMap := map[string]string{}
	tagsList := []map[string]string{
		{
			"key":   "testkey",
			"value": "testval",
		},
	}
	tagsMap = TagsListToMap(tagsMap, &tagsList, defaultNamespace)

	assert.Len(t, tagsMap, 1)
	assert.Equal(t, "testval", tagsMap[fmt.Sprintf("testkey,%s", defaultNamespace)])
}

func TestTagsListToMapSingleEmptyNamespace(t *testing.T) {
	tagsMap := map[string]string{}
	tagsList := []map[string]string{
		{
			"namespace": "",
			"key":       "testkey",
			"value":     "testval",
		},
	}
	tagsMap = TagsListToMap(tagsMap, &tagsList, defaultNamespace)

	assert.Len(t, tagsMap, 1)
	assert.Equal(t, "testval", tagsMap["testkey,"])
}

func TestTagsListToMapMultiple(t *testing.T) {
	tagsMap := map[string]string{
		"testkey,ns": "oldvalue",
	}
	tagsList := []map[string]string{
		{
			"namespace": "ns",
			"key":       "testkey",
			"value":     "testval",
		},
		{
			"namespace": "ns",
			"key":       "anotherkey",
			"value":     "magicalvalue",
		},
	}
	tagsMap = TagsListToMap(tagsMap, &tagsList, defaultNamespace)

	assert.Len(t, tagsMap, 2)
	assert.Equal(t, "testval", tagsMap["testkey,ns"])
	assert.Equal(t, "magicalvalue", tagsMap["anotherkey,ns"])
}

func TestTagsMapToListEmpty(t *testing.T) {
	tagsMap := map[string]string{}

	tagsList, err := TagsMapToList(tagsMap)

	assert.NoError(t, err)
	assert.Empty(t, tagsList)
}

func TestTagsMapToListSingle(t *testing.T) {
	tagsMap := map[string]string{
		"akey,ns": "somevalue",
	}

	tagsList, err := TagsMapToList(tagsMap)

	assert.NoError(t, err)
	assert.Len(t, tagsList, 1)
	assert.Equal(t, []map[string]string{
		{
			"namespace": "ns",
			"key":       "akey",
			"value":     "somevalue",
		},
	}, tagsList)
}

func TestTagsMapToListMultiple(t *testing.T) {
	tagsMap := map[string]string{
		"akey,ns":           "somevalue",
		"anotherkey,diffns": "someothervalue",
	}

	tagsList, err := TagsMapToList(tagsMap)

	assert.NoError(t, err)
	assert.Len(t, tagsList, 2)
	assert.Contains(t, tagsList, map[string]string{
		"namespace": "ns",
		"key":       "akey",
		"value":     "somevalue",
	})
	assert.Contains(t, tagsList, map[string]string{
		"namespace": "diffns",
		"key":       "anotherkey",
		"value":     "someothervalue",
	})
}

func TestTagsMapToListMultipleSameNamespace(t *testing.T) {
	tagsMap := map[string]string{
		"akey,ns":       "somevalue",
		"anotherkey,ns": "someothervalue",
	}

	tagsList, err := TagsMapToList(tagsMap)

	assert.NoError(t, err)
	assert.Len(t, tagsList, 2)
	assert.Contains(t, tagsList, map[string]string{
		"namespace": "ns",
		"key":       "akey",
		"value":     "somevalue",
	})
	assert.Contains(t, tagsList, map[string]string{
		"namespace": "ns",
		"key":       "anotherkey",
		"value":     "someothervalue",
	})
}

func TestTagsMapToListEmptyNamespace(t *testing.T) {
	tagsMap := map[string]string{
		"akey,": "somevalue",
	}

	tagsList, err := TagsMapToList(tagsMap)

	assert.NoError(t, err)
	assert.Len(t, tagsList, 1)
	assert.Equal(t, []map[string]string{
		{
			"namespace": "",
			"key":       "akey",
			"value":     "somevalue",
		},
	}, tagsList)
}

func TestTagsMapToListMissingNamespace(t *testing.T) {
	tagsMap := map[string]string{
		"akey": "somevalue",
	}

	_, err := TagsMapToList(tagsMap)

	assert.Error(t, err)
}

func TestTagsMapToListNoNamespaceEmpty(t *testing.T) {
	tagsMap := map[string]string{}

	list := TagsMapToListNoNamespace(tagsMap, "ns")
	assert.Empty(t, list)
}

func TestTagsMapToListNoNamespaceSingle(t *testing.T) {
	tagsMap := map[string]string{
		"akey": "avalue",
	}

	list := TagsMapToListNoNamespace(tagsMap, "ns")
	assert.Len(t, list, 1)
	assert.Equal(t, list[0], map[string]string{
		"key":       "akey",
		"value":     "avalue",
		"namespace": "ns",
	})
}

func TestTagsMapToListNoNamespaceMultiple(t *testing.T) {
	tagsMap := map[string]string{
		"akey":       "avalue",
		"anotherkey": "anothervalue",
	}

	list := TagsMapToListNoNamespace(tagsMap, "ns")
	assert.Len(t, list, 2)
	assert.Contains(t, list, map[string]string{
		"key":       "akey",
		"value":     "avalue",
		"namespace": "ns",
	})
	assert.Contains(t, list, map[string]string{
		"key":       "anotherkey",
		"value":     "anothervalue",
		"namespace": "ns",
	})
}

func TestTagsListSetNamespaceEmpty(t *testing.T) {
	tagsList := []map[string]string{}

	TagsListSetDefaultNamespace(tagsList, "defns")

	assert.Len(t, tagsList, 0)
}

func TestTagsListSetNamespaceWithNamespace(t *testing.T) {
	tagsList := []map[string]string{
		{
			"key":       "akey",
			"value":     "avalue",
			"namespace": "someotherns",
		},
	}

	TagsListSetDefaultNamespace(tagsList, "defns")

	assert.Len(t, tagsList, 1)
	assert.Equal(t, map[string]string{
		"key":       "akey",
		"value":     "avalue",
		"namespace": "someotherns",
	}, tagsList[0])
}

func TestTagsListSetNamespaceWithoutNamespace(t *testing.T) {
	tagsList := []map[string]string{
		{
			"key":   "akey",
			"value": "avalue",
		},
	}

	TagsListSetDefaultNamespace(tagsList, "defns")

	assert.Len(t, tagsList, 1)
	assert.Equal(t, map[string]string{
		"key":       "akey",
		"value":     "avalue",
		"namespace": "defns",
	}, tagsList[0])
}

func TestSplitTagsByNamespaceEmpty(t *testing.T) {
	tags := []map[string]string{}

	matched, unmatched, err := SplitTagsByNamespace(tags, "ns")
	assert.NoError(t, err)
	assert.Empty(t, matched)
	assert.Empty(t, unmatched)
}

func TestSplitTagsByNamespaceOnlyMatching(t *testing.T) {
	tags := []map[string]string{
		{
			"key":       "akey",
			"value":     "avalue",
			"namespace": "ns",
		},
	}

	matched, unmatched, err := SplitTagsByNamespace(tags, "ns")
	assert.NoError(t, err)
	assert.Len(t, matched, 1)
	assert.Equal(t, "avalue", matched["akey"])
	assert.Empty(t, unmatched)
}

func TestSplitTagsByNamespaceOnlyUnmatching(t *testing.T) {
	tags := []map[string]string{
		{
			"key":       "akey",
			"value":     "avalue",
			"namespace": "not-ns",
		},
	}

	matched, unmatched, err := SplitTagsByNamespace(tags, "ns")
	assert.NoError(t, err)
	assert.Empty(t, matched)
	assert.Len(t, unmatched, 1)
	assert.Equal(t, "avalue", unmatched["akey"])
}

func TestSplitTagsByNamespaceOneOfEach(t *testing.T) {
	tags := []map[string]string{
		{
			"key":       "akey",
			"value":     "avalue",
			"namespace": "not-ns",
		},
		{
			"key":       "anotherkey",
			"value":     "anothervalue",
			"namespace": "ns",
		},
	}

	matched, unmatched, err := SplitTagsByNamespace(tags, "ns")
	assert.NoError(t, err)
	assert.Len(t, matched, 1)
	assert.Equal(t, "anothervalue", matched["anotherkey"])
	assert.Len(t, unmatched, 1)
	assert.Equal(t, "avalue", unmatched["akey"])
}

func TestSplitTagsByNamespaceMissingKey(t *testing.T) {
	tags := []map[string]string{
		{
			"value":     "avalue",
			"namespace": "ns",
		},
	}

	_, _, err := SplitTagsByNamespace(tags, "ns")
	assert.Error(t, err)
}

func TestSplitTagsByNamespaceMissingValue(t *testing.T) {
	tags := []map[string]string{
		{
			"key":       "akey",
			"namespace": "ns",
		},
	}

	_, _, err := SplitTagsByNamespace(tags, "ns")
	assert.Error(t, err)
}

func TestSplitTagsByNamespaceMissingNamespace(t *testing.T) {
	tags := []map[string]string{
		{
			"key":   "akey",
			"value": "avalue",
		},
	}

	_, _, err := SplitTagsByNamespace(tags, "ns")
	assert.Error(t, err)
}
