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

package exechelper

import (
	"bytes"
)

// Executor is the interface for executing commands.
type Executor interface {
	RunCommandWithOutputRaw(name string, args ...string) ([]byte, error)
	RunCommandWithOutput(name string, args ...string) (string, error)
	RunCommandWithOutputTimed(command []string, timeoutSeconds int) (string, error)

	RunCommand(params ExecParams) ExecResult
}

// ExecParams parameters to execute a command
type ExecParams struct {
	CmdName string
	CmdArgs []string
	Timeout int
}

// ExecResult result of executing a command
type ExecResult struct {
	OutBuf   *bytes.Buffer
	ErrBuf   *bytes.Buffer
	ExitCode int
	Error    error
}
