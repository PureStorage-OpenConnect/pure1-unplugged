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

package pure1unplugged

import (
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/pure1unplugged"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
	"github.com/spf13/cobra"
)

func upgradeCommand() *cobra.Command {
	cmd := cli.BuildCommand(
		"upgrade",
		"Upgrade Pure1 Unplugged Application",
		"Upgrade the currently running Pure1 Unplugged application",
		pure1unplugged.Upgrade,
	)

	return cmd
}
