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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/config"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/exechelper"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/kube"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

// Upgrade is the entrypoint for starting the workflow to upgrade Pure1 Unplugged infrastructure
func Upgrade(ctx cli.Context) error {

	// Pass through a version variable that gets populated in the first step.
	currentVersion := ""
	ctx.CmdContext = &currentVersion

	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("Checking kubernetes cluster version", checkKubernetesVersionForUpgrade),
		cli.NewWorkflowStep("", confirmUpgrade),
		cli.NewWorkflowStep("Updating system state for newer kubernetes version (restarts some services)", runPrereqPlaybook),
		cli.NewWorkflowStep("Waiting for critical system pods to come back online", waitForCriticalSystemPods),
		cli.NewWorkflowStep("Unpacking container images", loadDockerImages),
		cli.NewWorkflowStep("Upgrading Kubernetes control plane", upgradeKubernetesControlPlane),
		cli.NewWorkflowStep("Restarting kubelet", restartKubelet),
		cli.NewWorkflowStep("Waiting for initial system pods to come back online", waitForInitialSystemPods),
		cli.NewWorkflowStep("Upgrading CNI plugins", installCalico),
		cli.NewWorkflowStep("Waiting for remaining critical system pods to come back online ", waitForCriticalSystemPods),
		cli.NewWorkflowStep("Upgrading Helm", helmInit),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("\n Upgrade complete!\n")

	return nil
}

func confirmUpgrade(ctx cli.Context) error {
	version := ctx.CmdContext.(*string)

	confirmationMessage := "This will upgrade the app infrastructure, service interruption will occur (non-HA cluster)."
	if *version == "" {
		confirmationMessage = "WARNING!! Unable to detect current kubernetes version! Attempting to continue with the upgrade is *NOT* recommended. Proceed anyway?"
	}

	confirmed, err := cli.WaitForConfirmation(confirmationMessage)
	if err != nil {
		return err
	}
	if !confirmed {
		return errors.New("aborting upgrade")
	}
	return nil
}

func upgradeKubernetesControlPlane(ctx cli.Context) error {
	if *(ctx.CmdContext.(*string)) == kube.KubeVersion {
		log.Infof("Skipping kubeadm upgrade! Version is already up to date.")
		return nil
	}

	// NOTE: Ensure that on upgrade the kubeconfig template still works!! Some changes might be required that are
	//  specific to the new version! This would require re-generating the config like what is done in `init`

	params := exechelper.ExecParams{
		CmdName: "kubeadm",
		CmdArgs: []string{
			"upgrade",
			"apply",
			kube.KubeVersion,
			"-y",
			"--config=" + filepath.Join(ctx.Config.GetString(config.Pure1UnpluggedInstallDirKey), "/infra/kubeadm/kubeadm-init-config.yaml"),
		},
		Timeout: 600,
	}
	result := ctx.Exec.RunCommand(params)
	if result.Error != nil {
		return fmt.Errorf("failed to kubeadm upgrade: %s", result.ErrBuf.String())
	}

	return nil
}

func checkKubernetesVersionForUpgrade(ctx cli.Context) error {
	out, err := kube.RunKubeCTL(ctx.Exec, "version", "--output=json")
	currentVersion := ""
	if err == nil {
		/*
			Expect something like:

			{
			  "clientVersion": {
				"major": "1",
				"minor": "13",
				"gitVersion": "v1.13.4",
				"gitCommit": "c27b913fddd1a6c480c229191a087698aa92f0b1",
				"gitTreeState": "clean",
				"buildDate": "2019-02-28T13:37:52Z",
				"goVersion": "go1.11.5",
				"compiler": "gc",
				"platform": "linux/amd64"
			  },
			  "serverVersion": {
				"major": "1",
				"minor": "13",
				"gitVersion": "v1.13.3",
				"gitCommit": "721bfa751924da8d1680787490c54b9179b1fed0",
				"gitTreeState": "clean",
				"buildDate": "2019-02-01T20:00:57Z",
				"goVersion": "go1.11.5",
				"compiler": "gc",
				"platform": "linux/amd64"
			  }
			}
		*/
		versionInfo := map[string]map[string]string{}
		err = json.Unmarshal([]byte(out), &versionInfo)
		if err != nil {
			log.Errorf("Failed to unmarshal verison info: %s", err.Error())
		} else {
			log.Debugf("kubernetes version info: %+v", versionInfo)
			serverVerion, ok := versionInfo["serverVersion"]
			if !ok {
				log.Errorf("Unable to get serverVersion from version info '%s'", versionInfo)
			} else {
				currentVersion = serverVerion["gitVersion"]
			}
		}
	}

	// Set our context (converted to a string pointer) to be the version we found
	*(ctx.CmdContext.(*string)) = currentVersion

	if kube.KubeVersion == currentVersion {
		log.Warningf("Current version '%s' == upgrade target version '%s'", currentVersion, kube.KubeVersion)
	}

	return nil
}

func restartKubelet(ctx cli.Context) error {
	_, err := ctx.Exec.RunCommandWithOutput("systemctl", "daemon-reload")
	if err != nil {
		return fmt.Errorf("failed to reload systemd configurations: %s", err.Error())
	}

	_, err = ctx.Exec.RunCommandWithOutput("systemctl", "restart", "kubelet")
	if err != nil {
		return fmt.Errorf("failed to restart kubelet: %s", err.Error())
	}

	return nil
}
