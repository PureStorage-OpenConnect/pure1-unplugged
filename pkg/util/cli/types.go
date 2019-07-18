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
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/exechelper"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/fshelper"
	"github.com/spf13/viper"
)

// Context has all the context/config required for CLI command handlers to operate
type Context struct {
	Args       []string
	Exec       exechelper.Executor
	Filesystem fshelper.FileSystem
	Config     *viper.Viper
	CmdContext interface{}
}

// WorkflowEngine holds onto a series of workflow steps and can process them with error handling.
type WorkflowEngine struct {
	steps []WorkFlowStep
}

// WorkFlowStep defines a step in a CLI workflow with some helper to run the step
// and provide info in the UI (terminal, usually)
type WorkFlowStep interface {
	// Run should run the method and *not* print anything on stdout, only logging
	Run(ctx Context) error
	// Text is the description or title or whatever we want to show to the user as we print progress through
	// the workflow.
	Text() string
	// Silent returns whether or not the step should be considered as silent (no Text() or progress will be shown)
	Silent() bool
}
