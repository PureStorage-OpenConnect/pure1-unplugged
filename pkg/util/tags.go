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
	"strings"
)

// TagsListToMap takes in a list of tags ({ key, namespace, value }) and converts it into a flat map of the format
// "key,namespace" (a single string) => value
func TagsListToMap(tagsMap map[string]string, tagsList *[]map[string]string, defaultNamespace string) map[string]string {
	for _, m := range *tagsList {
		_, ok := m["namespace"]
		if !ok {
			m["namespace"] = defaultNamespace
		}
		key := m["key"] + "," + m["namespace"]
		val := m["value"]
		tagsMap[key] = val
	}
	return tagsMap
}

// TagsMapToList converts a flat map (usually produced by TagsListToMap) with the format "key,namespace" => value
// into a list of tags ({ key, namespace, value })
func TagsMapToList(tagsMap map[string]string) ([]map[string]string, error) {
	returnList := []map[string]string{}

	for k, v := range tagsMap {
		if len(strings.Split(k, ",")) != 2 {
			return returnList, fmt.Errorf("Found a tag without namespace")
		}
		key := strings.Split(k, ",")[0]
		ns := strings.Split(k, ",")[1]
		m := map[string]string{
			"namespace": ns,
			"key":       key,
			"value":     v,
		}
		returnList = append(returnList, m)
	}

	return returnList, nil
}

// TagsMapToListNoNamespace converts a flat map in the format "key->value" into a list of tags ({ key, namespace, value}),
// setting the namespace to the provided parameter
func TagsMapToListNoNamespace(tagsMap map[string]string, setNamespace string) []map[string]string {
	returnList := []map[string]string{}

	for k, v := range tagsMap {
		m := map[string]string{
			"namespace": setNamespace,
			"key":       k,
			"value":     v,
		}
		returnList = append(returnList, m)
	}

	return returnList
}

// TagsListSetDefaultNamespace iterates through every tag in a list of tags, and if
// it doesn't have a namespace set it will set it to the provided default namespacece
func TagsListSetDefaultNamespace(tags []map[string]string, namespace string) {
	for _, m := range tags {
		if m != nil && len(m) != 0 {
			_, ok := m["namespace"]
			if !ok {
				m["namespace"] = namespace
			}
		}
	}
}

// SplitTagsByNamespace searches through a list of tags and separates them by namespace. Specifically, it returns
// one map of all tags (key->value) that matched the given namespace, and another map of all tags (key->value) that didn't.
// Note that this *is* a lossy operation, since all tags that don't match the namespace won't have the namespace
// stored, it's just dropped.
func SplitTagsByNamespace(tags []map[string]string, searchNamespace string) (map[string]string, map[string]string, error) {
	matchedMap := map[string]string{}
	unmatchedMap := map[string]string{}
	for _, m := range tags {
		key, keyok := m["key"]
		val, valok := m["value"]
		ns, nsok := m["namespace"]
		if !keyok || !valok || !nsok {
			return matchedMap, unmatchedMap, fmt.Errorf("Tag is missing either key, value, or namespace field")
		}
		if nsok && ns == searchNamespace {
			matchedMap[key] = val
		} else {
			unmatchedMap[key] = val
		}
	}
	return matchedMap, unmatchedMap, nil
}
