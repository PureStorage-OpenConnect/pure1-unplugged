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
	"sync"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/common/resources"
	log "github.com/sirupsen/logrus"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	typev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
)

const (
	// The key the data is stored in inside the secret
	secretDataKey = "value"

	// The name of the secret to save/load from. Note: must be lowercase alphanumeric, periods, or hyphens
	deviceTokenSecretName = "pure1-unplugged-device-token-secret"
)

// GetKubeSecretInterface creates a SecretInterface hooked up to the API server of
// the current cluster
func GetKubeSecretInterface(namespace string) (typev1.SecretInterface, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return clientset.CoreV1().Secrets(namespace), nil
}

// NewKubeSecretAPITokenStore generates a new kubeSecretAPITokenStore,
// also attempting to load the existing secret if possible (if not, it will
// start empty). Note that currently the secret will only be loaded once per run, so
// any changes made to the secret outside of this executable will be overwritten. This could
// easily be added with a secret watch, but isn't implemented yet since it's a relatively niche
// use case.
func NewKubeSecretAPITokenStore(secretAccess typev1.SecretInterface) resources.APITokenStorage {
	secretChan := make(chan bool, 1) // Buffer of 1 since we can write without blocking
	toReturn := kubeSecretDeviceTokenStorage{
		secretAccess:    secretAccess,
		mapLock:         &sync.Mutex{},
		secretWriteChan: secretChan,
		stopChan:        make(chan bool, 1),
	}
	// Ignore any errors we get back: if it errored out, it's handled in the next call to ensure
	// maps are initialized with empty values
	err := toReturn.parseFromSecret(deviceTokenSecretName)
	if err == nil {
		log.Debug("Loaded device token data successfully from secret!")
	} else {
		log.WithError(err).Error("Error parsing data from secret: continuing with blank device token info")
	}
	// Ensure map is initialized
	if toReturn.tokens == nil {
		toReturn.tokens = map[string]string{}
	}

	go toReturn.secretSaveLoop(deviceTokenSecretName)

	return &toReturn
}

func (k *kubeSecretDeviceTokenStorage) parseFromSecret(secretName string) error {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()

	secret, err := k.secretAccess.Get(secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	data, ok := secret.Data[secretDataKey]
	if !ok {
		return fmt.Errorf("Key %s not found in data", secretDataKey)
	}

	var loaded map[string]string
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		return err
	}

	k.tokens = loaded
	return nil
}

func (k *kubeSecretDeviceTokenStorage) secretSaveLoop(secretName string) {
	for true {
		select {
		case _ = <-k.secretWriteChan:
			{
				err := k.writeSecret(secretName)
				// Don't really handle the error at all, just let the logs know. It's not the end of the world if tokens are lost, and it's
				// not worth crashing over or anything. If it's a temporary error, it'll be rewritten with the next time anything changes (like a device being updated, deleted, etc.),
				// or if it's a permanent error then they'll see this and can debug more.
				if err != nil {
					log.WithError(err).Error("Error saving device token data to secret. No action is being taken, but be aware that device API tokens may be lost in case of a crash")
				} else {
					log.Debug("Saved to secret successfully")
				}
			}
		case _ = <-k.stopChan:
			// The channel says we should stop, so let's stop!
			return
		}
	}
}

func (k *kubeSecretDeviceTokenStorage) notifyToWrite() {
	select {
	case k.secretWriteChan <- true:
		// message sent
	default:
		// message dropped
	}
}

func (k *kubeSecretDeviceTokenStorage) writeSecret(secretName string) error {
	k.mapLock.Lock()
	marshalled, err := json.Marshal(k.tokens)
	k.mapLock.Unlock()
	if err != nil {
		return err
	}

	toSave := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: "pure1-unplugged",
		},
		Data: map[string][]byte{
			secretDataKey: marshalled,
		},
	}

	_, err = k.secretAccess.Update(toSave)
	// If we got an error, that's okay: that either means a) an operation failed or b) it didn't exist.
	// Either way, let's try and create a new one: if it already exists, that will fail too and we'll know
	if err != nil {
		_, err = k.secretAccess.Create(toSave)
		return err
	}
	return nil
}

// SaveToken associates the given API token with the given device ID, and returns
// if there's an error in the saving process.
func (k *kubeSecretDeviceTokenStorage) SaveToken(deviceID string, token string) error {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()

	k.tokens[deviceID] = token

	k.notifyToWrite()
	return nil
}

// HasToken checks if there's an API token associated with the given device ID.
// An error is not thrown if the device ID doesn't exist, but an error is thrown
// if there is an issue in the checking process itself.
func (k *kubeSecretDeviceTokenStorage) HasToken(deviceID string) (bool, error) {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()

	_, ok := k.tokens[deviceID]
	return ok, nil
}

// GetToken fetches the API token associated with the given device ID. An
// error is thrown if the key doesn't exist or there's an issue in the fetching
// process.
func (k *kubeSecretDeviceTokenStorage) GetToken(deviceID string) (string, error) {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()

	token, ok := k.tokens[deviceID]

	if !ok {
		return "", fmt.Errorf("Token not found for device ID %s", deviceID)
	}

	return token, nil
}

// DeleteToken deletes the token with the given device ID from this storage.
// Note that a nonexistent ID should *not* be considered an error, and the
// error return value is reserved for an issue with the actual deletion process
// itself.
func (k *kubeSecretDeviceTokenStorage) DeleteToken(deviceID string) error {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()

	delete(k.tokens, deviceID)

	k.notifyToWrite()

	return nil
}
