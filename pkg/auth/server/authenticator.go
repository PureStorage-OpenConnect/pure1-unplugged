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
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/tokenstore"

	purehttp "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/http"
	oidc "github.com/coreos/go-oidc"
	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

var (
	refreshMutex = &sync.Mutex{}
)

// ParseJWT parses a JWT token from a string, checks for validity, and returns it
func ParseJWT(token string, hmacSecret string) (*jwt.Token, error) {
	return jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(hmacSecret), nil
	})
}

// Authorized tests an http request for the Authorization: Bearer [token] header
// and returns whether or not it's allowed to execute that request
func Authorized(a *dexApp, req *http.Request) (bool, error) {
	ctx := oidc.ClientContext(req.Context(), a.client)

	apiToken, err := purehttp.GetRequestAuthorizationToken(req)
	if err != nil {
		log.Debug("Couldn't find API token from request")
		return false, err
	}

	// Don't need to do anything with the token, just need to make sure it's a valid token and hasn't expired
	_, err = ParseJWT(apiToken, tokenstore.HmacSecret)
	if err != nil {
		log.WithError(err).Debug("Couldn't parse JWT")
		return false, err
	}

	userID, err := a.apiTokenStore.GetUserForToken(apiToken)
	if err != nil {
		log.Debug("API token not found")
		return false, err
	}

	if !a.apiTokenStore.HasUserCredentials(userID) {
		log.WithField("user", userID).Debug("API token valid and mapped to user, but user isn't associated with a valid token")
		return false, fmt.Errorf("API token valid and mapped to user, but user isn't associated with a valid token")
	}

	userToken, err := a.apiTokenStore.GetTokenForUser(userID)
	if err != nil {
		log.WithError(err).Debug("Failed to get OAuth token for user")
		return false, err
	}
	if userToken == nil {
		log.Debug("User token fetched, but was nil")
		return false, fmt.Errorf("User token fetched, but was nil")
	}

	if time.Now().After(userToken.Expiry) {
		log.WithField("user", userID).Debug("Refreshing token for user")
		newToken, err := a.oauth2config.ExchangeRefreshForToken(ctx, userToken)
		if err != nil {
			log.WithFields(log.Fields{
				"err":  err,
				"user": userID,
			}).Debug("Failed to refresh token for user")
			return false, err
		}
		// Update the stored token
		err = a.apiTokenStore.StoreUser(userID, newToken)
		if err != nil {
			log.WithError(err).Debug("Failed to store new token for user")
			return false, err
		}
		log.WithField("user", userID).Debug("Refreshed token for user successfully")
	}

	// If we got this far, we know the token is valid, since the tokens are only generated
	// internally and expiry is set by Dex and never changed by us.
	return true, nil
}

func getScopes(a *dexApp, r *http.Request) []string {
	scopes := []string{"openid", "profile", "email"}
	if extraScopes := r.FormValue("extra_scopes"); extraScopes != "" {
		scopes = append(scopes, strings.Split(extraScopes, " ")...)
	}

	var clients []string
	if crossClients := r.FormValue("cross_client"); crossClients != "" {
		clients = strings.Split(crossClients, " ")
	}
	for _, client := range clients {
		scopes = append(scopes, "audience:server:client_id:"+client)
	}

	if a.offlineAsScope {
		scopes = append(scopes, "offline_access")
	}

	return scopes
}

func generateStateToken(rd string, hmacSecret string, expiryTime int64) (string, error) {
	claims := tokenClaims{
		RedirectURL: rd,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiryTime,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenstore.HmacSecret))
}

// NewAuthCodeURL make a new authCodeURL with "rd" query, scopes, and stateToken
func NewAuthCodeURL(a *dexApp, r *http.Request) (string, error) {
	rd := r.FormValue("rd")
	// No redirect URL should just take them to the root URL
	if rd == "" {
		rd = "/"
	}

	// Construct the JWT state token, to be verified and used at the end of the authentication process
	// 5 minute expiry time to hopefully prevent any replay attacks on the authentication process
	tokenString, err := generateStateToken(rd, tokenstore.HmacSecret, time.Now().Add(time.Minute*5).Unix())

	if err != nil {
		// See explanation in handlers.go as to why this is wildly
		// unlikely to ever be tripped
		return "", err
	}

	scopes := getScopes(a, r)

	authCodeOptions := []oauth2.AuthCodeOption{}
	if r.FormValue("offline_access") == "yes" && !a.offlineAsScope {
		// We want a refresh token but we can't use a scope
		authCodeOptions = append(authCodeOptions, oauth2.AccessTypeOffline)
	}

	// use stateToken to pass original rd
	authCodeURL := a.oauth2config.GenerateAuthCodeURL(tokenString, scopes, authCodeOptions)
	return authCodeURL, nil
}

// GetOauth2Token obtains a new oauth2Token from dex
func GetOauth2Token(a *dexApp, r *http.Request) (*oauth2.Token, *jwt.Token, error) {
	ctx := oidc.ClientContext(r.Context(), a.client)

	// Authorization redirect callback from OAuth2 auth flow.
	// Catch any OAuth errors and bubble them up to our own handler
	if errMsg := r.FormValue("error"); errMsg != "" {
		return nil, nil, fmt.Errorf(errMsg + ": " + r.FormValue("error_description"))
	}

	code := r.FormValue("code")
	if code == "" {
		return nil, nil, fmt.Errorf("no code in request: %q", r.Form)
	}

	state := r.FormValue("state")
	if state == "" {
		return nil, nil, fmt.Errorf("No state parameter in token response")
	}

	jwt, err := ParseJWT(state, tokenstore.HmacSecret)

	// If there wasn't an error parsing, then the signature was valid
	if err != nil {
		return nil, nil, fmt.Errorf("Error parsing state token: %v", err)
	}

	oauthToken, err := a.oauth2config.ExchangeCodeForToken(ctx, code)

	return oauthToken, jwt, err
}
