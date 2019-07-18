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
	"fmt"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/config"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/docker"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/kube"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
	log "github.com/sirupsen/logrus"
	"net"
	"os"
	"path/filepath"
	"strings"
)

const pure1unpluggedHTTPSCertSecretName = "pure1-unplugged-https-cert"

// Install is the entrypoint for the workflow which will deploy the Pure1 Unplugged application
func Install(ctx cli.Context) error {
	wf := cli.NewWorkFlowEngine(
		cli.NewWorkflowStep("Checking for existing deployments", checkAlreadyInstalled),
		cli.NewWorkflowStep("Validating global.publicAddress", validateGlobalAddress),
		cli.NewWorkflowStep("Unpacking container images", loadDockerImages),
		cli.NewWorkflowStep("Checking public facing SSL certs", validateSSLCerts),
		cli.NewWorkflowStep("Ensuring 'pure1-unplugged' app namespace exists", createNamespace),
		cli.NewWorkflowStep("Configuring secrets", configureSecrets),
		cli.NewWorkflowStep("Installing Pure1 Unplugged via Helm", helmInstall),
	)

	err := wf.Run(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Finished!")
	fmt.Printf("\n\n\n\t Pure1 Unplugged Dashboard will be available at: https://%s/\n\n", ctx.Config.GetString(config.PublicAddressKey))

	return nil
}

func checkAlreadyInstalled(ctx cli.Context) error {
	out, err := kube.RunHelm(ctx.Exec, 30, "list", "--all", "--short")
	if err != nil {
		return fmt.Errorf("unable to list helm deployments, check system status: %s", err.Error())
	}

	for _, line := range strings.Split(out, "\n") {
		if strings.Contains(line, "pure1-unplugged") {
			return fmt.Errorf("found existing deployment, aborting")
		}
	}

	log.Debugf("Didn't see existing deployment in helm list output: %s", out)

	return nil
}

func validateGlobalAddress(ctx cli.Context) error {
	publicAddress := ctx.Config.GetString(config.PublicAddressKey)
	log.Infof("Found global.publicAddress == %s", publicAddress)

	// make sure we can reach the public address
	addrs, err := net.LookupHost(publicAddress)
	if err != nil {
		err = fmt.Errorf("failed to lookup global.publicAddress '%s' error: %s", publicAddress, err.Error())
		log.Error(err.Error())
	} else if len(addrs) == 0 {
		err = fmt.Errorf("failed to lookup global.publicAddress '%s'", publicAddress)
		log.Error(err.Error())
	} else {
		log.Debugf("successfully performed host lookup for global.publicAdress '%s', addrs='%s'", publicAddress, addrs)
	}
	return err
}

func loadDockerImages(ctx cli.Context) error {
	err := docker.LoadDockerImagesFromFiles(ctx, filepath.Join(ctx.Config.GetString(config.Pure1UnpluggedInstallDirKey), "/images/apps/pure1-unplugged/*"))
	if err != nil {
		return fmt.Errorf("failed to find docker image files: %s", err.Error())
	}
	return nil
}

func validateSSLCerts(ctx cli.Context) error {
	certFile := ctx.Config.GetString(config.SSLCertFileKey)
	keyFile := ctx.Config.GetString(config.SSLKeyFileKey)
	createKeys := ctx.Config.GetBool(config.CreateSelfSignedCertsKey)

	if !createKeys {
		log.Infof("Attempting to use specified certificates")
		certFileOK := false
		keyFileOK := false

		certFileOK, err := ctx.Filesystem.PathExists(certFile)
		if err != nil {
			log.Debugf("file %s stat error: %v", certFile, err)
			certFileOK = false
		}

		keyFileOK, err = ctx.Filesystem.PathExists(keyFile)
		if err != nil {
			log.Debugf("file %s stat error: %v", keyFile, err)
			keyFileOK = false
		}

		// If either are missing then raise an error
		if !keyFileOK || !certFileOK {
			return fmt.Errorf("failed to find SSL Cert and Key files (%s, %s)", certFile, keyFile)
		}
	} else {
		// Use self signed certs
		log.Infof("Generating self signed certs!")

		// ensure the base directories exist first
		err := os.MkdirAll(filepath.Dir(certFile), os.ModeDir)
		if err != nil {
			return fmt.Errorf("failed to create base directory for %s: %s", certFile, err.Error())
		}

		err = os.MkdirAll(filepath.Dir(keyFile), os.ModeDir)
		if err != nil {
			return fmt.Errorf("failed to create base directory for %s: %s", keyFile, err.Error())
		}

		// Get the publicAddress from our config, we will set the "Common Name" (CN) in the cert to match it
		publicAddress := ctx.Config.GetString(config.PublicAddressKey)

		// Generate the certificate
		_, err = ctx.Exec.RunCommandWithOutputRaw("openssl", "req", "-x509", "-nodes", "-days",
			"3650", "-newkey", "rsa:2048", "-keyout", keyFile, "-out", certFile, "-subj",
			fmt.Sprintf("/CN=%s/O=Pure1 Unplugged", publicAddress))
		if err != nil {
			return fmt.Errorf("failed to generate self signed SSL cert and key: %s", err.Error())
		}

		// lock down permissions on the files
		err = os.Chmod(certFile, 0600)
		if err != nil {
			return fmt.Errorf("failed to set 0600 file mode for %s: %s", certFile, err.Error())
		}

		err = os.Chmod(keyFile, 0600)
		if err != nil {
			return fmt.Errorf("failed to set 0600 file mode for %s: %s", keyFile, err.Error())
		}
	}

	return nil
}

func createNamespace(ctx cli.Context) error {
	namespaceExists := true
	_, err := kube.RunKubeCTL(ctx.Exec, "get", "namespace", kube.Pure1UnpluggedNamespace)
	if err != nil {
		namespaceExists = false
	}

	if namespaceExists {
		log.Debugf("namespace %s already exists, skipping to create", kube.Pure1UnpluggedNamespace)
	} else {
		_, err = kube.RunKubeCTL(ctx.Exec, "create", "namespace", kube.Pure1UnpluggedNamespace)
	}

	return err
}

func configureSecrets(ctx cli.Context) error {
	certFile := ctx.Config.GetString(config.SSLCertFileKey)
	keyFile := ctx.Config.GetString(config.SSLKeyFileKey)
	out, err := kube.RunKubeCTLWithNamespace(ctx.Exec, kube.Pure1UnpluggedNamespace, "create", "secret",
		"tls", pure1unpluggedHTTPSCertSecretName, "--key", keyFile, "--cert", certFile)
	if err != nil {
		return fmt.Errorf("failed to create https cert secret '%s': %s", pure1unpluggedHTTPSCertSecretName, out)
	}
	return nil
}

func findHelmChart(ctx cli.Context) (string, error) {
	// Find the helm chart, looking it up makes it easier for development than hard coding it in here
	chartGlob := filepath.Join(ctx.Config.GetString(config.Pure1UnpluggedInstallDirKey), "pure1-unplugged-*.tgz")
	log.Debugf("Looking for helm charts with glob %s", chartGlob)
	potentialChartFiles, err := filepath.Glob(chartGlob)
	if err != nil || len(potentialChartFiles) == 0 {
		return "", fmt.Errorf("failed to find helm chart in %s: %s", ctx.Config.GetString(config.Pure1UnpluggedInstallDirKey), err.Error())
	} else if len(potentialChartFiles) > 1 {
		return "", fmt.Errorf("found too many chart file options: %s", potentialChartFiles)
	}
	chartFile := potentialChartFiles[0]
	log.Debugf("Found helm chart file %s", chartFile)
	return chartFile, nil
}

func helmInstall(ctx cli.Context) error {
	chartFile, err := findHelmChart(ctx)
	if err != nil {
		return err
	}

	_, err = kube.RunHelm(ctx.Exec,
		900, // 15 min should be enough time.. longer probably means an error
		"install",
		chartFile,
		"--name=pure1-unplugged",
		"--namespace=pure1-unplugged",
		"--values=/etc/pure1-unplugged/config.yaml",
		"--wait",
	)

	if err != nil {
		return fmt.Errorf("failed to helm install: %s", err.Error())
	}
	return nil
}
