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
	"io/ioutil"
	"os"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/kube"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
	log "github.com/sirupsen/logrus"
)

const (
	outputDirectoryPrefix = "pure1-unplugged-logs-"
	// 600 perms so that only root can read, nobody else can. Since this contains possibly "sensitive" information (even though API tokens are purged) we wanna be careful
	outputPermissions = 0600
)

var fullOutputDirectory string

// PackLogs is the entry point to package logs in the kubernetes cluster
func PackLogs(ctx cli.Context) error {
	// Steps:
	// 1. Run describe all on cluster. Error if this fails.
	// 2. Get list of all pods. Error if this fails.
	// 3. Run logs on every pod. Error if ALL fail only.

	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("Creating output directory in home", createOutputDirectory),
		cli.NewWorkflowStep("Running 'describe all' on cluster", describeAllStep),
		cli.NewWorkflowStep("Fetching logs from all Pure1 Unplugged pods in cluster", unpluggedLogStep),
		cli.NewWorkflowStep("Fetching logs from all system pods in cluster", systemLogStep),
		cli.NewWorkflowStep("Compressing logs", compressStep),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Finished!")

	return nil
}

func createOutputDirectory(ctx cli.Context) error {
	fullOutputDirectory = fmt.Sprintf("%s%s", outputDirectoryPrefix, time.Now().UTC().Format("2006-01-02-15:04:05"))
	return os.Mkdir(fullOutputDirectory, outputPermissions)
}

func describeAllStep(ctx cli.Context) error {
	// Run `kubectl describe all`
	output, err := kube.RunKubeCTLWithNamespace(ctx.Exec, kube.Pure1UnpluggedNamespace, "describe", "all")
	if err != nil {
		return err
	}

	// Output to file
	filePath := fmt.Sprintf("%s/describe-all", fullOutputDirectory)
	err = ioutil.WriteFile(filePath, []byte(output), outputPermissions)
	return err
}

func unpluggedLogStep(ctx cli.Context) error {
	// Find all the pods involved in the P1UP deployment that we'll pull logs from
	pods, err := kube.GetPods(ctx, kube.Pure1UnpluggedNamespace)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("No pods found")
	}

	hadSuccess := false // Only one needs to succeed for us to "succeed"

	for _, podName := range pods {
		// Get logs
		output, err := kube.RunKubeCTLWithNamespace(ctx.Exec, kube.Pure1UnpluggedNamespace, "logs", podName)
		if err != nil {
			// Log the error, since we aren't returning an error but it's good to know just in case anyways
			log.WithError(err).WithField("pod", podName).Error("Failed to get logs from pod")
			continue
		}

		// Save logs to file
		filePath := fmt.Sprintf("%s/logs-%s", fullOutputDirectory, podName)
		err = ioutil.WriteFile(filePath, []byte(output), outputPermissions)
		if err != nil {
			continue
		}

		hadSuccess = true
		log.WithField("pod", podName).Error("Successfully fetched logs from pod")
	}

	if !hadSuccess {
		return fmt.Errorf("Failed to get logs from any pod")
	}

	return nil
}

func systemLogStep(ctx cli.Context) error {
	// Find all the pods in kube-system that we'll pull logs from
	pods, err := kube.GetPods(ctx, kube.SystemNamespace)
	if err != nil {
		return err
	}
	if len(pods) == 0 {
		return fmt.Errorf("No system pods found")
	}

	hadSuccess := false // Only one needs to succeed for us to "succeed"

	for _, podName := range pods {
		// Get logs
		output, err := kube.RunKubeCTLWithNamespace(ctx.Exec, kube.SystemNamespace, "logs", podName)
		if err != nil {
			// Log the error, since we aren't returning an error but it's good to know just in case anyways
			log.WithError(err).WithField("pod", podName).Error("Failed to get system logs from pod")
			continue
		}

		// Save logs to file
		filePath := fmt.Sprintf("%s/logs-system-%s", fullOutputDirectory, podName)
		err = ioutil.WriteFile(filePath, []byte(output), outputPermissions)
		if err != nil {
			continue
		}

		hadSuccess = true
		log.WithField("pod", podName).Error("Successfully fetched system logs from pod")
	}

	if !hadSuccess {
		return fmt.Errorf("Failed to get system logs from any pod")
	}

	return nil
}

func compressStep(ctx cli.Context) error {
	// Note: using "RunCommandWithOutput" because it has nice clean built-in error handling, even though we don't need the output (there is no output from tar except for errors)
	_, err := ctx.Exec.RunCommandWithOutput("tar", "-zcf", fmt.Sprintf("./%s.tar.gz", fullOutputDirectory), fmt.Sprintf("./%s", fullOutputDirectory)) // Run `tar -zcf [outputdirectory].tar.gz [outputdirectory]`
	return err
}
