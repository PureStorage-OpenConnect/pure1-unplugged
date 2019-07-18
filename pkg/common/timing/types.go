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
	"time"

	log "github.com/sirupsen/logrus"
)

// StageTimer is a utility used for timing processes, sometimes
// in multiple stages, and logging the output at the end
type StageTimer struct {
	processName string
	extraFields log.Fields
	marks       []step
}

type step struct {
	stepName string
	stepTime time.Time
}
