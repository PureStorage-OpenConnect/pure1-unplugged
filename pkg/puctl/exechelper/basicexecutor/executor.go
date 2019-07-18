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

package basicexecutor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"syscall"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/exechelper"
	log "github.com/sirupsen/logrus"
)

type basicExecutor struct {
	formatRegex *regexp.Regexp
}

const (
	defaultExecTimeout = 30

	exitCodeTimeout    = 124
	exitCodeErrDefault = 1
	exitCodeSuccess    = 0
)

// New creates a new basicExecutor instance, which implements
// exechelper.Executor interface
func New() exechelper.Executor {
	return &basicExecutor{}
}

// formatOutput can be used to format a raw output []byte into a single line of text
// better suited for log output.
func (e *basicExecutor) formatOutput(output []byte) string {
	return e.squashString(string(output[:]))
}

func (e *basicExecutor) squashString(str string) string {
	if e.formatRegex == nil {
		e.formatRegex = regexp.MustCompile("[\t\n\r]+")
	}
	return e.formatRegex.ReplaceAllString(str, " ")
}

func (e *basicExecutor) runCommand(timeout int, name string, args ...string) ([]byte, error) {
	log.WithFields(log.Fields{
		"command": name,
		"args":    args,
		"timeout": timeout,
	}).Debug("Running command")

	// Create a new timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(timeout))
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	out, err := cmd.CombinedOutput()

	if ctx.Err() == context.DeadlineExceeded {
		err = fmt.Errorf("Command %s %s timed out after %d seconds", name, args, timeout)
		return out, err
	}

	// If there's no context error, we know the command completed, with or without error
	if err != nil {
		err = errors.New(e.squashString(err.Error()))
	}
	return out, err
}

func (e *basicExecutor) RunCommandWithOutputRaw(name string, args ...string) ([]byte, error) {
	out, err := e.runCommand(defaultExecTimeout, name, args...)
	log.WithFields(log.Fields{
		"command": name,
		"args":    args,
		"error":   err,
		"output":  "****",
	}).Debug("Finished running command (output redacted)")
	return out, err
}

// RunCommandWithOutput runs a command to completion, and returns
// stdout and stderr in one byte array along with any associated error
func (e *basicExecutor) RunCommandWithOutput(name string, args ...string) (string, error) {
	out, err := e.runCommand(defaultExecTimeout, name, args...)
	formattedOutput := ""
	if out != nil {
		formattedOutput = e.formatOutput(out)
	}
	log.WithFields(log.Fields{
		"command": name,
		"args":    args,
		"output":  formattedOutput,
		"error":   err,
	}).Debug("Finished running command")
	return formattedOutput, err
}

// RunCommandWithOutputTimed is same as ExecCommandTraceOutput, with
// a timeout ...
func (e *basicExecutor) RunCommandWithOutputTimed(command []string, timeoutSeconds int) (string, error) {
	if len(command) <= 0 {
		return "", errors.New("Unable to run command, command is empty")
	}
	commandName := command[0]

	var args []string
	if len(command) > 1 {
		args = command[1:]
	}

	out, err := e.runCommand(timeoutSeconds, commandName, args...)
	formattedOutput := ""
	if out != nil {
		formattedOutput = e.formatOutput(out)
	}

	log.WithFields(log.Fields{
		"command": commandName,
		"args":    args,
		"timeout": timeoutSeconds,
		"output":  formattedOutput,
		"error":   err,
	}).Debug("Finished running command")
	return formattedOutput, err
}

// RunCommand run a command, and get result
func (e *basicExecutor) RunCommand(params exechelper.ExecParams) exechelper.ExecResult {
	log.WithFields(log.Fields{"params": params}).Debug("Running command")

	// Create a new timeout context
	if params.Timeout == 0 {
		params.Timeout = defaultExecTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(params.Timeout))
	defer cancel()

	outbuf, errbuf := bytes.NewBufferString(""), bytes.NewBufferString("")
	cmd := exec.CommandContext(ctx, params.CmdName, params.CmdArgs...)
	cmd.Stdout = outbuf
	cmd.Stderr = errbuf
	err := cmd.Run()

	result := exechelper.ExecResult{
		OutBuf:   outbuf,
		ErrBuf:   errbuf,
		ExitCode: exitCodeSuccess,
		Error:    err,
	}

	if ctx.Err() == context.DeadlineExceeded {
		result.ExitCode = exitCodeTimeout
		result.Error = fmt.Errorf("Command %s %s timed out after %d seconds", params.CmdName, params.CmdArgs, params.Timeout)
		err = result.Error
	}

	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			result.ExitCode = ws.ExitStatus()
		} else {
			// failed to get exit code, use default code
			result.ExitCode = exitCodeErrDefault
		}
		result.Error = errors.New(e.squashString(err.Error()))
	}

	log.WithFields(log.Fields{
		"command": params.CmdName,
		"args":    params.CmdArgs,
		"timeout": params.Timeout,
		"stdout":  outbuf.String(),
		"stderr":  errbuf.String(),
		"error":   err,
	}).Debug("Finished running command")

	return result
}
