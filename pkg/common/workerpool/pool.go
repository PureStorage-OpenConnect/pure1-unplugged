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

import (
	"time"

	log "github.com/sirupsen/logrus"
)

// CreateThreadPool instantiates a thread pool with the given number of workers and a job buffer with the given length
func CreateThreadPool(workerCount int, jobBufferLength int) Pool {
	jobChan := make(chan jobEnqueuement, jobBufferLength)
	pool := Pool{
		jobQueue: jobChan,
	}
	pool.initializeWorkers(workerCount)
	return pool
}

// Enqueue puts this job into the thread pool's queue, so that it will be run eventually
func (p *Pool) Enqueue(job Job, staleAfter time.Duration) {
	// Starts the enqueue attempt in a separate goroutine, so that this is a non-blocking call.
	go p.enqueueOrLetStale(job, staleAfter)
}

func poolWorkerThread(workerIndex int, jobQueue chan jobEnqueuement) {
	// Run forever: we have no reason to want to kill a worker thread
	for {
		// Get a job to run
		jobEntry := <-jobQueue
		// Get the description for printing purposes
		description := jobEntry.job.Description()

		startTime := time.Now().UTC()

		if startTime.After(jobEntry.staleTime) {
			log.WithFields(log.Fields{
				"worker_index": workerIndex,
				"description":  description,
				"enqueued_at":  jobEntry.enqueueTime,
				"stale_at":     jobEntry.staleTime,
			}).Debug("Job is stale, skipping")
			continue
		}

		log.WithFields(log.Fields{
			"worker_index": workerIndex,
			"start_time":   startTime,
			"description":  description,
			"wait_time":    startTime.Sub(jobEntry.enqueueTime).String(),
		}).Trace("Starting job on worker thread")

		jobEntry.job.Execute()

		endTime := time.Now().UTC()
		timeDiff := endTime.Sub(startTime)
		// Report the runtime back
		log.WithFields(log.Fields{
			"worker_index": workerIndex,
			"start_time":   startTime,
			"end_time":     endTime,
			"description":  description,
			"run_time":     timeDiff.String(),
		}).Debug("Job finished on worker thread")
	}
}

func (p *Pool) initializeWorkers(workerCount int) {
	log.WithField("size", workerCount).Trace("Spinning up thread pool")
	for i := 0; i < workerCount; i++ {
		go poolWorkerThread(i, p.jobQueue)
	}
}

// enqueueOrLetStale takes the given job and attempts to enqueue it to this worker pool. If the
// job goes stale before being enqueued, it will simply be dropped and never added.
func (p *Pool) enqueueOrLetStale(job Job, staleAfter time.Duration) {
	desc := job.Description()
	now := time.Now().UTC()
	staleAt := now.Add(staleAfter)
	log.WithFields(log.Fields{
		"description": desc,
		"stale_at":    staleAt,
	}).Trace("Enqueueing job to thread pool")
	enqueuement := jobEnqueuement{
		job:         job,
		enqueueTime: now,
		staleTime:   staleAt,
	}
	staleTimer := time.NewTimer(staleAfter)
	select {
	case p.jobQueue <- enqueuement:
		log.WithFields(log.Fields{
			"description": desc,
			"stale_at":    staleAt,
		}).Trace("Job accepted to thread pool (now in channel buffer)")
	case <-staleTimer.C: // Note (tested by experiment): If this triggers, the above push into channel never completes and the item never enters
		log.WithFields(log.Fields{
			"description": desc,
			"stale_at":    staleAt,
		}).Trace("Job turned stale before entering thread pool, dropping")
	}
}
