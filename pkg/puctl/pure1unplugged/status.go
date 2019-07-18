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
	"encoding/json"
	"fmt"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/kube"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli/status"
	log "github.com/sirupsen/logrus"
	"strings"
)

// Status is the entrypoint for the workflow which will determine the status of the Pure1 Unplugged application
func Status(ctx cli.Context) error {
	// Pass along a map of statuses with our context
	var currentStatus status.CLIContext
	ctx.CmdContext = &currentStatus

	wf := cli.NewWorkFlowEngine(
		cli.NewNonTerminatingWorkflowStep("Checking Helm deployment", checkPure1UnpluggedHelmDeployment),
		cli.NewNonTerminatingWorkflowStep("Checking Pure1 Unplugged pods", checkRequiredPure1UnpluggedPods),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	status.PrintTable(currentStatus)

	return nil
}

// Helm deployment status codes according to:
// https://github.com/helm/helm/blob/master/_proto/hapi/release/status.proto
const (
	// helmStatusUnknown indicates that a release is in an uncertain state.
	helmStatusUnknown = 0
	// helmStatusDeployed indicates that the release has been pushed to Kubernetes.
	helmStatusDeployed = 1
	// helmStatusDeleted indicates that a release has been deleted from Kubernetes.
	helmStatusDeleted = 2
	// helmStatusSuperseded indicates that this release object is outdated and a newer one exists.
	helmStatusSuperseded = 3
	// helmStatusFailed indicates that the release was not successfully deployed.
	helmStatusFailed = 4
	// helmStatusDeleting indicates that a delete operation is underway.
	helmStatusDeleting = 5
	// helmStatusPendingInstall indicates that an install operation is underway.
	helmStatusPendingInstall = 6
	// helmStatusPendingUpgrade indicates that an upgrade operation is underway.
	helmStatusPendingUpgrade = 7
	// helmStatusPendingRollback indicates that an rollback operation is underway.
	helmStatusPendingRollback = 8

	// helmStatusMissingObjects indicates that the helm status shows some item is "missing"
	// Note that this is *not* part of the official status codes!!
	helmStatusMissingObjects = -1
)

var helmStatusCodeStrings = map[int]string{
	helmStatusUnknown:         "Unknown",
	helmStatusDeployed:        "Deployed",
	helmStatusDeleted:         "Deleted",
	helmStatusSuperseded:      "Superseded",
	helmStatusFailed:          "Failed",
	helmStatusDeleting:        "Deleting",
	helmStatusPendingInstall:  "Pending Install",
	helmStatusPendingUpgrade:  "Pending Upgrade",
	helmStatusPendingRollback: "Pending Rollback",
	helmStatusMissingObjects:  "Missing Object(s)",
}

type helmStatusReponse struct {
	Name string         `json:"Name"`
	Info helmStatusInfo `json:"Info"`
}

type helmStatusInfo struct {
	Status      helmStatus `json:"status"`
	Description string     `json:"description"`
}

type helmStatus struct {
	Code      int    `json:"code"`
	Resources string `json:"resources"`
}

func checkPure1UnpluggedHelmDeployment(ctx cli.Context) error {
	curStatus := status.Info{
		Name: "helm deployment",
		Ok:   status.NotOK, // unless proven otherwise
	}

	statusCode := helmStatusUnknown
	statusDescription := "??"

	out, err := kube.RunHelm(ctx.Exec, 15, "status", "pure1-unplugged", "--output=json")
	if err != nil {
		log.WithFields(log.Fields{
			"output": out,
			"error":  err.Error(),
		}).Error("Failed to get status of 'pure1-unplugged' deployment")
	} else {
		log.Debugf("Helm status response: '%s'", out)
		var resp helmStatusReponse
		err = json.Unmarshal([]byte(out), &resp)
		if err != nil {
			log.Errorf("Failed to parse helm status response: %s", err.Error())
		} else {
			statusCode = resp.Info.Status.Code
			statusDescription = resp.Info.Description
			lines := strings.Split(resp.Info.Status.Resources, "\n")
			for i := 0; i < len(lines); i++ {
				line := lines[i]
				if strings.Contains(line, "MISSING") {
					// Uh oh.. helm sees we are missing something. Log it and set status accordingly
					// expect the output to be like:
					/*

						==> MISSING
						KIND                                      NAME
						apps/v1beta2, Resource=deployments        pure1-unplugged-monitor-server
						extensions/v1beta1, Resource=ingresses    pure1-unplugged-web-content

					*/
					// increment one more line (should always be safe) to get into the MISSING section
					for i = i + 1; i < len(lines); i++ {
						line := lines[i]
						if strings.HasPrefix(line, "KIND") {
							continue
						}
						log.Errorf("Helm found missing object: '%s'", line)
					}
					statusCode = helmStatusMissingObjects
				}
			}
		}
	}

	if statusCode == helmStatusDeployed {
		curStatus.Ok = status.OK
	}

	curStatus.Details = fmt.Sprintf("pure1-unplugged status: '%s', Last Update: %s", helmStatusCodeStrings[statusCode], statusDescription)

	ctx.CmdContext.(*status.CLIContext).Statuses = append(ctx.CmdContext.(*status.CLIContext).Statuses, curStatus)
	return nil
}

func checkRequiredPure1UnpluggedPods(ctx cli.Context) error {
	expectedPodsMap := requiredPodsMap()
	systemPodStatus, err := kube.GetPodStatuses(ctx, kube.Pure1UnpluggedNamespace, expectedPodsMap)
	gotPodStatus := true
	if err != nil {
		// if we didn't get back our pods, assume they are all bad, but show status for them
		gotPodStatus = false
		systemPodStatus = expectedPodsMap
	}
	log.Debugf("pure1-unplugged pod status: %s", systemPodStatus)

	// We're going to add a status for each pod group separately
	for podName, podInfo := range systemPodStatus {
		curStatus := status.Info{
			Name:    fmt.Sprintf("%s", podName),
			Details: fmt.Sprintf("Pods ready: %d/%d [%s]", podInfo.ReadyCount, podInfo.ExpectedCount, strings.Join(podInfo.Pods, ",")),
			Ok:      status.NotOK,
		}
		if podInfo.ReadyCount == podInfo.ExpectedCount {
			curStatus.Ok = status.OK
		}
		if !gotPodStatus {
			// we didn't really get the status, mark the details as such
			curStatus.Details = "unable to get pod details"
		}
		ctx.CmdContext.(*status.CLIContext).Statuses = append(ctx.CmdContext.(*status.CLIContext).Statuses, curStatus)
	}
	return err
}

// requiredPodsMap defines a map of identifiers for system pods we expect to be running
// for pure1-unplugged to work.
func requiredPodsMap() map[string]*kube.PodGroupInfo {
	return map[string]*kube.PodGroupInfo{
		"api-server":  kube.NewPodGroupInfo("pure1-unplugged-api-server-.*", 1),
		"auth-server": kube.NewPodGroupInfo("pure1-unplugged-auth-server-.*", 1),
		"dex":         kube.NewPodGroupInfo("pure1-unplugged-dex-.*", 1),
		"elasticsearch-client":          kube.NewPodGroupInfo("pure1-unplugged-elasticsearch-client-.*", 1),
		"elasticsearch-data":            kube.NewPodGroupInfo("pure1-unplugged-elasticsearch-data-.*", 1),
		"elasticsearch-master":          kube.NewPodGroupInfo("pure1-unplugged-elasticsearch-master-.*", 1),
		"kibana":                        kube.NewPodGroupInfo("pure1-unplugged-kibana-.*", 1),
		"metrics-client":                kube.NewPodGroupInfo("pure1-unplugged-metrics-client-.*", 1),
		"monitor-server":                kube.NewPodGroupInfo("pure1-unplugged-monitor-server-.*", 1),
		"nginx-ingress-controller":      kube.NewPodGroupInfo("pure1-unplugged-nginx-ingress-controller-.*", 1),
		"nginx-ingress-default-backend": kube.NewPodGroupInfo("pure1-unplugged-nginx-ingress-default-backend-.*", 1),
		"swagger-server":                kube.NewPodGroupInfo("pure1-unplugged-swagger-server-.*", 1),
		"web-content":                   kube.NewPodGroupInfo("pure1-unplugged-web-content-.*", 1),
	}
}
