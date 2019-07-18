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
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/version"
	log "github.com/sirupsen/logrus"
	"strings"
)

// minimal structs to parse helm list output
// ex: {
// 			"Next":"",
// 			"Releases": [
// 				{
// 					"Name":"pure1-unplugged",
// 					"Revision":1,
// 					"Updated":"Thu Mar 14 14:35:18 2019",
// 					"Status":"DEPLOYED",
// 					"Chart":"pure1-unplugged-1.2.3",
// 					"AppVersion":"1.2.3",
// 					"Namespace":"pure1-unplugged"
// 				}
// 			]
// 		}
type helmListResponse struct {
	Releases []helmReleaseInfo `json:"Releases,omitempty"`
}

type helmReleaseInfo struct {
	Name       string `json:"Name"`
	AppVersion string `json:"AppVersion"`
}

// Version will check the current version of Pure1 Unplugged
func Version(ctx cli.Context) error {
	out, err := kube.RunHelm(ctx.Exec, 15, "list", "--output=json")
	log.Debug("output: " + out)
	deployedVersion := "??"
	if err != nil {
		log.Errorf("failed to helm release list: %s", err.Error())
	} else if strings.TrimSpace(out) != "" {
		var listResp helmListResponse
		err = json.Unmarshal([]byte(out), &listResp)
		if err != nil {
			log.Errorf("failed to parse helm list response: %s", err.Error())
		} else if listResp.Releases != nil {
			for _, release := range listResp.Releases {
				if release.Name == "pure1-unplugged" {
					deployedVersion = release.AppVersion
					break
				}
			}
		}
	}

	fmt.Printf("Pure1 Unplugged Available Version: %s\n", version.Get())
	fmt.Printf("Pure1 Unplugged Running Version: '%s'\n", deployedVersion)
	return nil
}
