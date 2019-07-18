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
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common"

	"github.com/stretchr/testify/assert"
)

func TestParseFromMapDefaults(t *testing.T) {
	parsed, err := ParseArrayFromREST(map[string]interface{}{})
	assert.NoError(t, err)
	assert.Equal(t, "", parsed.InternalID)
	assert.Equal(t, "", parsed.Name)
	assert.Equal(t, "", parsed.MgmtEndPoint)
	assert.Equal(t, "", parsed.APIToken)
	assert.Equal(t, "", parsed.Status)
	assert.Equal(t, "", parsed.DeviceType)
	assert.Equal(t, time.Time{}, parsed.Lastseen)
	assert.Equal(t, time.Time{}, parsed.Lastupdated)
}

func TestParseFromMapSetAll(t *testing.T) {
	parsed, err := ParseArrayFromREST(map[string]interface{}{
		"id":            "1234567890abcdefedcba098",
		"name":          "test_array",
		"mgmt_endpoint": "192.168.99.100",
		"api_token":     "asdf",
		"status":        "Connected",
		"device_type":   common.FlashBlade,
		"_as_of":        "2019-01-30T17:19:26.000",
		"_last_updated": "2019-01-30T17:19:26.000",
	})
	assert.NoError(t, err)
	assert.Equal(t, "1234567890abcdefedcba098", parsed.InternalID)
	assert.Equal(t, "test_array", parsed.Name)
	assert.Equal(t, "192.168.99.100", parsed.MgmtEndPoint)
	assert.Equal(t, "asdf", parsed.APIToken)
	assert.Equal(t, "Connected", parsed.Status)
	assert.Equal(t, common.FlashBlade, parsed.DeviceType)
	assert.EqualValues(t, 1548868766, parsed.Lastseen.Unix())
	assert.EqualValues(t, 1548868766, parsed.Lastupdated.Unix())
}

func TestParseFromMapBadID(t *testing.T) {
	_, err := ParseArrayFromREST(map[string]interface{}{
		"id": "asdf",
	})
	assert.Error(t, err)
}

func TestParseFromMapBadAsOf(t *testing.T) {
	_, err := ParseArrayFromREST(map[string]interface{}{
		"_as_of": "asdf",
	})
	assert.Error(t, err)
}

func TestParseFromMapBadLastUpdated(t *testing.T) {
	_, err := ParseArrayFromREST(map[string]interface{}{
		"_last_updated": "asdf",
	})
	assert.Error(t, err)
}

func TestApplyPatchEmpty(t *testing.T) {
	array := Array{}
	err := array.ApplyPatch(map[string]interface{}{})
	assert.NoError(t, err)
	// No change should occur, not even LastUpdated, since, well, nothing was updated
	assert.Equal(t, Array{}, array)
}

func TestApplyPatchFull(t *testing.T) {
	array := Array{}
	err := array.ApplyPatch(map[string]interface{}{
		"name":          "test_dev1",
		"mgmt_endpoint": "192.168.99.100",
		"device_type":   common.FlashBlade,
		"api_token":     "asdf",
		"status":        "Connected",
		"_as_of":        "2019-01-30T17:19:26.000",
	})
	assert.NotEqual(t, time.Time{}, array.Lastupdated) // Make sure that Lastupdated got set, then clear it for testing purposees
	array.Lastupdated = time.Time{}
	assert.NoError(t, err)
	assert.Equal(t, "test_dev1", array.Name)
	assert.Equal(t, "192.168.99.100", array.MgmtEndPoint)
	assert.Equal(t, common.FlashBlade, array.DeviceType)
	assert.Equal(t, "asdf", array.APIToken)
	assert.Equal(t, "Connected", array.Status)
	assert.EqualValues(t, 1548868766, array.Lastseen.Unix())
}

func TestApplyPatchEmptyName(t *testing.T) {
	array := Array{}
	err := array.ApplyPatch(map[string]interface{}{
		"name": "  ",
	})
	assert.Error(t, err)
}

func TestApplyPatchEmptyManagement(t *testing.T) {
	array := Array{}
	err := array.ApplyPatch(map[string]interface{}{
		"mgmt_endpoint": "  ",
	})
	assert.Error(t, err)
}

func TestApplyPatchEmptyDeviceType(t *testing.T) {
	array := Array{}
	err := array.ApplyPatch(map[string]interface{}{
		"device_type": "  ",
	})
	assert.Error(t, err)
}

func TestApplyPatchEmptyAPIToken(t *testing.T) {
	array := Array{}
	err := array.ApplyPatch(map[string]interface{}{
		"api_token": "  ",
	})
	assert.Error(t, err)
}

func TestApplyPatchEmptyAsOf(t *testing.T) {
	array := Array{}
	err := array.ApplyPatch(map[string]interface{}{
		"_as_of": "  ",
	})
	assert.Error(t, err)
}

func TestApplyPatchInvalidAsOf(t *testing.T) {
	array := Array{}
	err := array.ApplyPatch(map[string]interface{}{
		"_as_of": "asdf",
	})
	assert.Error(t, err)
}

func TestTagKeyAssert(t *testing.T) {
	tag := map[string]string{
		"key":       "test_key",
		"value":     "test_value",
		"namespace": "test_ns",
	}
	err := assertTagHasProperKeys(tag)
	assert.NoError(t, err)
}

func TestTagKeyAssertMissingKey(t *testing.T) {
	tag := map[string]string{
		"value":     "test_value",
		"namespace": "test_ns",
	}
	err := assertTagHasProperKeys(tag)
	assert.Error(t, err)
}

func TestTagKeyAssertMissingValue(t *testing.T) {
	tag := map[string]string{
		"key":       "test_key",
		"namespace": "test_ns",
	}
	err := assertTagHasProperKeys(tag)
	assert.Error(t, err)
}

func TestTagKeyAssertMissingNamespace(t *testing.T) {
	tag := map[string]string{
		"key":   "test_key",
		"value": "test_value",
	}
	err := assertTagHasProperKeys(tag)
	assert.Error(t, err)
}

func TestApplyTagPatchNewTag(t *testing.T) {
	array := Array{}
	err := array.ApplyTagPatch([]map[string]string{
		map[string]string{
			"key":       "test_key",
			"value":     "test_value",
			"namespace": "test_ns",
		},
	})
	assert.NoError(t, err)
	assert.Len(t, array.Tags, 1)
	assert.Equal(t, "test_key", array.Tags[0]["key"])
	assert.Equal(t, "test_value", array.Tags[0]["value"])
	assert.Equal(t, "test_ns", array.Tags[0]["namespace"])
}

func TestApplyTagPatchNilExistingTags(t *testing.T) {
	array := Array{Tags: nil}
	err := array.ApplyTagPatch([]map[string]string{
		map[string]string{
			"key":       "test_key",
			"value":     "test_value",
			"namespace": "test_ns",
		},
	})
	assert.NoError(t, err)
	assert.Len(t, array.Tags, 1)
	assert.Equal(t, "test_key", array.Tags[0]["key"])
	assert.Equal(t, "test_value", array.Tags[0]["value"])
	assert.Equal(t, "test_ns", array.Tags[0]["namespace"])
}

func TestApplyTagPatchExistingTag(t *testing.T) {
	array := Array{
		Tags: []map[string]string{
			map[string]string{
				"key":       "test_key",
				"value":     "test_value",
				"namespace": "test_ns",
			},
		},
	}
	err := array.ApplyTagPatch([]map[string]string{
		map[string]string{
			"key":       "test_key",
			"value":     "new_value",
			"namespace": "test_ns",
		},
	})
	assert.NoError(t, err)
	assert.Len(t, array.Tags, 1)
	assert.Equal(t, "test_key", array.Tags[0]["key"])
	assert.Equal(t, "new_value", array.Tags[0]["value"])
	assert.Equal(t, "test_ns", array.Tags[0]["namespace"])
}

func TestApplyTagPatchBadTag(t *testing.T) {
	array := Array{
		Tags: []map[string]string{
			map[string]string{
				"key":       "test_key",
				"value":     "test_value",
				"namespace": "test_ns",
			},
		},
	}
	err := array.ApplyTagPatch([]map[string]string{
		map[string]string{
			"key":   "test_key",
			"value": "new_value",
			// Missing namespace
		},
	})
	assert.Error(t, err)
}

func TestDeleteTagsEmpty(t *testing.T) {
	array := Array{
		Tags: []map[string]string{
			map[string]string{
				"key":       "test_key",
				"value":     "test_value",
				"namespace": "test_ns",
			},
		},
	}
	array.DeleteTags([]string{})
	assert.Len(t, array.Tags, 1)
}

func TestDeleteTagsNonexistentKey(t *testing.T) {
	array := Array{
		Tags: []map[string]string{
			map[string]string{
				"key":       "test_key",
				"value":     "test_value",
				"namespace": "test_ns",
			},
		},
	}
	array.DeleteTags([]string{"other_key"})
	assert.Len(t, array.Tags, 1)
}

func TestDeleteTags(t *testing.T) {
	array := Array{
		Tags: []map[string]string{
			map[string]string{
				"key":       "test_key",
				"value":     "test_value",
				"namespace": "test_ns",
			},
		},
	}
	array.DeleteTags([]string{"test_key"})
	assert.Empty(t, array.Tags)
}

func TestDeleteTagsEmptyTags(t *testing.T) {
	array := Array{
		Tags: []map[string]string{},
	}
	array.DeleteTags([]string{"test_key"})
	assert.Empty(t, array.Tags)
}

func TestDeleteTagsNilTags(t *testing.T) {
	array := Array{Tags: nil}
	array.DeleteTags([]string{"test_key"})
	assert.Empty(t, array.Tags)
}

func TestHasRequiredFields(t *testing.T) {
	array := Array{
		InternalID:   "asdf",
		Name:         "test_dev1",
		MgmtEndPoint: "192.168.99.100",
		DeviceType:   common.FlashBlade,
		APIToken:     "asdf",
	}
	assert.NoError(t, array.HasRequiredPostFields())
}

func TestHasRequiredFieldsMissingID(t *testing.T) {
	array := Array{
		InternalID:   "  ",
		Name:         "test_dev1",
		MgmtEndPoint: "192.168.99.100",
		DeviceType:   common.FlashBlade,
		APIToken:     "asdf",
	}
	assert.Error(t, array.HasRequiredPostFields())
}

func TestHasRequiredFieldsMissingName(t *testing.T) {
	array := Array{
		InternalID:   "asdf",
		Name:         "  ",
		MgmtEndPoint: "192.168.99.100",
		DeviceType:   common.FlashBlade,
		APIToken:     "asdf",
	}
	assert.Error(t, array.HasRequiredPostFields())
}

func TestHasRequiredFieldsMissingDeviceType(t *testing.T) {
	array := Array{
		InternalID:   "asdf",
		Name:         "test_dev1",
		MgmtEndPoint: "192.168.99.100",
		DeviceType:   "  ",
		APIToken:     "asdf",
	}
	assert.Error(t, array.HasRequiredPostFields())
}

func TestHasRequiredFieldsMissingMgmtEndpoint(t *testing.T) {
	array := Array{
		InternalID:   "asdf",
		Name:         "test_dev1",
		MgmtEndPoint: "  ",
		DeviceType:   common.FlashBlade,
		APIToken:     "asdf",
	}
	assert.Error(t, array.HasRequiredPostFields())
}

func TestHasRequiredFieldsMissingApiToken(t *testing.T) {
	array := Array{
		InternalID:   "asdf",
		Name:         "test_dev1",
		MgmtEndPoint: "192.168.99.100",
		DeviceType:   common.FlashBlade,
		APIToken:     "  ",
	}
	assert.Error(t, array.HasRequiredPostFields())
}

func TestValidateHexValid(t *testing.T) {
	assert.NoError(t, ValidateHexObjectID("1234567890abcdefedcba098")) // All valid hex characters
}

func TestValidateHexTooShort(t *testing.T) {
	assert.Error(t, ValidateHexObjectID("00000000000000000000000")) // too short (23 characters)
}

func TestValidateHexTooLong(t *testing.T) {
	assert.Error(t, ValidateHexObjectID("0000000000000000000000000")) // too long (25 characters)
}

func TestValidateHexInvalid(t *testing.T) {
	assert.Error(t, ValidateHexObjectID("00000000000000000000000g")) // invalid hex character
}

func TestValidateHexEmpty(t *testing.T) {
	assert.Error(t, ValidateHexObjectID("")) // empty string should be invalid
}
