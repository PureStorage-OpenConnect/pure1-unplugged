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
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/infra"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
	"github.com/spf13/cobra"
)

func repoCommand() *cobra.Command {
	subcmds := []*cobra.Command{
		enableExternalRepoCommand(),
		disableExternalRepoCommand(),
		enablePure1UnpluggedRepoCommand(),
		disablePure1UnpluggedRepoCommand(),
	}

	cmd := cli.BuildCommand("repo", "Wrapper to manage repos", "CLI tools to manage Yum repos for both Pure1 Unplugged and external packages. Use this if you need to install a package that isn't available.", nil)
	cmd.AddCommand(subcmds...)

	return cmd
}

func enableExternalRepoCommand() *cobra.Command {
	cmd := cli.BuildCommand("enable-external", "Enable external repos", "Enables all external repos to allow installing packages from outside the provided set with Pure1 Unplugged. It is recommended to disable these after you're done to ensure compatibility is maintained.", infra.EnableExternalRepos)

	return cmd
}

func disableExternalRepoCommand() *cobra.Command {
	cmd := cli.BuildCommand("disable-external", "Disable external repos", "Disables all external repos to prohibit installing packages from outside the provided set with Pure1 Unplugged.", infra.DisableExternalRepos)

	return cmd
}

func enablePure1UnpluggedRepoCommand() *cobra.Command {
	cmd := cli.BuildCommand("enable-pure1-unplugged", "Enable Pure1 Unplugged repo", "Enables the Pure1 Unplugged repo to allow updating. Note that this requires a Pure1 Unplugged image to be mounted in /media/Pure1-Unplugged_x86_64.", infra.EnablePure1UnpluggedRepo)

	return cmd
}

func disablePure1UnpluggedRepoCommand() *cobra.Command {
	cmd := cli.BuildCommand("disable-pure1-unplugged", "Disable Pure1 Unplugged repo", "Disables the Pure1 Unplugged repo. This is helpful if it's enabled and you can't update other packages because there's no Pure1 Unplugged repo present.", infra.DisablePure1UnpluggedRepo)

	return cmd
}
