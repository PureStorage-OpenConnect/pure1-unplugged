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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/config"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/docker"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/exechelper"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/kube"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Initialize is the entry point for the workflow to initializing pure1-unplugged infra on the current host. By then end of
// this call there should be a kubernetes cluster with networking and helm all ready to go.
func Initialize(ctx cli.Context) error {
	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("Checking for existing Kubernetes cluster", checkClusterStatus),
		cli.NewWorkflowStep("Validating network config options (config.yaml)", validateNetworkConfig),
		cli.NewWorkflowStep("Checking default route config", checkDefaultRoute),
		cli.NewWorkflowStep("Ensuring system state is ready for Kubernetes initialization", runPrereqPlaybook),
		cli.NewWorkflowStep("Unpacking container images", loadDockerImages),
		cli.NewWorkflowStep("Generating CNI plugin configuration", generateCalicoYamls),
		cli.NewWorkflowStep("Generating kubeadm configuration", generateKubeadmConfig),
		cli.NewWorkflowStep("Configuring apiserver encryption", configureAPIServerEncryption),
		cli.NewWorkflowStep("Initializing Kubernetes controller", kubeadmInit),
		cli.NewWorkflowStep("Waiting for initial system pods to be running", waitForInitialSystemPods),
		cli.NewWorkflowStep("Untainting master node", setupNodeTaints),
		cli.NewWorkflowStep("Installing CNI plugin", installCalico),
		cli.NewWorkflowStep("Waiting for remaining critical system pods to initialize", waitForCriticalSystemPods),
		cli.NewWorkflowStep("Initializing Helm", helmInit),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Finished!")

	return nil
}

func checkClusterStatus(ctx cli.Context) error {
	_, err := kube.RunKubeCTL(ctx.Exec, "get", "cs")
	if err == nil {
		log.Debug("`kubectl get cs` shows the cluster is running")
		return errors.New("unable to initialize cluster, looks like it is already running.\n Try 'puctl infra reset' if you want to clear the system and re-install")
	}
	return nil
}

func validateNetworkConfig(ctx cli.Context) error {
	podCIDR := ctx.Config.GetString(config.PodCIDRKey)
	serviceCIDR := ctx.Config.GetString(config.ServiceCIDRKey)
	_, podNet, err := net.ParseCIDR(podCIDR)
	if err != nil {
		return fmt.Errorf("failed to parse podCIDR '%s', ensure it is a CIDR as defined by RFC 4632 and RFC 4291. (Ex: 10.1.2.0/24): %s", podCIDR, err.Error())
	}

	_, serviceNet, err := net.ParseCIDR(serviceCIDR)
	if err != nil {
		return fmt.Errorf("failed to parse serviceCIDR '%s', ensure it is a CIDR as defined by RFC 4632 and RFC 4291. (Ex: 10.1.2.0/24): %s", serviceCIDR, err.Error())
	}

	log.Infof("Using serviceCIDR (%s) and podCIDR (%s)\n", serviceCIDR, podCIDR)

	// ensure that our CIDR's don't overlap
	if podNet.Contains(serviceNet.IP) || serviceNet.Contains(podNet.IP) {
		return fmt.Errorf("serviceCIDR (%s) and podCIDR (%s) overlap! Ensure that they are separate ranges", serviceCIDR, podCIDR)
	}

	// ensure they are at least 16's
	podNetMaskLen, _ := podNet.Mask.Size()
	serviceNetMaskLen, _ := serviceNet.Mask.Size()
	if podNetMaskLen == 0 || serviceNetMaskLen == 0 {
		log.Errorf("WARNING: serviceCIDR (%s) and podCIDR (%s) are not in a format that can be validated. Networking may have issues!!!", serviceCIDR, podCIDR)
	} else if podNetMaskLen > 16 || serviceNetMaskLen > 16 {
		return fmt.Errorf("serviceCIDR (%s) and podCIDR (%s) need to be /16 or larger ranges", serviceCIDR, podCIDR)
	}

	return nil
}

func checkDefaultRoute(ctx cli.Context) error {
	// Not the smartest test in the world.. but should work in nearly every case to find the default route
	out, err := ctx.Exec.RunCommandWithOutputRaw("ip", "route")
	if err != nil {
		return fmt.Errorf("unable to determine default route: %s", err.Error())
	}
	foundDefault := false
	for _, line := range strings.Split(string(out), "\n") {
		if strings.HasPrefix(line, "default via") {
			log.Debugf("Found default route rule: '%s'", line)
			foundDefault = true
			break
		}
	}
	if !foundDefault {
		return errors.New("Unable to find default route, ensure it is present (see 'ip route' output)")
	}

	return nil
}

func runPrereqPlaybook(ctx cli.Context) error {
	params := exechelper.ExecParams{
		CmdName: "ansible-playbook",
		CmdArgs: []string{
			"--connection=local",
			"-i", "127.0.0.1,",
			filepath.Join(ctx.Config.GetString(config.Pure1UnpluggedInstallDirKey), "/infra/playbooks/infra-prereq.yaml"),
		},
		Timeout: 300, // 5 min timeout
	}
	result := ctx.Exec.RunCommand(params)
	log.Debug("stdout: " + result.OutBuf.String())
	log.Debug("stderr: " + result.ErrBuf.String())
	if result.Error != nil {
		return fmt.Errorf("failed to run playbook: %s", result.ErrBuf.String())
	}
	return nil
}

func loadDockerImages(ctx cli.Context) error {
	err := docker.LoadDockerImagesFromFiles(ctx, filepath.Join(ctx.Config.GetString(config.Pure1UnpluggedInstallDirKey), "/images/infra/*"))
	if err != nil {
		return fmt.Errorf("failed to find docker image files: %s", err.Error())
	}
	return nil
}

func generateFilesFromTemplates(ctx cli.Context, templateGlob string, conf map[string]string) error {
	templatesFiles, err := ctx.Filesystem.Glob(templateGlob)
	if err != nil || len(templatesFiles) == 0 {
		return fmt.Errorf("unable to find templates manifests in %s, error: %s", templateGlob, err.Error())
	}

	// Read in each template and swap in the pod CIDR
	for _, templateFile := range templatesFiles {
		tmpl, err := template.New(filepath.Base(templateFile)).ParseFiles(templateFile)
		if err != nil {
			return fmt.Errorf("failed to load template file %s; %s", templateFile, err.Error())
		}
		newFileName := strings.TrimSuffix(templateFile, ".template")
		outFile, err := os.Create(newFileName)
		if err != nil {
			return fmt.Errorf("failed to open file %s: %s", templateFile, err.Error())
		}
		err = tmpl.Execute(outFile, conf)
		closeErr := outFile.Close()
		if err != nil {
			return fmt.Errorf("failed to render template %s: %s", templateFile, err.Error())
		}
		if closeErr != nil {
			return fmt.Errorf("failed to open file %s: %s", templateFile, err.Error())
		}
	}
	return nil
}

func generateCalicoYamls(ctx cli.Context) error {
	calicoTemplateGlob := filepath.Join(ctx.Config.GetString(config.Pure1UnpluggedInstallDirKey), "/infra/calico/*.yaml.template")
	err := generateFilesFromTemplates(ctx, calicoTemplateGlob, map[string]string{"CalicoPodCIDR": ctx.Config.GetString(config.PodCIDRKey)})
	if err != nil {
		return fmt.Errorf("failed to generate Calico yaml files: %s", err.Error())
	}
	return nil
}

func generateKubeadmConfig(ctx cli.Context) error {
	// We need to provide a kubeadm bootstrap token in the config file
	out, err := ctx.Exec.RunCommandWithOutput("kubeadm", "token", "generate")
	if err != nil {
		return fmt.Errorf("failed to generate kubeadm bootstrap token: %s", err.Error())
	}
	bootstrapToken := strings.TrimSpace(out)

	// Lookup the hostname and use it as our "NodeName"
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get current hostname: %s", err.Error())
	}

	// The encryption feature for the api server needs a 32 byte key that is base64 encoded for
	// the aescbc encryption provider.
	data := make([]byte, 32)
	_, err = rand.Read(data)
	if err != nil {
		return fmt.Errorf("failed to generate random key data: %s", err.Error())
	}
	encryptionKey := base64.StdEncoding.EncodeToString(data)

	conf := map[string]string{
		"Token":             bootstrapToken,
		"NodeName":          hostname,
		"KubernetesVersion": kube.KubeVersion,
		"PodCIDR":           ctx.Config.GetString(config.PodCIDRKey),
		"ServiceCIDR":       ctx.Config.GetString(config.ServiceCIDRKey),
		"EncryptionKey":     encryptionKey,
	}

	kubeadmTemplateGlob := filepath.Join(ctx.Config.GetString(config.Pure1UnpluggedInstallDirKey), "/infra/kubeadm/*.yaml.template")
	err = generateFilesFromTemplates(ctx, kubeadmTemplateGlob, conf)
	if err != nil {
		return fmt.Errorf("failed to generate kubeadm config files: %s", err.Error())
	}

	return nil
}

func configureAPIServerEncryption(ctx cli.Context) error {
	// Expect the config to have already been generated
	targetPath := "/etc/kubernetes/secrets"
	err := os.MkdirAll(targetPath, os.ModeDir)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %s", targetPath, err.Error())
	}

	srcFile := filepath.Join(ctx.Config.GetString(config.Pure1UnpluggedInstallDirKey), "/infra/kubeadm/kube-api-encryption.yaml")
	in, err := ioutil.ReadFile(srcFile)
	if err != nil {
		return fmt.Errorf("failed to read encryption config file %s: %s", srcFile, err.Error())
	}

	destFile := filepath.Join(targetPath, filepath.Base(srcFile))
	err = ioutil.WriteFile(destFile, in, 0600)
	if err != nil {
		return fmt.Errorf("failed to write encryption config file %s: %s", srcFile, err.Error())
	}

	return nil
}

func kubeadmInit(ctx cli.Context) error {
	params := exechelper.ExecParams{
		CmdName: "kubeadm",
		CmdArgs: []string{
			"init",
			"--config=" + filepath.Join(ctx.Config.GetString(config.Pure1UnpluggedInstallDirKey), "/infra/kubeadm/kubeadm-init-config.yaml"),
		},
		Timeout: 300, // 5 min timeout
	}
	result := ctx.Exec.RunCommand(params)
	log.Debug("stdout: " + result.OutBuf.String())
	log.Debug("stderr: " + result.ErrBuf.String())
	if result.Error != nil {
		return fmt.Errorf("failed to init kubernetes, see log for details: err: %s", result.ErrBuf.String())
	}

	return nil
}

func installCalico(ctx cli.Context) error {
	installRoot := ctx.Config.GetString(config.Pure1UnpluggedInstallDirKey)
	// order of these is somewhat important, don't just glob for *.yaml
	calicoFiles := []string{
		"calico.yaml",
		"calicoctl.yaml",
	}
	for _, fileName := range calicoFiles {
		filePath := filepath.Join(installRoot, "/infra/calico/", fileName)
		// Allow some retries for the apply, sometimes not everything is online yet
		out, err := kube.RunKubeCTLWithNamespaceRetryOnError(ctx.Exec, kube.SystemNamespace, 30, "apply", "-f", filePath)
		if err != nil {
			log.Error(out)
			return fmt.Errorf("failed to apply %s: %s", filePath, err.Error())
		}
	}

	return nil
}

func waitForSystemPods(ctx cli.Context, timeoutSeconds int, filter []string) error {
	delay := 5
	retries := timeoutSeconds / delay
	for i := 0; i < retries; i++ {
		podInfoMap, err := kube.GetPodStatuses(ctx, kube.SystemNamespace, requiredSystemPodsMap())
		if err == nil {
			doneWaiting := true
			for name, podInfo := range podInfoMap {
				if filter != nil && len(filter) > 0 {
					shouldConsider := false
					for _, filterName := range filter {
						if name == filterName {
							shouldConsider = true
							break
						}
					}
					// If it wasn't in our filter, skip it
					if !shouldConsider {
						continue
					}
				}
				if podInfo.ReadyCount != podInfo.ExpectedCount {
					// not ready yet, break back to the outer retry loop
					doneWaiting = false
					break
				}
			}
			if doneWaiting {
				// Finished! All pods are accounted for
				return nil
			}
		} else {
			log.Warnf("failed to get kube-system pod status: %s", err.Error())
		}
		// sleep and retry on all but the last iteration
		if i+1 != retries {
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}
	// If we made it this far we didn't get into the right system state
	return fmt.Errorf("system pods are not ready, check `puctl infra status` and logs for details")
}

func waitForInitialSystemPods(ctx cli.Context) error {
	// initial system pods we are looking for include:
	requiredPods := []string{"etcd", "kube-apiserver", "kube-controller-manager", "kube-proxy", "kube-scheduler"}
	// The coredns, apiserver, controllers, and calico pods are not expected to be running until *after* we install the CNI plugin
	// only wait up to 5 min for this first batch of initial pods, it should be pretty quick.
	return waitForSystemPods(ctx, 300, requiredPods)
}

func waitForCriticalSystemPods(ctx cli.Context) error {
	// Check status for almost all our system pods and ensure they are all running
	requiredPods := []string{"etcd", "kube-apiserver", "kube-controller-manager", "kube-proxy", "kube-scheduler", "calico-node", "calicoctl", "core-dns"}
	// Ignore tiller and calicoctl for now, its the last one that will work
	// Give this one 10 min, some of the network+dns+proxy pod interactions take time to reconcile
	return waitForSystemPods(ctx, 600, requiredPods)
}

func setupNodeTaints(ctx cli.Context) error {
	out, err := kube.RunKubeCTL(ctx.Exec, "taint", "nodes", "--all", "node-role.kubernetes.io/master-")
	log.Debug(string(out))
	if err != nil {
		return fmt.Errorf("failed changing node taints: %s", err.Error())
	}
	return nil
}

func waitForTillerRBAC(ctx cli.Context) error {
	// Let them wait for 2 min, if it takes that long something is very wrong
	_, saErr := kube.RunKubeCTLWithNamespaceRetryOnError(ctx.Exec, kube.SystemNamespace, 120, "get", "serviceaccount", "tiller")
	_, crbErr := kube.RunKubeCTLWithNamespaceRetryOnError(ctx.Exec, kube.SystemNamespace, 120, "get", "clusterrolebinding", "tiller")
	if saErr == nil && crbErr == nil {
		log.Debugf("Found tiller cluster role binding and service account")
		return nil
	}
	// If we made it this far we didn't get into the right system state
	return fmt.Errorf("tiller service account and role binding are not ready, check `puctl infra kube -- get all --all-namespaces` and logs for details")
}

func helmInit(ctx cli.Context) error {
	helmRBACYaml := filepath.Join(ctx.Config.GetString(config.Pure1UnpluggedInstallDirKey), "/infra/helm/tiller-rbac.yaml")
	// Allow some retries for the apply, sometimes not everything is online yet
	out, err := kube.RunKubeCTLWithNamespaceRetryOnError(ctx.Exec, kube.SystemNamespace, 30, "apply", "-f", helmRBACYaml)
	log.Debug(string(out))
	if err != nil {
		return fmt.Errorf("failed to setup tiller rbac: %s", err.Error())
	}

	err = waitForTillerRBAC(ctx)
	if err != nil {
		return err
	}

	out, err = kube.RunHelm(ctx.Exec, 30,
		"init",
		"--service-account=tiller",
		"--skip-refresh",
		"--upgrade",
		"--force-upgrade",
	)
	log.Debug(out)
	if err != nil {
		return fmt.Errorf("failed to helm init: %s", err.Error())
	}

	// Wait for tiller to start, it should only need a minute or two
	err = waitForSystemPods(ctx, 300, []string{"tiller"})
	if err != nil {
		return err
	}
	return nil
}
