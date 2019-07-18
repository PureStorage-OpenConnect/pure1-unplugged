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
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

const (
	// DateTimeFormat holds the format string for the date time
	// going in and out of Elasticsearch
	DateTimeFormat = "2006-01-02T15:04:05.000"
)

// ConvertToArrayMap converts this array into a string->interface map
// suitable for marshalling, with only the array properties (sans-tags)
func (s *Array) ConvertToArrayMap() map[string]interface{} {
	return map[string]interface{}{
		"id":            s.InternalID,
		"name":          s.Name,
		"mgmt_endpoint": s.MgmtEndPoint,
		"api_token":     s.APIToken,
		"status":        s.Status,
		"device_type":   s.DeviceType,
		"model":         s.Model,
		"version":       s.Version,
		"_as_of":        s.Lastseen,
		"_last_updated": s.Lastupdated,
	}
}

// ConvertToStatusMap converts this array into a string->interface map
// suitable for marshalling, with only the array ID and status
func (s *Array) ConvertToStatusMap() map[string]interface{} {
	return map[string]interface{}{
		"id":     s.InternalID,
		"status": s.Status,
		"_as_of": s.Lastseen,
	}
}

// ConvertToTagsMap converts this array into a string->interface map
// suitable for marshalling, with only the array ID and tags
func (s *Array) ConvertToTagsMap() map[string]interface{} {
	toReturn := map[string]interface{}{
		"id":     s.InternalID,
		"_as_of": s.Lastseen,
	}
	if s.Tags == nil {
		toReturn["tags"] = []map[string]string{}
	} else {
		toReturn["tags"] = s.Tags
	}
	return toReturn
}

// ApplyPatch applies the given patches to this array, with the given
// map in the format accepted by the REST API ("mgmt_endpoint", "device_type", etc.)
func (s *Array) ApplyPatch(m map[string]interface{}) error {
	if m == nil || len(m) == 0 {
		return nil
	}

	// Used to mark if anything *consequential* was changed:
	// specifically display name, mgmt_endpoint, device_type, or api_token
	changed := false

	if _, ok := m["name"]; ok {
		if len(strings.TrimSpace(m["name"].(string))) == 0 {
			return fmt.Errorf("Key name cannot be empty")
		}
		s.Name = m["name"].(string)
		changed = true
	}
	if _, ok := m["mgmt_endpoint"]; ok {
		if len(strings.TrimSpace(m["mgmt_endpoint"].(string))) == 0 {
			return fmt.Errorf("Key mgmt_endpoint cannot be empty")
		}
		s.MgmtEndPoint = m["mgmt_endpoint"].(string)
		changed = true
	}
	if _, ok := m["device_type"]; ok {
		if len(strings.TrimSpace(m["device_type"].(string))) == 0 {
			return fmt.Errorf("Key device_type cannot be empty")
		}
		s.DeviceType = m["device_type"].(string)
		changed = true
	}
	if _, ok := m["api_token"]; ok {
		if len(strings.TrimSpace(m["api_token"].(string))) == 0 {
			return fmt.Errorf("Key api_token cannot be empty")
		}
		s.APIToken = m["api_token"].(string)
		changed = true
	}
	if _, ok := m["status"]; ok {
		// This is valid to be empty
		s.Status = m["status"].(string)
	}
	if _, ok := m["model"]; ok {
		if len(strings.TrimSpace(m["model"].(string))) == 0 {
			return fmt.Errorf("Key model cannot be empty")
		}
		s.Model = m["model"].(string)
	}
	if _, ok := m["version"]; ok {
		if len(strings.TrimSpace(m["version"].(string))) == 0 {
			return fmt.Errorf("Key version cannot be empty")
		}
		s.Version = m["version"].(string)
	}

	if _, ok := m["_as_of"]; ok {
		if len(strings.TrimSpace(m["_as_of"].(string))) == 0 {
			return fmt.Errorf("Key _as_of cannot be empty")
		}
		lastseen, err := time.Parse("2006-01-02T15:04:05.000", m["_as_of"].(string)) // yyyy-MM-dd'T'HH:mm:ss.SSS
		if err != nil {
			return err
		}
		s.Lastseen = lastseen.UTC()
	}

	if changed {
		s.Lastupdated = time.Now().UTC()
	}

	return nil
}

func assertTagHasProperKeys(tag map[string]string) error {
	if _, ok := tag["key"]; !ok {
		return fmt.Errorf("Tag must have key")
	}
	if _, ok := tag["value"]; !ok {
		return fmt.Errorf("Tag must have value")
	}
	if _, ok := tag["namespace"]; !ok {
		return fmt.Errorf("Tag must have namespace")
	}
	return nil
}

// ApplyTagPatch patches the tags of this array
func (s *Array) ApplyTagPatch(tags []map[string]string) error {
	var existingTags []map[string]string

	if s.Tags == nil {
		existingTags = []map[string]string{}
	} else {
		existingTags = s.Tags
	}

	// This algorithm is O(n^2), but given that tags aren't in astronomical quantity and
	// this request isn't used very often I'm not super concerned
	for _, patch := range tags {
		err := assertTagHasProperKeys(patch)
		if err != nil {
			return err
		}

		patched := false
		for _, tag := range existingTags {
			if tag["namespace"] == patch["namespace"] && tag["key"] == patch["key"] {
				tag["value"] = patch["value"]
				patched = true
				break
			}
		}
		// Tag doesn't exist already, let's add it
		if !patched {
			existingTags = append(existingTags, patch)
		}
	}

	// Delay tag patching until it's all done (so if there's errors we don't corrupt existing state)
	s.Tags = existingTags
	return nil
}

// DeleteTags deletes the given tags from this device
func (s *Array) DeleteTags(tags []string) {
	if s.Tags == nil || len(s.Tags) == 0 {
		return
	}

	newTags := []map[string]string{}

	for _, tag := range s.Tags {
		// If it doesn't have a key, don't even add it
		if _, ok := tag["key"]; ok {
			shouldBeDeleted := false
			for _, toDelete := range tags {
				if tag["key"] == toDelete {
					shouldBeDeleted = true
				}
			}
			if !shouldBeDeleted {
				newTags = append(newTags, tag)
			}
		}
	}

	s.Tags = newTags
}

// ParseArrayFromREST parses a map in the format of the REST API call (lower_case)
// and converts to an Array struct. The reason we can't use raw JSON marshal/unmarshal
// is because we can only convert to/from one JSON format that way, and since the output
// needs more variability we'll reserve JSON marshal for the more standard case of Elastic
// input/output.
func ParseArrayFromREST(m map[string]interface{}) (Array, error) {
	toReturn := Array{}
	if _, ok := m["id"]; ok {
		toReturn.InternalID = m["id"].(string)
		validationErr := ValidateHexObjectID(toReturn.InternalID)
		if validationErr != nil {
			return Array{}, validationErr
		}
	}
	if _, ok := m["name"]; ok {
		toReturn.Name = m["name"].(string)
	}
	if _, ok := m["mgmt_endpoint"]; ok {
		toReturn.MgmtEndPoint = m["mgmt_endpoint"].(string)
	}
	if _, ok := m["api_token"]; ok {
		toReturn.APIToken = m["api_token"].(string)
	}
	if _, ok := m["status"]; ok {
		toReturn.Status = m["status"].(string)
	}
	if _, ok := m["device_type"]; ok {
		toReturn.DeviceType = m["device_type"].(string)
	}
	if _, ok := m["_as_of"]; ok {
		lastseen, err := time.Parse("2006-01-02T15:04:05.000", m["_as_of"].(string)) // yyyy-MM-dd'T'HH:mm:ss.SSS
		if err != nil {
			return Array{}, err
		}
		toReturn.Lastseen = lastseen.UTC()
	}
	if _, ok := m["_last_updated"]; ok {
		lastupdated, err := time.Parse("2006-01-02T15:04:05.000", m["_last_updated"].(string)) // yyyy-MM-dd'T'HH:mm:ss.SSS
		if err != nil {
			return Array{}, err
		}
		toReturn.Lastupdated = lastupdated.UTC()
	}

	return toReturn, nil
}

// HasRequiredPostFields checks to ensure that a storage array has all required fields filled in
func (s *Array) HasRequiredPostFields() error {
	if len(strings.TrimSpace(s.InternalID)) == 0 {
		return fmt.Errorf("Array is missing ID")
	}
	if len(strings.TrimSpace(s.Name)) == 0 {
		return fmt.Errorf("Array is missing name")
	}
	if len(strings.TrimSpace(s.MgmtEndPoint)) == 0 {
		return fmt.Errorf("Array is missing management endpoint")
	}
	if len(strings.TrimSpace(s.APIToken)) == 0 {
		return fmt.Errorf("Array is missing API token")
	}
	if len(strings.TrimSpace(s.DeviceType)) == 0 {
		return fmt.Errorf("Array is missing device type")
	}
	return nil
}

// MarshalJSON provides a custom marshal override which
// formats dates in the format Elastic expects
func (s *Array) MarshalJSON() ([]byte, error) {
	type Alias Array
	marshalled, err := json.Marshal(&struct {
		*Alias
		Lastseen    string `json:"AsOf"`
		Lastupdated string `json:"LastUpdated"`
	}{
		Lastseen:    s.Lastseen.Format(DateTimeFormat),
		Lastupdated: s.Lastupdated.Format(DateTimeFormat),
		Alias:       (*Alias)(s),
	})
	return marshalled, err
}

// UnmarshalJSON provides a custom unmarshal override which
// parses dates in the format provided by Elastic
func (s *Array) UnmarshalJSON(data []byte) error {
	type Alias Array
	aliased := &struct {
		Lastseen    string `json:"AsOf"`
		Lastupdated string `json:"LastUpdated"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := json.Unmarshal(data, &aliased); err != nil {
		return err
	}
	parsedLastSeen, err := time.Parse(DateTimeFormat, aliased.Lastseen)
	if err != nil {
		return err
	}
	s.Lastseen = parsedLastSeen.UTC()

	parsedLastUpdated, err := time.Parse(DateTimeFormat, aliased.Lastupdated)
	s.Lastupdated = parsedLastUpdated.UTC()
	return err
}

// ValidateHexObjectID checks if an internalID is a valid BSON id
func ValidateHexObjectID(internalID string) error {
	d, err := hex.DecodeString(internalID)
	if err != nil || len(d) != 12 {
		return fmt.Errorf("invalid input to ObjectIdHex: %q", internalID)
	}
	return nil
}
