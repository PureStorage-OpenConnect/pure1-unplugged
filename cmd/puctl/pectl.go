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

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/cmd/puctl/infra"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/cmd/puctl/pure1unplugged"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/cmd/puctl/upgrade"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/config"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/version"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	defaultConfig = "/etc/pure1-unplugged/config.yaml"
)

// config file for puctl
var cfgFile string

var logFile string

func initRootCommand() *cobra.Command {
	subCmds := []*cobra.Command{
		infra.Command(),
		pure1unplugged.Command(),
		upgrade.Command(),
	}

	desc := fmt.Sprintf("Pure1 Unplugged CLI - %s", version.Get())
	rootCmd := cli.BuildCommand("puctl", desc, desc, nil)

	// Hook the sub commands onto the root command
	rootCmd.AddCommand(subCmds...)

	// Add in some sweet bash completion generator commands
	completionCmd := getCompletionCmd(rootCmd)
	rootCmd.AddCommand(completionCmd)

	rootCmd.PersistentFlags().StringVarP(&logFile, "logfile", "l", "", "path to log file, if set to '' (default) a random tmp file is used")
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", defaultConfig, "path to config.yaml")

	return rootCmd
}

// Setup logging and config
func initialize() {
	if logFile == "" {
		logFile = filepath.Join("/tmp/pure1-unplugged/", util.RandomString(10)+".log")
	}

	fmt.Printf("\nLogging to %s\n\n", logFile)

	logDir := filepath.Dir(logFile)
	err := os.MkdirAll(logDir, os.ModeDir)
	if err != nil {
		fmt.Printf("Unable to create log directory %s: %s\n", logDir, err.Error())
		os.Exit(1)
	}

	logFileH, err := os.Create(logFile)
	if err != nil {
		fmt.Printf("Unable to open log file %s: %s\n", logDir, err.Error())
		os.Exit(1)
	}

	// Add a hook to close our log file
	logrus.RegisterExitHandler(func() {
		if logFileH != nil {
			logFileH.Close()
		}
	})

	// Set all log output to go to our file, disable colors, and ensure timestamps are present
	logrus.SetOutput(logFileH)
	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true, FullTimestamp: true})

	// Always log at debug level
	logrus.SetLevel(logrus.DebugLevel)

	// See config options and defaults in pkg/puctl/config/config.go
	// We just do the viper-related stuff in here
	for k, v := range config.Defaults {
		viper.SetDefault(k, v)
	}

	logrus.Debugf("Using config: %s\n", cfgFile)
	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err != nil {
		logrus.Errorf("Can't read config: %s", err)
		os.Exit(1)
	}
}

// getCompletionCmd adds a 'completion' sub command to puctl which will generate bash command
// autocomplete script(s).
func getCompletionCmd(rootCmd *cobra.Command) *cobra.Command {
	var completionCmd = &cobra.Command{
		Use:   "completion OUTPUT_FILE",
		Args:  cobra.MinimumNArgs(1),
		Short: "Generates bash completion scripts",
		Long: `To load completion run

. <(cat OUTPUT_FILE)

To configure your bash shell to load completions for each session add to your bashrc

# ~/.bashrc or ~/.profile
. <(cat OUTPUT_FILE)

Alternatively just copy it to

/etc/bash_completion.d/

`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Saving completion script to " + args[0])
			err := rootCmd.GenBashCompletionFile(args[0])
			if err != nil {
				fmt.Printf("ERROR saving completion script: %s\n", err.Error())
			}
		},
	}
	return completionCmd
}
