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
	"strings"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/kube"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
)

var podsToDelete []string

// DeletePod is the entry point to delete pods in the kubernetes cluster
func DeletePod(ctx cli.Context) error {
	if len(ctx.Args) == 0 {
		return fmt.Errorf("Usage: puctl infra delete-pod [patterns]")
	}

	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("", findPods),
		cli.NewWorkflowStep("Deleting pods", deletePods),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Finished!")

	return nil
}

func findPods(ctx cli.Context) error {
	podsToDelete = []string{}

	for _, rawPattern := range ctx.Args {
		pattern := strings.TrimSpace(rawPattern)

		if len(pattern) == 0 {
			return fmt.Errorf("pattern cannot be empty")
		}

		pods, err := kube.GetPodsFiltered(ctx, kube.Pure1UnpluggedNamespace, pattern)
		if err != nil {
			return fmt.Errorf("failed to get pods list: %v", err)
		}

		podsToDelete = append(podsToDelete, pods...)
	}

	if len(podsToDelete) == 0 {
		return fmt.Errorf(`found no pods matching the given patterns`)
	}

	fmt.Println("Pods to be deleted:")
	for _, pod := range podsToDelete {
		fmt.Println(pod)
	}
	fmt.Println()

	return nil
}

func deletePods(ctx cli.Context) error {
	args := []string{"delete", "pods"}
	_, err := kube.RunKubeCTLWithNamespace(ctx.Exec, kube.Pure1UnpluggedNamespace, append(args, podsToDelete...)...)

	return err
}
