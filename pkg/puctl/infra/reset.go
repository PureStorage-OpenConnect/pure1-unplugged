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
	"github.com/pkg/errors"
)

// Reset is the entrypoint to reset Pure1 Unplugged infrastructure on the current host
func Reset(ctx cli.Context) error {
	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("", confirmReset),
		cli.NewWorkflowStep("Resetting Kubernetes installation, all containers will be removed ", resetKubernetes),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("\nReset complete! A reboot may be required to reset all network configuration.\n")

	return nil
}

func confirmReset(ctx cli.Context) error {
	confirmed, err := cli.WaitForConfirmation("This will reset the app infrastructure (Kubernetes) and data may be lost.")
	if err != nil {
		return err
	}
	if !confirmed {
		return errors.New("aborting reset")
	}
	return nil
}

func resetKubernetes(ctx cli.Context) error {
	params := exechelper.ExecParams{
		CmdName: "kubeadm",
		CmdArgs: []string{
			"reset",
			"--force",
		},
		Timeout: 180,
	}
	result := ctx.Exec.RunCommand(params)
	if result.Error != nil {
		return fmt.Errorf("failed to kubeadm reset: %s", result.ErrBuf.String())
	}

	return nil
}
