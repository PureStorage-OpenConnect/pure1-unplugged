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

package http

import (
	"fmt"
	"net/http"
	"strings"
)

// GetRequestTokenHeader gets a token of the specified type ("Bearer", "Basic") from
// the headers of a request
func GetRequestTokenHeader(header http.Header, tokenType string) (string, error) {
	authHeader := header.Get("Authorization") // Returns "" if header is missing

	if authHeader == "" {
		return "", fmt.Errorf("Authorization header not present")
	}

	if !strings.HasPrefix(authHeader, tokenType+" ") {
		return "", fmt.Errorf("Authorization header isn't %s token", tokenType)
	}

	return authHeader[len(tokenType)+1:], nil
}

// GetRequestTokenCookie gets a token of the specified type ("Bearer", "Basic") from
// the cookies of a request
func GetRequestTokenCookie(req *http.Request) (string, error) {
	authCookie, err := req.Cookie("pure1-unplugged-token")

	if err != nil {
		return "", err
	}

	return authCookie.Value, nil
}

// GetRequestAuthorizationToken gets a bearer token from either headers or cookies, or an
// error if it is contained in neither
func GetRequestAuthorizationToken(req *http.Request) (string, error) {
	headerToken, headerErr := GetRequestTokenHeader(req.Header, "Bearer")

	// Header takes precedence, since it was most likely manually set for a request
	// (vs. a cookie which is sticky and usually doesn't have to be manually set)
	if headerErr == nil {
		return headerToken, nil
	}

	// No header present, test cookie
	cookieToken, cookieErr := GetRequestTokenCookie(req)

	if cookieErr == nil {
		return cookieToken, nil
	}

	return "", fmt.Errorf("No token available in request")
}

// GetRequestAPIToken gets the API token from the given header
func GetRequestAPIToken(req *http.Request, headerName string) (string, error) {
	tokenHeader := req.Header.Get(headerName) // Returns "" if header is missing

	if tokenHeader == "" {
		return "", fmt.Errorf("API token header not present")
	}

	return tokenHeader, nil
}
