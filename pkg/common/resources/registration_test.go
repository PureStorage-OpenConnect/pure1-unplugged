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

	"github.com/stretchr/testify/assert"
)

func TestIsEqualEqual(t *testing.T) {
	assert.True(t, (&ArrayRegistrationInfo{
		ID:           "12345",
		Name:         "test-device",
		MgmtEndpoint: "8.8.8.8",
		APIToken:     "TOKEN!",
		DeviceType:   "Flasharray",
	}).IsEqual(&ArrayRegistrationInfo{
		ID:           "12345",
		Name:         "test-device",
		MgmtEndpoint: "8.8.8.8",
		APIToken:     "TOKEN!",
		DeviceType:   "Flasharray",
	}))
}

func TestIsEqualDifferentID(t *testing.T) {
	assert.False(t, (&ArrayRegistrationInfo{
		ID:           "12345",
		Name:         "test-device",
		MgmtEndpoint: "8.8.8.8",
		APIToken:     "TOKEN!",
		DeviceType:   "Flasharray",
	}).IsEqual(&ArrayRegistrationInfo{
		ID:           "67890",
		Name:         "test-device",
		MgmtEndpoint: "8.8.8.8",
		APIToken:     "TOKEN!",
		DeviceType:   "Flasharray",
	}))
}

func TestIsEqualDifferentName(t *testing.T) {
	assert.False(t, (&ArrayRegistrationInfo{
		ID:           "12345",
		Name:         "test-device",
		MgmtEndpoint: "8.8.8.8",
		APIToken:     "TOKEN!",
		DeviceType:   "Flasharray",
	}).IsEqual(&ArrayRegistrationInfo{
		ID:           "12345",
		Name:         "other-device",
		MgmtEndpoint: "8.8.8.8",
		APIToken:     "TOKEN!",
		DeviceType:   "Flasharray",
	}))
}

func TestIsEqualDifferentMgmtEndpoint(t *testing.T) {
	assert.False(t, (&ArrayRegistrationInfo{
		ID:           "12345",
		Name:         "test-device",
		MgmtEndpoint: "8.8.8.8",
		APIToken:     "TOKEN!",
		DeviceType:   "Flasharray",
	}).IsEqual(&ArrayRegistrationInfo{
		ID:           "12345",
		Name:         "test-device",
		MgmtEndpoint: "1.1.1.1",
		APIToken:     "TOKEN!",
		DeviceType:   "Flasharray",
	}))
}

func TestIsEqualDifferentToken(t *testing.T) {
	assert.False(t, (&ArrayRegistrationInfo{
		ID:           "12345",
		Name:         "test-device",
		MgmtEndpoint: "8.8.8.8",
		APIToken:     "TOKEN!",
		DeviceType:   "Flasharray",
	}).IsEqual(&ArrayRegistrationInfo{
		ID:           "12345",
		Name:         "test-device",
		MgmtEndpoint: "8.8.8.8",
		APIToken:     "notatoken",
		DeviceType:   "Flasharray",
	}))
}

func TestIsEqualDifferentDeviceType(t *testing.T) {
	assert.False(t, (&ArrayRegistrationInfo{
		ID:           "12345",
		Name:         "test-device",
		MgmtEndpoint: "8.8.8.8",
		APIToken:     "TOKEN!",
		DeviceType:   "Flasharray",
	}).IsEqual(&ArrayRegistrationInfo{
		ID:           "12345",
		Name:         "test-device",
		MgmtEndpoint: "8.8.8.8",
		APIToken:     "TOKEN!",
		DeviceType:   "Flashblade",
	}))
}

func TestLogFieldsNoEndpoint(t *testing.T) {
	fields := (&ArrayRegistrationInfo{
		ID:           "12345",
		Name:         "test-device",
		MgmtEndpoint: "8.8.8.8",
		APIToken:     "TOKEN!",
		DeviceType:   "Flasharray",
	}).GetLogFields(false)
	assert.Len(t, fields, 2)
	assert.Equal(t, "12345", fields["device_id"])
	assert.Equal(t, "test-device", fields["device_name"])
}

func TestLogFieldsEndpoint(t *testing.T) {
	fields := (&ArrayRegistrationInfo{
		ID:           "12345",
		Name:         "test-device",
		MgmtEndpoint: "8.8.8.8",
		APIToken:     "TOKEN!",
		DeviceType:   "Flasharray",
	}).GetLogFields(true)
	assert.Len(t, fields, 3)
	assert.Equal(t, "12345", fields["device_id"])
	assert.Equal(t, "test-device", fields["device_name"])
	assert.Equal(t, "8.8.8.8", fields["device_mgmt_endpoint"])
}
