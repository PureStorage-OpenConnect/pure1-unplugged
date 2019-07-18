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
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/kube"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/version"
)

// Version will check the current version of the app infrastructure (kubernetes)
func Version(ctx cli.Context) error {
	out, err := kube.RunKubeCTL(ctx.Exec, "version")
	fmt.Printf("Pure1 Unplugged Version: %s\n", version.Get())
	fmt.Printf("Kubernetes available version: '%s'\n", kube.KubeVersion)
	fmt.Printf("Currently running Kubernetes Details:\n%s\n'", out)
	return err
}
