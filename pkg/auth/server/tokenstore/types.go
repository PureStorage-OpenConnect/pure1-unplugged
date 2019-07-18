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

package tokenstore

import (
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"
)

var (
	// HmacSecret is generated randomly on startup and used to encode JWTs
	HmacSecret string
)

// APITokenClaims represents the claims stored in a Pure1 Unplugged API token
type APITokenClaims struct {
	Email      string `json:"email,omitempty"`
	Randomizer int    `json:"randomizer,omitempty"`
	jwt.StandardClaims
}

// APITokenStore provides an interface for all operations involved with API tokens and their persistence
type APITokenStore interface {
	// GenerateAPIToken generates a new API token for long-term use
	GenerateAPIToken(userID string, email string) (string, error)

	// GenerateSessionToken generates a token for a user session (short-lived)
	GenerateSessionToken(userID string, email string) (string, error)

	// GetAPITokenNames gets a list of all API token names in this store
	GetAPITokenNames() []string

	// ContainsAPIToken checks if this APITokenStore contains an API token with this name
	ContainsAPIToken(tokenName string) bool

	// GetUserForToken fetches which user this token represents. If the given token doesn't exist, returns an error
	GetUserForToken(apiToken string) (string, error)

	// StoreAPIToken registers an API token in this store mapped to the given user. WARNING: this WILL overwrite
	// existing token names, api tokens, or users (and thus can be used for upserts), check all relevant methods
	// before calling this
	StoreAPIToken(tokenName string, apiToken string, userID string) error

	// DeleteAPIToken deletes the API token with the given name. Note that "this token doesn't exist" is NOT an
	// error
	DeleteAPIToken(tokenName string) error

	// HasUserCredentials checks if this user's OAuth token is stored in this token store
	HasUserCredentials(userID string) bool

	// Remove this user from the mapping. This is usually called if a refresh fails and we want to remove it so we can
	// re-authenticate on next login (this way the user doesn't get locked out with bad credentials)
	InvalidateUser(userID string) error

	// StoreUser associates the given userID with their OAuth token. If the given user already has an OAuth token stored,
	// returns an error. WARNING: this WILL overwrite existing user IDs and tokens, this functions as an upsert.
	StoreUser(userID string, token *oauth2.Token) error

	// GetTokenForUser gets the OAuth token for the given user
	GetTokenForUser(userID string) (*oauth2.Token, error)
}
