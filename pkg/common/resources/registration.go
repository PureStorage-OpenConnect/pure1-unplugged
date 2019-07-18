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

import log "github.com/sirupsen/logrus"

// GetLogFields is an ease-of-use method to get the easy describing fields of
// a device's registration info
func (a *ArrayRegistrationInfo) GetLogFields(includeEndpoint bool) log.Fields {
	fields := log.Fields{
		"device_id":   a.ID,
		"device_name": a.Name,
	}
	if includeEndpoint {
		fields["device_mgmt_endpoint"] = a.MgmtEndpoint
	}
	return fields
}

// IsEqual compares this struct to the given one
func (a *ArrayRegistrationInfo) IsEqual(other *ArrayRegistrationInfo) bool {
	return a.ID == other.ID &&
		a.DeviceType == other.DeviceType &&
		a.MgmtEndpoint == other.MgmtEndpoint &&
		a.APIToken == other.APIToken &&
		a.Name == other.Name
}
