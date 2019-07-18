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
	"regexp"
	"strings"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/kube"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
)

// Upgrade is the entrypoint for the workflow to upgrade the currently running Pure1 Unplugged application
func Upgrade(ctx cli.Context) error {
	// Pass through a version variable that gets populated in the first step.
	currentVersion := ""
	ctx.CmdContext = &currentVersion

	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("", confirmUpgrade),
		cli.NewWorkflowStep("Unpacking container images", loadDockerImages),
		cli.NewWorkflowStep("Upgrading Pure1 Unplugged app(s)", helmUpgrade),
		cli.NewWorkflowStep("Waiting to let the upgrade settle", settlingDelay),
		cli.NewWorkflowStep("Restarting Pure1 Unplugged auth services", authCycle),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("\n Upgrade complete!\n\n")

	return nil
}

func confirmUpgrade(ctx cli.Context) error {
	confirmed, err := cli.WaitForConfirmation("This will upgrade Pure1 Unplugged, service interruption will likely occur (unless already at latest version)")
	if err != nil {
		return err
	}
	if !confirmed {
		return errors.New("aborting upgrade")
	}
	return nil
}

func helmUpgrade(ctx cli.Context) error {
	chartFile, err := findHelmChart(ctx)
	if err != nil {
		return err
	}

	_, err = kube.RunHelm(
		ctx.Exec,
		900,
		"upgrade",
		"--force",
		"--wait",
		"--values=/etc/pure1-unplugged/config.yaml",
		"pure1-unplugged",
		chartFile,
	)

	if err != nil {
		return fmt.Errorf("failed to helm upgrade pure1-unplugged deployment: %s", err.Error())
	}

	return nil
}

func settlingDelay(ctx cli.Context) error {
	overallTimeout := time.NewTimer(time.Second * 90)

	for {
		out, err := kube.RunKubeCTLWithTimeout(ctx.Exec, 30, "get", "pods", "-n", "pure1-unplugged")
		if err != nil {
			return fmt.Errorf("Error getting pods list: %s", err.Error())
		}

		if !strings.Contains(strings.ToLower(out), "terminating") {
			return nil // All pods have started (all dying pods are dead), we're safe to restart auth services
		}

		// Wait a bit longer
		delay := time.NewTimer(time.Second * 5)
		select {
		case <-delay.C:
			continue
		case <-overallTimeout.C:
			return fmt.Errorf("Pods took too long to settle")
		}
	}
}

func authCycle(ctx cli.Context) error {
	out, err := kube.RunKubeCTLWithTimeout(ctx.Exec, 30, "get", "pods", "-n", "pure1-unplugged")
	if err != nil {
		return fmt.Errorf("Error getting pods list: %s", err.Error())
	}

	authExp := regexp.MustCompile(`pure1-unplugged-auth-server\S*`)
	authServers := authExp.FindAllString(out, -1)

	if authServers == nil {
		return fmt.Errorf("Couldn't find any instances of auth server to kill")
	}

	dexExp := regexp.MustCompile(`pure1-unplugged-dex\S*`)
	dexes := dexExp.FindAllString(out, -1)

	if dexes == nil {
		return fmt.Errorf("Couldn't find any instances of dex to kill")
	}

	allPodsToKill := append(authServers, dexes...)

	killArgs := []string{"delete", "pods", "-n", "pure1-unplugged"}
	killArgs = append(killArgs, allPodsToKill...)

	_, err = kube.RunKubeCTLWithTimeout(ctx.Exec, 30, killArgs...)
	if err != nil {
		return fmt.Errorf("Failed to delete pods: %s", err.Error())
	}

	return nil
}
