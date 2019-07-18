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
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/exechelper/basicexecutor"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/fshelper"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// BuildCommand will construct a cobra command object with our version injected and a default Run method
func BuildCommand(name string, short string, long string, runFunc func(ctx Context) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:     name,
		Short:   short,
		Long:    long,
		Version: version.Get(),
		Run:     nil,
	}

	if runFunc != nil {
		cmd.Run = func(cmd *cobra.Command, args []string) {

			// Pass the helpers in from higher up?
			e := basicexecutor.New()
			fs := new(fshelper.LinuxFileSystem)

			// Run the inner function as specified (or not) by the caller
			err := runFunc(Context{
				Args:       args,
				Exec:       e,
				Filesystem: fs,
				Config:     viper.GetViper(),
			})

			// Cobra cli is kinda weird, catch the error of sub command methods and exit 1 here, otherwise it will exit 0
			if err != nil {
				fmt.Printf("\nERROR: %s\n\n", err.Error())
				os.Exit(1)
			}
		}
	}

	return cmd
}

// WaitForConfirmation will print a message and prompt for a Y/N confirmation.
// It will return the confirmation value and and error if failing to read stdin.
func WaitForConfirmation(message string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Printf("\n%s Please confirm [Y/N]: ", message)
	text, err := reader.ReadString('\n')

	if err != nil {
		return false, fmt.Errorf("failed to read stdin: %s", err)
	}

	fmt.Printf("\n\n\n")

	if strings.TrimSpace(strings.ToUpper(text)) == "Y" {
		return true, nil
	}
	return false, nil
}
