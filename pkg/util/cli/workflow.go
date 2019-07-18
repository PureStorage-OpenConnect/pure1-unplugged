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

package cli

import (
	"fmt"
	"github.com/briandowns/spinner"
	"time"
)

// NewWorkflowStep creates a new generic WorkFlowStep that runs the specified
// function and shows the text given.
func NewWorkflowStep(text string, fn func(ctx Context) error) WorkFlowStep {
	return &genericStep{
		fn:        fn,
		text:      text,
		stopOnErr: true,
	}
}

// NewNonTerminatingWorkflowStep creates a new generic WorkFlowStep that runs the specified
// function and shows the text given.
func NewNonTerminatingWorkflowStep(text string, fn func(ctx Context) error) WorkFlowStep {
	return &genericStep{
		fn:        fn,
		text:      text,
		stopOnErr: false,
	}
}

// NewWorkFlowEngine constructs a workflow engine instance with the specified
// workflows steps.
func NewWorkFlowEngine(workflowSteps ...WorkFlowStep) *WorkflowEngine {
	return &WorkflowEngine{steps: workflowSteps}
}

// Run will process the steps of the workflow executing the functions in order
// until finished or one func returns an error. On error the workflow is stopped
// and the error is returned. If the workflow succeeds the return value will be nil.
func (w *WorkflowEngine) Run(ctx Context) error {
	for _, step := range w.steps {
		err := w.runStep(ctx, step)
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *WorkflowEngine) runStep(ctx Context, step WorkFlowStep) error {
	var s *spinner.Spinner

	// if it isn't a silent step print out the text and start a spinner
	if !step.Silent() {
		s = spinner.New(spinner.CharSets[9], 150*time.Millisecond)
		s.Prefix = fmt.Sprintf("->>> %s ", step.Text())
		s.FinalMSG = fmt.Sprintf("--|  %s... done!\n", step.Text())
		s.Start()
	}

	// Run the func in a background thread while we wait and conditionality show a spinner
	errorChan := make(chan error)
	go func() {
		errorChan <- step.Run(ctx)
	}()

	// wait for the step to finsh
	err := <-errorChan

	// Only set if we started a spinner
	if s != nil {
		s.Stop()
	}

	return err
}
