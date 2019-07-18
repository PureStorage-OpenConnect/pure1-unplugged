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
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	purehttp "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/http"
	pureerrors "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/http/errors"
	jwt "github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

func (a *dexApp) handleVerify(w http.ResponseWriter, r *http.Request) {
	authorized, _ := Authorized(a, r)

	if authorized {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Authorized"))
	} else {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}
}

func (a *dexApp) handleLogin(w http.ResponseWriter, r *http.Request) {
	authCodeURL, err := NewAuthCodeURL(a, r)
	if err != nil {
		// This is highly unlikely to happen: the only way it can happen is if
		// somehow the JWT fails to sign, which in turn is only because of
		// JSON marshalling errors or some issue with an invalid HMAC Secret,
		// which after some digging is incredibly unlikely and basically means
		// a core static variable didn't get initialized (the hash method itself
		// is unavailable)
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}

	log.WithField("redirect_uri", authCodeURL).Debug("/login redirecting")
	http.Redirect(w, r, authCodeURL, http.StatusTemporaryRedirect)
}

func getSubjectAndEmailFromIDToken(w http.ResponseWriter, idToken string) (string, string, *pureerrors.HTTPErr) {
	parsedIDToken, err := jwt.Parse(idToken, nil)
	if err != nil {
		if valError, ok := err.(*jwt.ValidationError); ok && valError.Errors&jwt.ValidationErrorMalformed != 0 {
			http.Error(w, fmt.Sprintf("error parsing JWT: %v", err), http.StatusBadRequest)
			return "", "", pureerrors.MakeBadRequestHTTPErr(err)
		}
	}

	idTokenClaims, ok := parsedIDToken.Claims.(jwt.MapClaims)
	if !ok {
		return "", "", pureerrors.MakeInternalHTTPErr(fmt.Errorf("error converting ID token claims to MapClaims"))
	}

	userID, ok := idTokenClaims["sub"].(string)
	if !ok {
		return "", "", pureerrors.MakeBadRequestHTTPErr(fmt.Errorf("error fetching subject from ID token"))
	}

	email, ok := idTokenClaims["email"].(string)
	if !ok {
		return "", "", pureerrors.MakeBadRequestHTTPErr(fmt.Errorf("error fetching email from ID token"))
	}

	return userID, email, nil
}

// validateCallbackRequest checks that the state token is valid and retrieves the token, id token and redirect url
func (a *dexApp) validateCallbackRequest(r *http.Request) (*oauth2.Token, string, string, *pureerrors.HTTPErr) {
	token, stateToken, err := GetOauth2Token(a, r)
	if err != nil {
		return nil, "", "", pureerrors.MakeBadRequestHTTPErr(err)
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return nil, "", "", pureerrors.MakeBadRequestHTTPErr(fmt.Errorf("no id_token in token response"))
	}

	claims, ok := stateToken.Claims.(jwt.MapClaims)
	if !ok || !stateToken.Valid {
		return nil, "", "", pureerrors.MakeInternalHTTPErr(fmt.Errorf("state token claims invalid"))
		// While we could proceed alright from here, since we're just getting the redirect URL,
		// this is an indicator that something is definitely not right and we should be wary
	}

	rd, ok := claims["rd"].(string)
	if !ok {
		return nil, "", "", pureerrors.MakeBadRequestHTTPErr(fmt.Errorf("no rd claim present in state token"))
	}

	return token, rawIDToken, rd, nil
}

func (a *dexApp) handleCallback(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, fmt.Sprintf("method not supported: %s, please use GET", r.Method), http.StatusMethodNotAllowed)
		return
	}

	token, rawIDToken, rd, validationErr := a.validateCallbackRequest(r)

	if validationErr != nil {
		http.Error(w, fmt.Sprintf("error validating callback request: %v", validationErr.Error()), validationErr.Code)
		log.WithError(validationErr).Error("Error validating callback request")
		return
	}

	// Also handles the http response itself, simply moved the code into a submethod
	userID, email, fieldError := getSubjectAndEmailFromIDToken(w, rawIDToken)
	if fieldError != nil {
		http.Error(w, fieldError.Error(), fieldError.Code)
		log.WithError(fieldError).Error("Error getting userID or email from request")
		return
	}

	apiToken, err := a.apiTokenStore.GenerateSessionToken(userID, email)
	if err != nil {
		http.Error(w, fmt.Sprintf("error generating API token: %v\n", err), http.StatusInternalServerError)
		return
	}

	log.WithField("token", apiToken).Debug("Generated API token")

	err = a.apiTokenStore.StoreAPIToken(fmt.Sprintf("_session_%s", userID), apiToken, userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("error storing API token: %v", err), http.StatusInternalServerError)
		// Try deleting the one we just registered
		delErr := a.apiTokenStore.DeleteAPIToken(fmt.Sprintf("_session_%s", userID))
		if delErr != nil {
			log.WithFields(log.Fields{
				"deletion_error":   delErr,
				"registration_err": err,
			}).Warn("API token registration failed, and so did deletion. There's a slim chance that some remnant may be left over depending on internal implementation")
		}
		return
	}
	// Update the stored credentials: even if we had old ones, these are even newer and are more guaranteed to be successful (like if the old one was revoked or something)
	err = a.apiTokenStore.StoreUser(userID, token)
	if err != nil {
		http.Error(w, fmt.Sprintf("error storing user token in server: %v", err), http.StatusInternalServerError)
		// We won't try to delete the user token here, since some token is better than none in this case (it *might* still work)
		return
	}

	log.Debug("Setting API token cookie")
	tokenCookie := http.Cookie{Name: "pure1-unplugged-token", Value: apiToken, Path: "/"}
	http.SetCookie(w, &tokenCookie)

	log.Debug("Setting session name cookie")
	sessionNameCookie := http.Cookie{Name: "pure1-unplugged-session", Value: fmt.Sprintf("_session_%s", userID), Path: "/"}
	http.SetCookie(w, &sessionNameCookie)

	// redirect to original URL
	log.WithField("rd", rd).Debug("Redirecting browser to rd url")
	http.Redirect(w, r, rd, http.StatusSeeOther)
}

func (a *dexApp) handleAPIToken(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Request to list all API token names
		toFormat := map[string]interface{}{
			"tokens": a.apiTokenStore.GetAPITokenNames(),
		}
		bytes, err := json.Marshal(toFormat)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Error getting list of API tokens: %v\n", err)))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(bytes)
		break
	case "POST":
		// Request to create an API token with name and given refresh token
		var name string
		if name = strings.TrimSpace(r.FormValue("name")); len(name) == 0 {
			// Name is empty or nonexistent
			http.Error(w, "name parameter missing. Please specify the name of the API token you wish to create\n", http.StatusBadRequest)
			return
		}

		contained := a.apiTokenStore.ContainsAPIToken(name)
		if contained {
			http.Error(w, fmt.Sprintf("API token %s already exists\n", name), http.StatusBadRequest)
			return
		}

		userToken, err := purehttp.GetRequestAuthorizationToken(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("error getting user token from request: %v\n", err), http.StatusBadRequest)
			return
		}

		userID, err := a.apiTokenStore.GetUserForToken(userToken)
		if err != nil {
			http.Error(w, fmt.Sprintf("error getting user ID for token: %v\n", err), http.StatusBadRequest)
			return
		}

		generatedToken, err := a.apiTokenStore.GenerateAPIToken(userID, "asdf@example.com")
		if err != nil {
			http.Error(w, fmt.Sprintf("error generating API token: %v\n", err), http.StatusInternalServerError)
			return
		}
		err = a.apiTokenStore.StoreAPIToken(name, generatedToken, userID)
		if err != nil {
			http.Error(w, fmt.Sprintf("error storing new API token: %v\n", err), http.StatusInternalServerError)
			delErr := a.apiTokenStore.DeleteAPIToken(name)
			if delErr != nil {
				log.WithFields(log.Fields{
					"name":           name,
					"store_error":    err,
					"deletion_error": delErr,
				}).Warn("Error storing API token, and second error on deleting it. There may possibly be remnants left over")
			}
		}

		result := map[string]interface{}{
			"api_token":  generatedToken,
			"token_name": name,
			"warning":    "Please make sure to store this API token securely, you won't be able to see it again after you leave this page.",
		}

		marshalled, err := json.Marshal(result)

		if err != nil {
			http.Error(w, fmt.Sprintf("error marshalling result: %v\n", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(marshalled)

		break
	case "DELETE":
		// Request to delete an API token by name
		var name string
		if name = r.FormValue("name"); len(strings.TrimSpace(name)) == 0 {
			// Name is empty or nonexistent
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("name parameter missing. Please specify the name of the API token you wish to delete\n"))
			return
		}
		err := a.apiTokenStore.DeleteAPIToken(name)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("error deleting API token: %v\n", err)))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf("API token %s successfully deleted (if it existed)\n", name)))
		break
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(fmt.Sprintf("Method not supported: %s. Please use GET, POST, or DELETE\n", r.Method)))
	}
}
