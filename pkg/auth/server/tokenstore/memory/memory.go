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

package memory

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/tokenstore"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
)

// NewInMemoryAPITokenStore generates a new InMemoryAPITokenStore
func NewInMemoryAPITokenStore() tokenstore.APITokenStore {
	return &inMemoryAPITokenStore{
		namesLock:  &sync.Mutex{},
		tokensLock: &sync.Mutex{},
		usersLock:  &sync.Mutex{},
		names:      map[string]string{},
		tokens:     map[string]string{},
		users:      map[string]*oauth2.Token{},
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

func (s *inMemoryAPITokenStore) isTokenUnique(token string) bool {
	s.tokensLock.Lock()
	defer s.tokensLock.Unlock()
	_, ok := s.tokens[token]
	return !ok
}

func (s *inMemoryAPITokenStore) generateUniqueToken(userID string, email string, expiry time.Duration) (string, error) {
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
		if s.isTokenUnique(generatedToken) {
			isUnique = true
			token = generatedToken
		}
	}
	return token, nil
}

// GenerateAPIToken generates a new API token for long-term use
func (s *inMemoryAPITokenStore) GenerateAPIToken(userID string, email string) (string, error) {
	return s.generateUniqueToken(userID, email, time.Hour*24*365*100) // Expire 100 years from now (just make it an obscenely far away expiration date)
}

// GenerateSessionToken generates a new API token for short-term use (like a web UI session)
func (s *inMemoryAPITokenStore) GenerateSessionToken(userID string, email string) (string, error) {
	return s.generateUniqueToken(userID, email, time.Hour) // Expire 1 hour from now (relatively short lived)
}

// GetAPITokenNames gets a list of all API token names in this store
func (s *inMemoryAPITokenStore) GetAPITokenNames() []string {
	s.namesLock.Lock()
	defer s.namesLock.Unlock()
	names := []string{}
	for key := range s.names {
		names = append(names, key)
	}
	return names
}

// ContainsAPIToken checks if this APITokenStore contains an API token with this name
func (s *inMemoryAPITokenStore) ContainsAPIToken(name string) bool {
	s.namesLock.Lock()
	defer s.namesLock.Unlock()
	_, ok := s.names[name]
	return ok
}

// GetUserForToken fetches which user this token represents. If the given token doesn't exist, returns an error
func (s *inMemoryAPITokenStore) GetUserForToken(apiToken string) (string, error) {
	s.tokensLock.Lock()
	defer s.tokensLock.Unlock()
	if _, ok := s.tokens[apiToken]; !ok {
		return "", fmt.Errorf("API token not found")
	}
	return s.tokens[apiToken], nil
}

// StoreAPIToken registers an API token in this store mapped to the given user. WARNING: this WILL overwrite
// existing token names, api tokens, or users (and thus can be used for upserts), check all relevant methods
// before calling this. Also note that this makes no guarantee if the given user is authenticated or not,
// that needs to be handled separately
func (s *inMemoryAPITokenStore) StoreAPIToken(tokenName string, apiToken string, userID string) error {
	s.namesLock.Lock()
	s.tokensLock.Lock()
	defer s.namesLock.Unlock()
	defer s.tokensLock.Unlock()

	s.names[tokenName] = apiToken
	s.tokens[apiToken] = userID
	return nil
}

// DeleteAPIToken deletes the API token with the given name. Note that "this token doesn't exist" is NOT an
// error
func (s *inMemoryAPITokenStore) DeleteAPIToken(tokenName string) error {
	s.namesLock.Lock()
	defer s.namesLock.Unlock()
	if apiToken, ok := s.names[tokenName]; ok {
		s.tokensLock.Lock()
		defer s.tokensLock.Unlock()
		// Delete the API token -> user mapping
		// Don't worry about deleting the user->OAuth token mapping, it may be shared by multiple tokens
		delete(s.tokens, apiToken)
	}
	delete(s.names, tokenName)
	return nil
}

// HasUserCredentials checks if this user's OAuth token is stored in this token store
func (s *inMemoryAPITokenStore) HasUserCredentials(userID string) bool {
	s.usersLock.Lock()
	defer s.usersLock.Unlock()
	_, ok := s.users[userID]
	return ok
}

// Remove this user from the mapping. This is usually called if a refresh fails and we want to remove it so we can
// re-authenticate on next login (this way the user doesn't get locked out with bad credentials)
func (s *inMemoryAPITokenStore) InvalidateUser(userID string) error {
	s.usersLock.Lock()
	defer s.usersLock.Unlock()
	delete(s.users, userID)
	return nil
}

// StoreUser associates the given userID with their OAuth token. If the given user already has an OAuth token stored,
// returns an error. WARNING: this WILL overwrite existing user IDs and tokens, this functions as an upsert.
func (s *inMemoryAPITokenStore) StoreUser(userID string, token *oauth2.Token) error {
	s.usersLock.Lock()
	defer s.usersLock.Unlock()
	s.users[userID] = token
	return nil
}

// GetTokenForUser gets the OAuth token for the given user
func (s *inMemoryAPITokenStore) GetTokenForUser(userID string) (*oauth2.Token, error) {
	s.usersLock.Lock()
	defer s.usersLock.Unlock()
	if _, ok := s.users[userID]; !ok {
		return nil, fmt.Errorf("User ID not found")
	}
	return s.users[userID], nil
}
