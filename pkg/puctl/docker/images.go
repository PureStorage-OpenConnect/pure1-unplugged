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

package docker

import (
	"errors"
	"fmt"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/puctl/exechelper"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/util/cli"
	log "github.com/sirupsen/logrus"
)

// LoadDockerImagesFromFiles will use the given filepath glob to find docker images and load them into the engine
func LoadDockerImagesFromFiles(ctx cli.Context, imageGlob string) error {
	dockerImages, err := ctx.Filesystem.Glob(imageGlob)
	if err != nil {
		return fmt.Errorf("failed to find docker image files in '%s': %s", imageGlob, err.Error())
	}

	errorChan := make(chan error)
	for _, imageFile := range dockerImages {
		go func(fileName string) {
			// Use the more verbose API to exec so we can set a timeout
			params := exechelper.ExecParams{
				CmdName: "docker",
				CmdArgs: []string{"load", "-i", fileName},
				Timeout: 300, // 5 min timeout
			}
			result := ctx.Exec.RunCommand(params)
			log.Debug("stdout: " + result.OutBuf.String())
			log.Debug("stderr: " + result.ErrBuf.String())
			if result.Error != nil {
				errorChan <- fmt.Errorf("failed to docker load %s: %s", fileName, result.ErrBuf.String())
			} else {
				errorChan <- nil
			}
		}(imageFile)
	}

	// Wait for either an error or nil back from each one
	expectedCount := len(dockerImages)
	currentCount := 0
	hadErr := false
	for currentCount < expectedCount {
		err := <-errorChan
		currentCount++
		if err != nil {
			hadErr = true
			fmt.Println(err.Error())
		}
	}
	if hadErr {
		return errors.New("failed to load docker images")
	}
	return nil
}
