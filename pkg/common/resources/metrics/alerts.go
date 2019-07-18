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

package metrics

import "strings"

// PopulateSeverityIndex sets the SeverityIndex field
// of this alert based on the value of Severity
func (a *Alert) PopulateSeverityIndex() {
	switch strings.ToLower(a.Severity) {
	case "info":
		a.SeverityIndex = 1
		return
	case "warning":
		a.SeverityIndex = 2
		return
	case "critical":
		a.SeverityIndex = 3
		return
	default:
		a.SeverityIndex = 0
		return
	}
}
