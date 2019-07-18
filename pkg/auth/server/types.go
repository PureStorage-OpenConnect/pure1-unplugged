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

package server

import (
	"context"
	"net/http"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/tokenstore"
	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"

	oidc "github.com/coreos/go-oidc"
)

type tokenClaims struct {
	RedirectURL string `json:"rd,omitempty"`
	jwt.StandardClaims
}

// dexApp provides a dex client application structure
type dexApp struct {
	clientID     string
	clientSecret string
	redirectURI  string

	oauth2config OAuth2Config

	verifier      Verifier
	provider      *oidc.Provider
	apiTokenStore tokenstore.APITokenStore

	// Does the provider use "offline_access" scope to request a refresh token
	// or does it use "access_type=offline" (e.g. Google)?
	offlineAsScope bool

	client *http.Client
}

// Verifier provides an interface to verify a given ID token
type Verifier interface {
	Verify(ctx context.Context, rawIDToken string) error
}

// OAuth2Config provides an interface to perform standard OAuth2 functions with a provider
type OAuth2Config interface {
	GenerateAuthCodeURL(tokenString string, scopes []string, options []oauth2.AuthCodeOption) string
	ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error)
	ExchangeRefreshForToken(ctx context.Context, refresh *oauth2.Token) (*oauth2.Token, error)
	GetTokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource
}
