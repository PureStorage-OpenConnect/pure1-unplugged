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

package kube

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
	log "github.com/sirupsen/logrus"
)

// PodGroupInfo defines a pod group that matches some pattern for the name
// and details about the group. Useful for checking on pods in a replica-set
// or deployment where you don't know the exact name (and don't care), but
// want to make sure enough are online.
type PodGroupInfo struct {
	Pattern       string
	ReadyCount    int
	ExpectedCount int
	Pods          []string
}

// NewPodGroupInfo is just a helper to create the PodGroupInfo objects
func NewPodGroupInfo(pattern string, expectedCount int) *PodGroupInfo {
	return &PodGroupInfo{
		Pattern:       pattern,
		ExpectedCount: expectedCount,
		ReadyCount:    0,
		Pods:          []string{},
	}
}

// PodListResponse provides minimal json parsing helpers for kubectl output of pod definitions
type PodListResponse struct {
	APIVersion string `json:"apiVersion"`
	Items      []Pod  `json:"items,omitempty"`
}

// Pod provides minimal json parsing helpers for kubectl output of pod definitions
type Pod struct {
	Metadata PodMetadata `json:"metadata,omitempty"`
	Status   PodStatus   `json:"status,omitempty"`
}

// PodMetadata provides minimal json parsing helpers for kubectl output of pod definitions
type PodMetadata struct {
	Name string `json:"name,omitempty"`
}

// PodStatus provides minimal json parsing helpers for kubectl output of pod definitions
type PodStatus struct {
	Phase             string            `json:"phase,omitempty"`
	ContainerStatuses []ContainerStatus `json:"containerStatuses,omitempty"`
}

// ContainerStatus provides minimal json parsing helpers for kubectl output of pod definitions
type ContainerStatus struct {
	Ready bool `json:"ready,omitempty"`
}

// GetPodStatuses will list all pods in the given namespace and then fill in the PodGroupInfo map passed in, and return
// a reference to the completed version. We only consider pods in the group ready if we can find one matching the pattern,
// and it is in the correct ready status.
func GetPodStatuses(ctx cli.Context, namespace string, expectedPodGroups map[string]*PodGroupInfo) (map[string]*PodGroupInfo, error) {
	log.Debug("Getting pod states")
	podsRawJSON, err := RunKubeCTLWithNamespace(ctx.Exec, namespace, "get", "pods", "--output=json")
	if err != nil {
		return nil, fmt.Errorf("failed to get pod list: %s", err.Error())
	}
	log.Debugf("Raw pod json: %s", podsRawJSON)
	var parsedJSON PodListResponse
	err = json.Unmarshal([]byte(podsRawJSON), &parsedJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pod list json: %s", err.Error())
	}

	if parsedJSON.APIVersion != "v1" {
		return nil, fmt.Errorf("unable to use pod list with api version %s", parsedJSON.APIVersion)
	}

	// For each pod get its name and current status,
	for _, pod := range parsedJSON.Items {
		name := pod.Metadata.Name
		status := pod.Status.Phase
		log.Debugf("Evaluating pod: %+v", pod)
		log.Debugf("Pod Name: %s, Status: %s", name, status)
		// find the containers and check each of their states to see if they are running
		for _, sysPod := range expectedPodGroups {
			matched, err := regexp.MatchString(sysPod.Pattern, name)
			log.Debugf("Pod Name '%s', maching against pattern '%s', matched: %d, err: %s", name, sysPod.Pattern, matched, err)
			if err == nil && matched {
				sysPod.Pods = append(sysPod.Pods, name)
				if strings.ToLower(status) == "running" {
					// If it has ready probes make sure they are all OK too
					ready := true
					if pod.Status.ContainerStatuses != nil {
						for _, cStatus := range pod.Status.ContainerStatuses {
							// if any of the sub-containers isn't ready then we need to wait
							ready = cStatus.Ready
						}
					}
					if ready {
						sysPod.ReadyCount++
					}
				}
				log.Debugf("System pod status: %+v", *sysPod)
			}
		}
	}

	for _, pod := range expectedPodGroups {
		log.Debugf("Final status after parsing checks: %+v", *pod)
	}

	return expectedPodGroups, nil
}

// GetPods gets a list of all pods in the given namespace
func GetPods(ctx cli.Context, namespace string) ([]string, error) {
	log.Debug("Getting pod states")
	podsList, err := RunKubeCTLWithNamespace(ctx.Exec, namespace, "get", "pods", "-o", `go-template='{{range .items}}{{.metadata.name}}{{"\n"}}{{- end}}'`)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod list: %s", err.Error())
	}
	log.Debugf("Raw pod list: %s", podsList)
	return strings.Split(podsList, "\n"), nil
}

// GetPodsFiltered gets a list of all pods in the given namespace, filtered by if the name contains the given filter string
func GetPodsFiltered(ctx cli.Context, namespace string, filter string) ([]string, error) {
	pods, err := GetPods(ctx, namespace)
	if err != nil {
		return nil, err
	}
	matching := []string{}
	for _, name := range pods {
		if strings.Contains(name, filter) {
			matching = append(matching, name)
		}
	}

	return matching, nil
}
