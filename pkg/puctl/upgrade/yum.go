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
	"path"
	"strings"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
)

const (
	pure1UnpluggedRepoName    = "pure1-unplugged-media"
	pure1UnpluggedPackageName = "pure1-unplugged.x86_64"
)

var (
	lookForFiles = []string{
		"TRANS.TBL",
	}
	lookForDirectories = []string{
		"EFI",
		"images",
		"isolinux",
		"LiveOS",
		"Packages",
		"repodata",
	}
	mountedVersion = ""
)

// YumUpgrade is the entry point for the workflow to upgrade the Pure1 Unplugged package once an upgrade .iso is mounted.
func YumUpgrade(ctx cli.Context) error {
	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("Verifying that mounted iso looks correct", verifyIsoFolder),
		cli.NewWorkflowStep("Cleaning yum metadata cache", yumCleanMetadata),
		cli.NewWorkflowStep("Verifying that the pure1-unplugged-media repo is enabled", verifyRepoActive),
		cli.NewWorkflowStep("Determining version of mounted upgrade iso", determineVersion),
		cli.NewWorkflowStep("Running yum upgrade (or downgrade)", doUpgrade),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Finished!")

	return nil
}

func verifyIsoFolder(ctx cli.Context) error {
	for _, file := range lookForFiles {
		fullPath := path.Join(mountDirectory, file)
		info, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("Mounted iso does not appear to be valid: missing file %s", fullPath)
		}
		if info.IsDir() {
			return fmt.Errorf("Mounted iso does not appear to be valid: %s is a directory, not a file", fullPath)
		}
	}
	for _, dir := range lookForDirectories {
		fullPath := path.Join(mountDirectory, dir)
		info, err := os.Stat(fullPath)
		if os.IsNotExist(err) {
			return fmt.Errorf("Mounted iso does not appear to be valid: missing directory %s", fullPath)
		}
		if !info.IsDir() {
			return fmt.Errorf("Mounted iso does not appear to be valid: %s is a file, not a directory", fullPath)
		}
	}
	// All files and directories are present
	return nil
}

func yumCleanMetadata(ctx cli.Context) error {
	_, err := ctx.Exec.RunCommandWithOutput("yum", "clean", "metadata")
	if err != nil {
		return fmt.Errorf("Error cleaning yum metadata: %v", err)
	}

	return nil
}

func verifyRepoActive(ctx cli.Context) error {
	listResult, err := ctx.Exec.RunCommandWithOutput("yum", "repolist", "-q")
	if err != nil {
		return fmt.Errorf("Error checking yum repolist: %v", err)
	}

	if !strings.Contains(listResult, pure1UnpluggedRepoName) {
		return fmt.Errorf("yum repolist does not show %s as an active repo. Make sure a valid upgrade .iso is mounted using 'puctl upgrade mount-iso [file]' or by attaching the .iso as a disk", pure1UnpluggedRepoName)
	}

	return nil
}

func determineVersion(ctx cli.Context) error {
	versionResult, err := ctx.Exec.RunCommandWithOutput("yum", "list", "available", pure1UnpluggedPackageName, "--show-duplicates", "-q", "--color=never")
	if err != nil {
		// Note that no package being found *is* an error which works out well here
		return fmt.Errorf("Error checking yum list for package: %v", err)
	}

	versionResult = strings.Replace(versionResult, "Available Packages", "", -1)

	versionLines := strings.Split(versionResult, "\n")

	for _, line := range versionLines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, pure1UnpluggedPackageName) {
			continue
		}
		lineSplit := strings.Split(trimmed, " ")
		textOnly := []string{}
		for _, part := range lineSplit {
			if len(strings.TrimSpace(part)) != 0 {
				textOnly = append(textOnly, part)
			}
		}
		if len(textOnly) != 3 {
			// This isn't a valid line from repolist, ignore it
			continue
		}
		mountedVersion = strings.TrimSpace(textOnly[1]) // Second column is actual version ID
		return nil
	}

	return fmt.Errorf("Couldn't find package %s in list of available packages", pure1UnpluggedPackageName)
}

func doUpgrade(ctx cli.Context) error {
	if len(strings.TrimSpace(mountedVersion)) == 0 {
		return fmt.Errorf("Mounted version doesn't appear to have been detected properly")
	}

	fullPackageName := fmt.Sprintf("pure1-unplugged-%s.x86_64", strings.TrimSpace(mountedVersion))

	// Try downgrade. Note that an attempted downgrade to a version that's an upgrade isn't an *error*, it's only indicated in stdout, which actually works well for us.
	downgradeResult, err := ctx.Exec.RunCommandWithOutputTimed([]string{"yum", "downgrade", fullPackageName, "-y"}, 90)
	if err != nil {
		return fmt.Errorf("Error running yum downgrade: %v", err)
	}

	if !strings.Contains(strings.ToLower(downgradeResult), "only upgrade available on package") {
		// Upgrade was a success! We can stop here
		return nil
	}

	// We can't downgrade to this, let's try an upgrade
	_, err = ctx.Exec.RunCommandWithOutputTimed([]string{"yum", "upgrade", fullPackageName, "-y"}, 90)
	if err != nil {
		return fmt.Errorf("Error running yum upgrade: %v", err)
	}

	return nil
}
