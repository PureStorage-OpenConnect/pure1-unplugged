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
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/upgrade"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
	"github.com/spf13/cobra"
)

func mountCommand() *cobra.Command {
	cmd := cli.BuildCommand("mount-iso",
		"Mount Pure1 Unplugged Upgrade ISO",
		"Mounts an upgrade .iso from a given file. Note that mounting this disk through your VM infrastructure as a disk is preferred, but this is an alternative if you are unable to do so.",
		upgrade.Mount,
	)

	return cmd
}

func unmountCommand() *cobra.Command {
	cmd := cli.BuildCommand("unmount-iso",
		"Unmount Pure1 Unplugged Upgrade ISO",
		"Unmounts an upgrade .iso that was previously mounted by using puctl upgrade mount-iso [file]",
		upgrade.Unmount,
	)

	return cmd
}
