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

package mock

import (
	"context"

	"golang.org/x/oauth2"
)

// Verify is a mocked stub
func (v *Verifier) Verify(ctx context.Context, rawIDToken string) error {
	args := v.Called(ctx, rawIDToken)
	return args.Error(0)
}

// GenerateAuthCodeURL is a mocked stub
func (o *OAuth2Config) GenerateAuthCodeURL(tokenString string, scopes []string, options []oauth2.AuthCodeOption) string {
	args := o.Called(tokenString, scopes, options)
	return args.String(0)
}

// ExchangeCodeForToken is a mocked stub
func (o *OAuth2Config) ExchangeCodeForToken(ctx context.Context, code string) (*oauth2.Token, error) {
	args := o.Called(ctx, code)
	token := args.Get(0)
	if token == nil {
		return nil, args.Error(1)
	}
	return token.(*oauth2.Token), args.Error(1)
}

// ExchangeRefreshForToken is a mocked stub
func (o *OAuth2Config) ExchangeRefreshForToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	args := o.Called(ctx, token)
	t := args.Get(0)
	if t == nil {
		return nil, args.Error(1)
	}
	return t.(*oauth2.Token), args.Error(1)
}

// GetTokenSource is a mocked stub
func (o *OAuth2Config) GetTokenSource(ctx context.Context, token *oauth2.Token) oauth2.TokenSource {
	args := o.Called(ctx, token)
	return args.Get(0).(oauth2.TokenSource)
}
