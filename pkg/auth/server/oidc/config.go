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

package oidc

import (
	"context"

	"golang.org/x/oauth2"
)

// GenerateAuthCodeURL generates the URL to visit to authenticate the user
func (o *OAuth2Config) GenerateAuthCodeURL(state string, scopes []string, options []oauth2.AuthCodeOption) string {
	cfg := oauth2.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		Endpoint:     o.Endpoint,
		Scopes:       scopes,
		RedirectURL:  o.RedirectURL,
	}
	return cfg.AuthCodeURL(state, options...)
}

// ExchangeCodeForToken takes the final step in the OAuth flow and fetches an ID token
func (o *OAuth2Config) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	cfg := oauth2.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		Endpoint:     o.Endpoint,
		RedirectURL:  o.RedirectURL,
	}
	return cfg.Exchange(ctx, code)
}

// ExchangeRefreshForToken exchanges a refresh token for a standard token, and thus can be used to extend sessions
func (o *OAuth2Config) ExchangeRefreshForToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	cfg := oauth2.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		Endpoint:     o.Endpoint,
		RedirectURL:  o.RedirectURL,
	}
	return cfg.TokenSource(ctx, token).Token()
}

// GetTokenSource returns a token source to get a token from, based on the provided token
func (o *OAuth2Config) GetTokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource {
	cfg := oauth2.Config{
		ClientID:     o.ClientID,
		ClientSecret: o.ClientSecret,
		Endpoint:     o.Endpoint,
		RedirectURL:  o.RedirectURL,
	}
	return cfg.TokenSource(ctx, token)
}
