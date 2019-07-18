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
	"math/rand"
	"sync"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/tokenstore"
	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
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
	authTokenSecretName = "pure1-unplugged-auth-token-secret"
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
func NewKubeSecretAPITokenStore(secretAccess typev1.SecretInterface) tokenstore.APITokenStore {
	secretChan := make(chan bool, 1) // Buffer of 1 since we can write without blocking
	toReturn := kubeSecretAPITokenStore{
		secretAccess:    secretAccess,
		mapLock:         &sync.Mutex{},
		secretWriteChan: secretChan,
		stopChan:        make(chan bool, 1),
	}
	// Ignore any errors we get back: if it errored out, it's handled in the next call to ensure
	// maps are initialized with empty values
	err := toReturn.parseFromSecret(authTokenSecretName)
	if err == nil {
		log.Debug("Loaded auth data successfully from secret!")
	} else {
		log.WithError(err).Error("Error parsing data from secret: continuing with blank auth info")
	}
	toReturn.ensureMapsInitialized()

	go toReturn.secretSaveLoop(authTokenSecretName)

	return &toReturn
}

func (k *kubeSecretAPITokenStore) notifyToWrite() {
	select {
	case k.secretWriteChan <- true:
		// message sent
	default:
		// message dropped
	}
}

func (k *kubeSecretAPITokenStore) parseFromSecret(secretName string) error {
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

	var loaded secretData
	err = json.Unmarshal(data, &loaded)
	if err != nil {
		return err
	}

	k.names = loaded.Names
	k.tokens = loaded.Tokens
	k.users = loaded.Users
	return nil
}

func (k *kubeSecretAPITokenStore) secretSaveLoop(secretName string) {
	for true {
		select {
		case _ = <-k.secretWriteChan:
			{
				err := k.writeSecret(secretName)
				// Don't really handle the error at all, just let the logs know. It's not the end of the world if sessions/tokens are lost, and it's
				// not worth crashing over or anything. If it's a temporary error, it'll be rewritten with the next time anything changes (like someone logging in),
				// or if it's a permanent error then they'll see this and can debug more.
				if err != nil {
					log.WithError(err).Error("Error saving auth data to secret. No action is being taken, but be aware that sessions may be lost in case of a crash")
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

func (k *kubeSecretAPITokenStore) writeSecret(secretName string) error {
	k.mapLock.Lock()
	marshalled, err := json.Marshal(secretData{
		Names:  k.names,
		Tokens: k.tokens,
		Users:  k.users,
	})
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

func (k *kubeSecretAPITokenStore) ensureMapsInitialized() {
	if k.names == nil {
		k.names = map[string]string{}
	}
	if k.tokens == nil {
		k.tokens = map[string]string{}
	}
	if k.users == nil {
		k.users = map[string]*oauth2.Token{}
	}
}

func getClaims(userID string, email string, expiry time.Duration) tokenstore.APITokenClaims {
	expiryTime := time.Now().Add(expiry)
	return tokenstore.APITokenClaims{
		Email:      email,
		Randomizer: rand.Int(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiryTime.Unix(),
			Subject:   userID,
		},
	}
}

func (k *kubeSecretAPITokenStore) isTokenUnique(token string) bool {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()
	_, ok := k.tokens[token]
	return !ok
}

func (k *kubeSecretAPITokenStore) generateUniqueToken(userID string, email string, expiry time.Duration) (string, error) {
	var token string

	isUnique := false

	for !isUnique {
		claims := getClaims(userID, email, expiry)
		t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		generatedToken, err := t.SignedString([]byte(tokenstore.HmacSecret))
		if err != nil {
			return "", err
		}

		// If the key isn't okay, it means it doesn't exist, which means this token is unique
		if k.isTokenUnique(generatedToken) {
			isUnique = true
			token = generatedToken
		}
	}
	return token, nil
}

// GenerateAPIToken generates a new API token for long-term use
func (k *kubeSecretAPITokenStore) GenerateAPIToken(userID string, email string) (string, error) {
	return k.generateUniqueToken(userID, email, time.Hour*24*365*100) // Expire 100 years from now (just make it an obscenely far away expiration date)
}

// GenerateSessionToken generates a new API token for short-term use (like a web UI session)
func (k *kubeSecretAPITokenStore) GenerateSessionToken(userID string, email string) (string, error) {
	return k.generateUniqueToken(userID, email, time.Hour) // Expire 1 hour from now (relatively short lived)
}

// GetAPITokenNames gets a list of all API token names in this store
func (k *kubeSecretAPITokenStore) GetAPITokenNames() []string {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()
	names := []string{}
	for key := range k.names {
		names = append(names, key)
	}
	return names
}

// ContainsAPIToken checks if this APITokenStore contains an API token with this name
func (k *kubeSecretAPITokenStore) ContainsAPIToken(name string) bool {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()
	_, ok := k.names[name]
	return ok
}

// GetUserForToken fetches which user this token represents. If the given token doesn't exist, returns an error
func (k *kubeSecretAPITokenStore) GetUserForToken(apiToken string) (string, error) {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()
	if _, ok := k.tokens[apiToken]; !ok {
		return "", fmt.Errorf("API token not found")
	}
	return k.tokens[apiToken], nil
}

// StoreAPIToken registers an API token in this store mapped to the given user. WARNING: this WILL overwrite
// existing token names, api tokens, or users (and thus can be used for upserts), check all relevant methods
// before calling this. Also note that this makes no guarantee if the given user is authenticated or not,
// that needs to be handled separately
func (k *kubeSecretAPITokenStore) StoreAPIToken(tokenName string, apiToken string, userID string) error {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()

	k.names[tokenName] = apiToken
	k.tokens[apiToken] = userID

	k.notifyToWrite()

	return nil
}

// DeleteAPIToken deletes the API token with the given name. Note that "this token doesn't exist" is NOT an
// error
func (k *kubeSecretAPITokenStore) DeleteAPIToken(tokenName string) error {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()
	if apiToken, ok := k.names[tokenName]; ok {
		// Delete the API token -> user mapping
		// Don't worry about deleting the user->OAuth token mapping, it may be shared by multiple tokens
		delete(k.tokens, apiToken)
	}
	delete(k.names, tokenName)

	k.notifyToWrite()

	return nil
}

// HasUserCredentials checks if this user's OAuth token is stored in this token store
func (k *kubeSecretAPITokenStore) HasUserCredentials(userID string) bool {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()
	_, ok := k.users[userID]
	return ok
}

// Remove this user from the mapping. This is usually called if a refresh fails and we want to remove it so we can
// re-authenticate on next login (this way the user doesn't get locked out with bad credentials)
func (k *kubeSecretAPITokenStore) InvalidateUser(userID string) error {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()
	delete(k.users, userID)

	k.notifyToWrite()

	return nil
}

// StoreUser associates the given userID with their OAuth token. If the given user already has an OAuth token stored,
// returns an error. WARNING: this WILL overwrite existing user IDs and tokens, this functions as an upsert.
func (k *kubeSecretAPITokenStore) StoreUser(userID string, token *oauth2.Token) error {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()
	k.users[userID] = token

	k.notifyToWrite()

	return nil
}

// GetTokenForUser gets the OAuth token for the given user
func (k *kubeSecretAPITokenStore) GetTokenForUser(userID string) (*oauth2.Token, error) {
	k.mapLock.Lock()
	defer k.mapLock.Unlock()
	if _, ok := k.users[userID]; !ok {
		return nil, fmt.Errorf("User ID not found")
	}
	return k.users[userID], nil
}
