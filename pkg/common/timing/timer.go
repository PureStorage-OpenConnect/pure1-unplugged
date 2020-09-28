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

package timing

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// DebugMessage is the log message output
	DebugMessage = "StageTimer finished"
)

// NewStageTimer constructs a new StageTimer, and adds an initial stage called "start"
func NewStageTimer(processName string, extraFields log.Fields) StageTimer {
	extraFields["process_name"] = processName
	return StageTimer{
		processName: processName,
		extraFields: extraFields,
		marks: []step{
			step{
				stepName: "start",
				stepTime: time.Now().UTC(),
			},
		},
	}
}

// Stage marks a delimiting point between sections of a timer
func (s *StageTimer) Stage(stageName string) {
	s.marks = append(s.marks, step{stepName: stageName, stepTime: time.Now().UTC()})
	log.WithFields(s.extraFields).WithField("stage_name", stageName).Trace("StageTimer started new stage")
}

// Finish adds a stage called "finish" and logs all of the durations
func (s *StageTimer) Finish() {
	s.Stage("finish")
	s.logOut()
}

func (s *StageTimer) logOut() {
	for i := 0; i < len(s.marks)-1; i++ {
		from := s.marks[i]
		to := s.marks[i+1]
		s.extraFields[fmt.Sprintf("%s_to_%s", from.stepName, to.stepName)] = to.stepTime.Sub(from.stepTime).Nanoseconds()
	}
	s.extraFields["total_runtime"] = s.marks[len(s.marks)-1].stepTime.Sub(s.marks[0].stepTime).Nanoseconds()
	log.WithFields(s.extraFields).Debug(DebugMessage)
}
