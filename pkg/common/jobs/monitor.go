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

package jobs

import (
	"fmt"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/workerpool"
	log "github.com/sirupsen/logrus"
)

// Type guard: ensure this implements the interface
var _ workerpool.Job = (*MonitorCheckJob)(nil)

// Description gets a string description of this job
func (m *MonitorCheckJob) Description() string {
	return fmt.Sprintf("Monitor check job for device ID %s (display name %s) at %s", m.DeviceInfo.ID, m.DeviceInfo.Name, m.DeviceInfo.MgmtEndpoint)
}

// Execute attempts to connect to the device and pulls metadata, pushing it to the API server
func (m *MonitorCheckJob) Execute() {
	if m.DeviceInfo == nil {
		log.Error("Tried to monitor check a nil array, stopping")
		return
	}
	if m.DeviceFactory == nil {
		log.WithFields(m.DeviceInfo.GetLogFields(true)).Error("Tried to monitor an array with a nil factory, stopping")
		return
	}
	if m.Metadata == nil {
		log.WithFields(m.DeviceInfo.GetLogFields(true)).Error("Tried to monitor an array with a nil metadata connection, stopping (nowhere to put data")
		return
	}
	backend, err := m.DeviceFactory.InitializeCollector(m.DeviceInfo)
	if err != nil {
		log.WithFields(m.DeviceInfo.GetLogFields(true)).WithError(err).Error("Error initializing array backend")
		patchErr := m.Metadata.Patch(m.DeviceInfo.ID, &resources.ArrayPatchInfo{
			Status: fmt.Sprintf("Unable to connect. Error: %s", err),
		})
		if patchErr != nil {
			log.WithFields(m.DeviceInfo.GetLogFields(true)).WithFields(log.Fields{
				"connection_err": err,
				"patch_err":      patchErr,
			}).Error("Unable to patch array status for connection error: status may not be updated properly on server")
			// Patch failed, honestly not much we can do past here to recover
		}
		return
	}

	model, err := backend.GetArrayModel()
	if err != nil {
		log.WithFields(m.DeviceInfo.GetLogFields(true)).WithError(err).Error("Error making model request to array backend")
		patchErr := m.Metadata.Patch(m.DeviceInfo.ID, &resources.ArrayPatchInfo{
			Status: fmt.Sprintf("Unable to connect. Error: %s", err),
		})
		if patchErr != nil {
			log.WithFields(m.DeviceInfo.GetLogFields(true)).WithFields(log.Fields{
				"connection_err": err,
				"patch_err":      patchErr,
			}).Error("Unable to patch device status for connection error: status may not be updated properly on server")
			// Patch failed, honestly not much we can do past here to recover
		}
		return
	}

	version, err := backend.GetArrayVersion()
	if err != nil {
		log.WithFields(m.DeviceInfo.GetLogFields(true)).WithError(err).Error("Error making version request to array backend")
		patchErr := m.Metadata.Patch(m.DeviceInfo.ID, &resources.ArrayPatchInfo{
			Status: fmt.Sprintf("Unable to connect. Error: %s", err),
		})
		if patchErr != nil {
			log.WithFields(m.DeviceInfo.GetLogFields(true)).WithFields(log.Fields{
				"connection_err": err,
				"patch_err":      patchErr,
			}).Error("Unable to patch device status for connection error: status may not be updated properly on server")
			// Patch failed, honestly not much we can do past here to recover
		}
		return
	}

	err = m.Metadata.Patch(m.DeviceInfo.ID, &resources.ArrayPatchInfo{
		Status:  "Connected",
		Model:   model,
		Version: version,
		AsOf:    time.Now().UTC().Format("2006-01-02T15:04:05.000"),
	})
	if err != nil {
		log.WithFields(m.DeviceInfo.GetLogFields(true)).WithFields(log.Fields{
			"array_model":   model,
			"array_version": version,
		}).WithError(err).Error("Error patching information for array: may not be updated properly on server")
		return
	}
	log.WithFields(m.DeviceInfo.GetLogFields(true)).WithFields(log.Fields{
		"array_model":   model,
		"array_version": version,
	}).Trace("Finished monitor checking array and patched successfully")
}
