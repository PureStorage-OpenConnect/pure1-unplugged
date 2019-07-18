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
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/kube"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli/status"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

// Status is the entrypoint for determining the status of the local Pure1 Unplugged infrastructure
func Status(ctx cli.Context) error {
	// Pass along a map of statuses with our context
	var currentStatus status.CLIContext
	ctx.CmdContext = &currentStatus

	wf := cli.NewWorkFlowEngine(
		cli.NewNonTerminatingWorkflowStep("Checking Firewall", getFirewalldStatus),
		cli.NewNonTerminatingWorkflowStep("Checking Selinux", getSelinuxStatus),
		cli.NewNonTerminatingWorkflowStep("Checking IP route", getRouteStatus),
		cli.NewNonTerminatingWorkflowStep("Checking Kubelet", getKubeletStatus),
		cli.NewNonTerminatingWorkflowStep("Checking Docker", getDockerStatus),
		cli.NewNonTerminatingWorkflowStep("Checking Kubernetes node", getNodeStatus),
		cli.NewNonTerminatingWorkflowStep("Checking Kubernetes system pods", getSystemPodStatus),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	status.PrintTable(currentStatus)

	return nil
}

func getFirewalldStatus(ctx cli.Context) error {
	curStatus := status.Info{
		Name: "firewalld",
		Ok:   status.NotOK, // unless proven otherwise
	}

	expectedServices := []string{"https"}

	out, err := ctx.Exec.RunCommandWithOutputRaw("firewall-cmd", "--list-all")

	if err == nil {
		for _, line := range strings.Split(string(out), "\n") {
			cleanLine := strings.TrimSpace(line)
			if strings.HasPrefix(cleanLine, "services:") {
				servicesLine := strings.TrimSpace(strings.TrimPrefix(cleanLine, "services:"))
				foundServices := 0
				for _, svc := range expectedServices {
					if strings.Contains(servicesLine, svc) {
						foundServices++
					}
				}
				if foundServices == len(expectedServices) {
					curStatus.Ok = status.OK
				}
			}
		}
	}

	curStatus.Details = "firewalld must be running and allow `https`"

	ctx.CmdContext.(*status.CLIContext).Statuses = append(ctx.CmdContext.(*status.CLIContext).Statuses, curStatus)
	return nil
}

func getSelinuxStatus(ctx cli.Context) error {
	curStatus := status.Info{
		Name: "selinux",
	}

	out, err := ctx.Exec.RunCommandWithOutput("getenforce")
	selinuxState := strings.TrimSpace(out)

	if err == nil && selinuxState == "Enforcing" {
		curStatus.Ok = status.OK
	} else {
		curStatus.Ok = status.NotOK
	}

	if err == nil {
		curStatus.Details = fmt.Sprintf("selinux state is '%s'", selinuxState)
	} else {
		curStatus.Details = "error running 'getenforce', see logs"
	}

	ctx.CmdContext.(*status.CLIContext).Statuses = append(ctx.CmdContext.(*status.CLIContext).Statuses, curStatus)
	return err
}

func getRouteStatus(ctx cli.Context) error {
	curStatus := status.Info{
		Name: "default ip route",
	}

	err := checkDefaultRoute(ctx)
	if err == nil {
		curStatus.Ok = status.OK
		curStatus.Details = "default ip route is set"
	} else {
		curStatus.Ok = status.NotOK
		curStatus.Details = "unable to determine default ip route"
	}

	ctx.CmdContext.(*status.CLIContext).Statuses = append(ctx.CmdContext.(*status.CLIContext).Statuses, curStatus)
	return nil
}

// getSystemdServiceStatus gets the "Active" field status for a specified systemd service
func getSystemdServiceStatus(ctx cli.Context, serviceName string) (string, error) {
	out, err := ctx.Exec.RunCommandWithOutputRaw("systemctl", "status", serviceName, "--plain", "--no-legend", "-n", "0")
	if err != nil {
		return "", err
	}

	for _, line := range strings.Split(string(out), "\n") {
		cleanLine := strings.TrimSpace(line)
		if strings.HasPrefix(cleanLine, "Active: ") {
			activeMessage := strings.TrimPrefix(cleanLine, "Active: ")
			err = nil
			if !strings.Contains(activeMessage, "active (running)") {
				err = errors.New("service state is not 'active (running)'")
			}
			return activeMessage, err
		}
	}

	return "", errors.New("service state is unknown")
}

func getKubeletStatus(ctx cli.Context) error {
	curStatus := status.Info{
		Name: "kubelet systemd service",
	}

	state, err := getSystemdServiceStatus(ctx, "kubelet")
	if err != nil {
		curStatus.Ok = status.NotOK
		curStatus.Details = err.Error()
	} else {
		curStatus.Ok = status.OK
		curStatus.Details = state
	}

	ctx.CmdContext.(*status.CLIContext).Statuses = append(ctx.CmdContext.(*status.CLIContext).Statuses, curStatus)
	return nil
}

func getDockerStatus(ctx cli.Context) error {
	curStatus := status.Info{
		Name: "docker systemd service",
	}

	state, err := getSystemdServiceStatus(ctx, "docker")
	if err != nil {
		curStatus.Ok = status.NotOK
		curStatus.Details = err.Error()
	} else {
		curStatus.Ok = status.OK
		curStatus.Details = state
	}

	ctx.CmdContext.(*status.CLIContext).Statuses = append(ctx.CmdContext.(*status.CLIContext).Statuses, curStatus)
	return nil
}

// minimal set of structs to parse node status

type nodeDetailsList struct {
	Items []nodeDetails `json:"items,omitempty"`
}

type nodeDetails struct {
	Status   nodeStatus   `json:"status,omitempty"`
	Metadata nodeMetadata `json:"metadata"`
}

type nodeMetadata struct {
	Name string `json:"name"`
}

type nodeStatus struct {
	Conditions []nodeConditions `json:"conditions,omitempty"`
}

type nodeConditions struct {
	Type    string `json:"type,omitempty"`
	Status  string `json:"status,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Message string `json:"message,omitempty"`
}

func getNodeStatus(ctx cli.Context) error {
	curStatus := status.Info{
		Name:    "kubelet node status",
		Details: "kubelet node is not ready",
		Ok:      status.NotOK,
	}

	out, err := kube.RunKubeCTL(ctx.Exec, "get", "nodes", "--output=json")
	if err == nil {
		log.Debugf("raw node json: " + out)
		var detailsList nodeDetailsList
		err = json.Unmarshal([]byte(out), &detailsList)
		if err == nil {
			// Look for our current node
			hostname, err := os.Hostname()
			if err == nil {
				for _, node := range detailsList.Items {
					log.Debugf("checking node %+v", node)
					if strings.Contains(node.Metadata.Name, hostname) {
						// found it, check its status conditions as defined by
						// https://kubernetes.io/docs/concepts/architecture/nodes/#condition
						// we mostly just care about "Ready"
						for _, condition := range node.Status.Conditions {
							log.Debugf("checking condition %+v", condition)
							if condition.Type == "Ready" {
								curStatus.Details = fmt.Sprintf("kubelet node 'Ready=%s'", condition.Status)
								if condition.Status == "True" {
									curStatus.Ok = status.OK
								}
								break
							}
						}
					}
				}
			}
		}
	}

	ctx.CmdContext.(*status.CLIContext).Statuses = append(ctx.CmdContext.(*status.CLIContext).Statuses, curStatus)
	return err
}

func getSystemPodStatus(ctx cli.Context) error {
	expectedPodsMap := requiredSystemPodsMap()
	systemPodStatus, err := kube.GetPodStatuses(ctx, kube.SystemNamespace, expectedPodsMap)
	gotPodStatus := true
	if err != nil {
		// if we didn't get back our pods, assume they are all bad, but show status for them
		gotPodStatus = false
		systemPodStatus = expectedPodsMap
	}
	log.Debugf("System pod status: %s", systemPodStatus)

	// We're going to add a status for each system pod separately
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

// requiredSystemPodsMap defines a map of identifiers for system pods we expect to be running
// for kubernetes to work.
func requiredSystemPodsMap() map[string]*kube.PodGroupInfo {
	return map[string]*kube.PodGroupInfo{
		"etcd":                    kube.NewPodGroupInfo("etcd-.*", 1),
		"kube-apiserver":          kube.NewPodGroupInfo("kube-apiserver-.*", 1),
		"kube-controller-manager": kube.NewPodGroupInfo("kube-controller-manager-.*", 1),
		"kube-proxy":              kube.NewPodGroupInfo("kube-proxy-.*", 1),
		"kube-scheduler":          kube.NewPodGroupInfo("kube-scheduler-.*", 1),
		"coredns":                 kube.NewPodGroupInfo("coredns-.*", 2),
		"calico-node":             kube.NewPodGroupInfo("calico-node-.*", 1),
		"calicoctl":               kube.NewPodGroupInfo("calicoctl", 1),
		"tiller":                  kube.NewPodGroupInfo("tiller-.*", 1),
	}
}
