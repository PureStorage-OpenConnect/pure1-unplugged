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
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/exechelper"
	log "github.com/sirupsen/logrus"
)

const kubeconf = "/etc/kubernetes/admin.conf"

// RunKubeCTL will run 'kubectl' and specify our known location for kubeconfig
func RunKubeCTL(exec exechelper.Executor, kubectlArgs ...string) (string, error) {
	return RunKubeCTLWithTimeout(exec, 30, kubectlArgs...)
}

// RunKubeCTLWithTimeout will run 'kubectl' and specify our known location for kubeconfig
func RunKubeCTLWithTimeout(exec exechelper.Executor, timeout int, kubectlArgs ...string) (string, error) {
	params := exechelper.ExecParams{
		CmdName: "kubectl",
		// Set the default KubeConfig for all commands we may try to run
		CmdArgs: []string{
			"--kubeconfig",
			kubeconf,
		},
		Timeout: timeout,
	}
	// Add in the user-specified parameters to the defaults
	params.CmdArgs = append(params.CmdArgs, kubectlArgs...)
	result := exec.RunCommand(params)

	combinedOutput := result.OutBuf.String() + result.ErrBuf.String()

	return combinedOutput, result.Error
}

// RunKubeCTLWithNamespace is a helper to inject namespace options into kubectl calls
// without having to always say "--namespace" blah blah
func RunKubeCTLWithNamespace(exec exechelper.Executor, namespace string, kubectlArgs ...string) (string, error) {
	args := append([]string{"--namespace", namespace}, kubectlArgs...)
	return RunKubeCTL(exec, args...)
}

// RunKubeCTLWithNamespaceRetryOnError is a wrapper around `RunKubeCTLWithNamespace` that will blindly retry on any error
// mostly useful for poking the API server after restarts to swallow errors as things come back online.
func RunKubeCTLWithNamespaceRetryOnError(exec exechelper.Executor, namespace string, timeout int, kubectlArgs ...string) (string, error) {
	delay := 10
	retries := timeout / delay
	var out string
	var err error
	for currentTry := 1; currentTry <= retries; currentTry++ {
		out, err = RunKubeCTLWithNamespace(exec, namespace, kubectlArgs...)
		log.Debugf("kubectl output: '%s', error: %s", out, err)
		if err == nil {
			break
		}
		log.Debugf("kubectl command failed, will retry (attempt %d/%d)...", currentTry, retries)
		// if an error occurs retry, callers need to ensure the thing we are calling can and should be retried
		time.Sleep(time.Duration(delay) * time.Second)
	}
	return out, err
}
