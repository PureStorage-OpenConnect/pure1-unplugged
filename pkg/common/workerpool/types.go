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

package workerpool

import "time"

// Pool represents a thread pool that distributes jobs
type Pool struct {
	jobQueue chan jobEnqueuement
}

type jobEnqueuement struct {
	job         Job
	enqueueTime time.Time
	staleTime   time.Time
}

// Job represents a task that can be executed by a goroutine pool. Errors in execution should be logged but
// not returned, there is no retry mechanism in place.
type Job interface {
	Execute()
	// Description is used to provide a summary of this job (used mainly for logging purposes)
	Description() string
}
