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

import "golang.org/x/oauth2"

// GenerateAPIToken is a mocked stub
func (m *TokenStore) GenerateAPIToken(userID string, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

// GenerateSessionToken is a mocked stub
func (m *TokenStore) GenerateSessionToken(userID string, email string) (string, error) {
	args := m.Called(userID, email)
	return args.String(0), args.Error(1)
}

// GetAPITokenNames is a mocked stub
func (m *TokenStore) GetAPITokenNames() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

// ContainsAPIToken is a mocked stub
func (m *TokenStore) ContainsAPIToken(tokenName string) bool {
	args := m.Called(tokenName)
	return args.Bool(0)
}

// GetUserForToken is a mocked stub
func (m *TokenStore) GetUserForToken(apiToken string) (string, error) {
	args := m.Called(apiToken)
	return args.String(0), args.Error(1)
}

// StoreAPIToken is a mocked stub
func (m *TokenStore) StoreAPIToken(tokenName string, apiToken string, userID string) error {
	args := m.Called(tokenName, apiToken, userID)
	return args.Error(0)
}

// DeleteAPIToken is a mocked stub
func (m *TokenStore) DeleteAPIToken(tokenName string) error {
	args := m.Called(tokenName)
	return args.Error(0)
}

// HasUserCredentials is a mocked stub
func (m *TokenStore) HasUserCredentials(userID string) bool {
	args := m.Called(userID)
	return args.Bool(0)
}

// InvalidateUser is a mocked stub
func (m *TokenStore) InvalidateUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

// StoreUser is a mocked stub
func (m *TokenStore) StoreUser(userID string, token *oauth2.Token) error {
	args := m.Called(userID, token)
	return args.Error(0)
}

// GetTokenForUser is a mocked stub
func (m *TokenStore) GetTokenForUser(userID string) (*oauth2.Token, error) {
	args := m.Called(userID)
	rawToken := args.Get(0)
	if rawToken == nil {
		return nil, args.Error(1)
	}
	return rawToken.(*oauth2.Token), args.Error(1)
}
