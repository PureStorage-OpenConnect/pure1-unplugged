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
	oidc "github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

// Verifier is an implementation of the Verifier interface
type Verifier struct {
	verifier *oidc.IDTokenVerifier
}

// OAuth2Config is an implementation of the OAuth2Config interface
type OAuth2Config struct {
	ClientID, ClientSecret string
	Endpoint               oauth2.Endpoint
	RedirectURL            string
}
