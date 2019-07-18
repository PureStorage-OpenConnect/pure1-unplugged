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

package version

var buildVersion string

// Get will return the version set. Typically it is used by setting the
// version string at link time. If one was not set it will return an
// ambiguous value "v?"
func Get() string {
	if buildVersion != "" {
		return buildVersion
	}
	return "v?"
}
