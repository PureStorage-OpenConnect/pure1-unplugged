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

var routes = []Route{
	Route{ // Returns a list of registered storage arrays
		"ArrayGet",
		"GET",
		"/arrays",
		[]string{
			"filter", "{filter}",
			"ids", "{ids}",
			"names", "{names}",
			"limit", "{limit}",
			"offset", "{offset}",
			"sort", "{sort}",
		},
		getArrays,
	},
	// with body
	Route{ // Registers a new array
		"ArrayPost",
		"POST",
		"/arrays",
		[]string{},
		postArray,
	},
	// with body
	Route{ // Updates an array's information
		"ArrayPatch",
		"PATCH",
		"/arrays",
		[]string{
			"ids", "{ids}",
			"names", "{names}",
		},
		patchArray,
	},
	// no body
	Route{ // Unregisters an array
		"ArrayDelete",
		"DELETE",
		"/arrays",
		[]string{
			"ids", "{ids}",
			"names", "{names}",
		},
		deleteArray,
	},
	// no body
	Route{ // Returns a list of statuses of registered storage arrays
		"ArrayStatusGet",
		"GET",
		"/arrays/status",
		[]string{
			"filter", "{filter}",
			"ids", "{ids}",
			"names", "{names}",
			"limit", "{limit}",
			"offset", "{offset}",
			"sort", "{sort}",
		},
		getArrayStatus,
	},
	// no body
	Route{ // Returns a map of tags of registered storage arrays
		"ArrayTagsGet",
		"GET",
		"/arrays/tags",
		[]string{
			"filter", "{filter}",
			"ids", "{ids}",
			"names", "{names}",
			"limit", "{limit}",
			"offset", "{offset}",
			"sort", "{sort}",
		},
		getArrayTags,
	},
	// with body
	Route{ // Updates the tags of a registered storage array
		"ArrayTagsPatch",
		"PATCH",
		"/arrays/tags",
		[]string{
			"ids", "{ids}",
			"names", "{names}",
		},
		patchArrayTags,
	},
	// no body
	Route{ // Deletes a tag from a registered storage array
		"ArrayTagsDelete",
		"DELETE",
		"/arrays/tags",
		[]string{
			"ids", "{ids}",
			"names", "{names}",
			"tags", "{tags}",
		},
		deleteArrayTags,
	},
}
