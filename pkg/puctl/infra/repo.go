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

package infra

import (
	"fmt"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/exechelper"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
)

// EnableExternalRepos enables all yum repos except for pure1-unplugged's
func EnableExternalRepos(ctx cli.Context) error {
	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("Enabling external repos", enableExternalReposStep),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Finished!")

	return nil
}

func enableExternalReposStep(ctx cli.Context) error {
	params := exechelper.ExecParams{
		CmdName: "sed",
		CmdArgs: []string{
			"-i",
			"s/enabled=0/enabled=1/g",
			"/etc/yum.repos.d/CentOS-Base.repo",
		},
		Timeout: 5, // 5 second timeout just in case
	}
	result := ctx.Exec.RunCommand(params)
	if result.Error != nil {
		return fmt.Errorf("failed to enable external repos: %s", result.ErrBuf.String())
	}

	return nil
}

// EnablePure1UnpluggedRepo enables the Pure1 Unplugged media repo
func EnablePure1UnpluggedRepo(ctx cli.Context) error {
	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("Enabling Pure1 Unplugged repo", enablePure1UnpluggedRepoStep),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Finished!")

	return nil
}

func enablePure1UnpluggedRepoStep(ctx cli.Context) error {
	params := exechelper.ExecParams{
		CmdName: "sed",
		CmdArgs: []string{
			"-i",
			"s/enabled=0/enabled=1/g",
			"/etc/yum.repos.d/Pure1-Unplugged-Media.repo",
		},
		Timeout: 5, // 5 second timeout just in case
	}
	result := ctx.Exec.RunCommand(params)
	if result.Error != nil {
		return fmt.Errorf("failed to enable Pure1 Unplugged repo: %s", result.ErrBuf.String())
	}

	return nil
}

// DisableExternalRepos disables all yum repos except for pure1-unplugged's
func DisableExternalRepos(ctx cli.Context) error {
	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("Disabling external repos", disableExternalReposStep),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Finished!")

	return nil
}

func disableExternalReposStep(ctx cli.Context) error {
	params := exechelper.ExecParams{
		CmdName: "sed",
		CmdArgs: []string{
			"-i",
			"s/enabled=1/enabled=0/g",
			"/etc/yum.repos.d/CentOS-Base.repo",
		},
		Timeout: 5, // 5 second timeout just in case
	}
	result := ctx.Exec.RunCommand(params)
	if result.Error != nil {
		return fmt.Errorf("failed to disable external repos: %s", result.ErrBuf.String())
	}

	return nil
}

// DisablePure1UnpluggedRepo disables the Pure1 Unplugged media repo
func DisablePure1UnpluggedRepo(ctx cli.Context) error {
	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("Disabling Pure1 Unplugged repo", disablePure1UnpluggedRepoStep),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Finished!")

	return nil
}

func disablePure1UnpluggedRepoStep(ctx cli.Context) error {
	params := exechelper.ExecParams{
		CmdName: "sed",
		CmdArgs: []string{
			"-i",
			"s/enabled=1/enabled=0/g",
			"/etc/yum.repos.d/Pure1-Unplugged-Media.repo",
		},
		Timeout: 5, // 5 second timeout just in case
	}
	result := ctx.Exec.RunCommand(params)
	if result.Error != nil {
		return fmt.Errorf("failed to disable Pure1 Unplugged repo: %s", result.ErrBuf.String())
	}

	return nil
}
