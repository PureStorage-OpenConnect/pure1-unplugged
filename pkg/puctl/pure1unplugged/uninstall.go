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
	"errors"
	"fmt"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/kube"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
)

// Uninstall is the entrypoint for the workflow to reset the currently deployed Pure1 Unplugged application
func Uninstall(ctx cli.Context) error {
	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("", confirmReset),
		cli.NewWorkflowStep("Uninstalling Pure1 Unplugged helm deployment", helmDeleteWithPurge),
		cli.NewWorkflowStep("Cleaning up", deleteLeftovers),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("\nUninstall complete!\n\n")
	fmt.Printf("Elastic data is still in /mnt/elastic-* (delete directories to clear data)\n\n")

	return nil
}

func confirmReset(ctx cli.Context) error {
	confirmed, err := cli.WaitForConfirmation("This will uninstall Pure1 Unplugged. Some Data could be lost!")
	if err != nil {
		return err
	}
	if !confirmed {
		return errors.New("aborting reset")
	}
	return nil
}

func helmDeleteWithPurge(ctx cli.Context) error {
	_, err := kube.RunHelm(ctx.Exec, 600, "delete", "--purge", "pure1-unplugged")
	if err != nil {
		err = fmt.Errorf("failed to helm delete: %s", err.Error())
	}
	return err
}

// We know that the helm delete will leave behind PVC's and the namespace (helm purposely won't delete them..)
// this is our chance to remove them
func deleteLeftovers(ctx cli.Context) error {
	_, err := kube.RunKubeCTLWithNamespace(ctx.Exec, kube.Pure1UnpluggedNamespace, "delete", "pvc", "--all")
	if err != nil {
		err = fmt.Errorf("failed to delete PVC's: %s", err.Error())
		return err
	}

	_, err = kube.RunKubeCTLWithTimeout(ctx.Exec, 300, "delete", "namespace", kube.Pure1UnpluggedNamespace)
	if err != nil {
		err = fmt.Errorf("failed to delete namespace: %s", err.Error())
		return err
	}

	return nil
}
