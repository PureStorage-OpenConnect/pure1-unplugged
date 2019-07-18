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
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/oidc"
)

// NewOIDCOAuth2Config creates a new oidc.OAuth2Config
// instance from the parameters in the given app
func (a *dexApp) NewOIDCOAuth2Config() oidc.OAuth2Config {
	return oidc.OAuth2Config{
		ClientID:     a.clientID,
		ClientSecret: a.clientSecret,
		Endpoint:     a.provider.Endpoint(),
		RedirectURL:  a.redirectURI,
	}
}
