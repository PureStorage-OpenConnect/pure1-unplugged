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

package upgrade

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/exechelper"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
)

const (
	mountDirectory              = "/media/Pure1-Unplugged_x86_64"
	loopFileHandle              = "/dev/loop7"
	pure1UnpluggedMediaRepoPath = "/etc/yum.repos.d/Pure1 Unplugged-Media.repo"
)

// Mount is the entry point for the workflow to mount a Pure1 Unplugged upgrade iso.
func Mount(ctx cli.Context) error {
	if len(ctx.Args) != 1 {
		return fmt.Errorf("Usage: puctl upgrade mount-iso [Pure1 Unplugged iso file]")
	}

	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("Detaching loop file handle if it exists", unmountLoopFileHandler),
		cli.NewWorkflowStep("Creating mount directory", createMountDirectory),
		cli.NewWorkflowStep("Creating loop file handle", createLoopFileHandler),
		cli.NewWorkflowStep("Mounting iso in directory", mountIsoDirectory),
		cli.NewWorkflowStep("Enabling Pure1 Unplugged yum repo", enableYumRepo),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Finished!")

	return nil
}

// Unmount is the entry point for the workflow to unmount a Pure1 Unplugged upgrade iso.
func Unmount(ctx cli.Context) error {
	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("Disabling Pure1 Unplugged yum repo", disableYumRepo),
		cli.NewWorkflowStep("Unmounting iso folder", unmountIsoFolder),
		cli.NewWorkflowStep("Detaching loop file handle if it exists", unmountLoopFileHandler),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Finished!")

	return nil
}

func unmountIsoFolder(ctx cli.Context) error {
	// Ignoring all errors and output here: we're literally running this to unmount only if it exists
	ctx.Exec.RunCommandWithOutput("umount", mountDirectory)

	return nil
}

func unmountLoopFileHandler(ctx cli.Context) error {
	listResult, err := ctx.Exec.RunCommandWithOutput("losetup", "-a")
	if err != nil {
		return fmt.Errorf("Error checking list of attached loop devices: %v", err)
	}

	exists := false

	resultLines := strings.Split(listResult, "\n")
	for _, line := range resultLines {
		if strings.HasPrefix(strings.TrimSpace(line)+":", loopFileHandle) {
			exists = true
			break
		}
	}

	if !exists {
		return nil
	}

	result := ctx.Exec.RunCommand(exechelper.ExecParams{
		CmdName: "losetup",
		CmdArgs: []string{
			"-d",
			loopFileHandle,
		},
		Timeout: 5,
	})

	if result.Error != nil {
		return fmt.Errorf("Failed to detach loop file handler: %s", result.ErrBuf.String())
	}

	return nil
}

func createMountDirectory(ctx cli.Context) error {
	return os.MkdirAll(mountDirectory, 0777) // 777 permissions should be fine since the upgrades themselves already need root
}

func createLoopFileHandler(ctx cli.Context) error {
	path, err := filepath.Abs(ctx.Args[0]) // Get the running directory
	if err != nil {
		return err
	}

	result := ctx.Exec.RunCommand(exechelper.ExecParams{
		CmdName: "losetup",
		CmdArgs: []string{
			loopFileHandle,
			path,
		},
		Timeout: 5,
	})

	if result.Error != nil {
		return fmt.Errorf("Failed to create loop file handler: %s", result.ErrBuf.String())
	}

	return nil
}

func mountIsoDirectory(ctx cli.Context) error {
	result := ctx.Exec.RunCommand(exechelper.ExecParams{
		CmdName: "mount",
		CmdArgs: []string{
			"-o",
			"ro",
			loopFileHandle,
			mountDirectory,
		},
		Timeout: 15,
	})

	if result.Error != nil {
		return fmt.Errorf("Failed to mount iso: %s", result.ErrBuf.String())
	}

	return nil
}

func enableYumRepo(ctx cli.Context) error {
	result := ctx.Exec.RunCommand(exechelper.ExecParams{
		CmdName: "sed",
		CmdArgs: []string{
			"-i",
			"s/enabled=0/enabled=1/g",
			pure1UnpluggedMediaRepoPath,
		},
		Timeout: 5,
	})

	if result.Error != nil {
		return fmt.Errorf("Failed to enable Pure1 Unplugged yum repo: %s", result.ErrBuf.String())
	}

	return nil
}

func disableYumRepo(ctx cli.Context) error {
	result := ctx.Exec.RunCommand(exechelper.ExecParams{
		CmdName: "sed",
		CmdArgs: []string{
			"-i",
			"s/enabled=1/enabled=0/g",
			pure1UnpluggedMediaRepoPath,
		},
		Timeout: 5,
	})

	if result.Error != nil {
		return fmt.Errorf("Failed to disable Pure1 Unplugged yum repo: %s", result.ErrBuf.String())
	}

	return nil
}
