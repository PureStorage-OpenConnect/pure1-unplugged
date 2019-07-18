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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRequestTokenHeader(t *testing.T) {
	result, err := GetRequestTokenHeader(map[string][]string{
		"Authorization": {
			"Bearer asdf",
		},
	}, "Bearer")
	assert.NoError(t, err)
	assert.Equal(t, "asdf", result)
}

func TestGetRequestTokenHeaderMissing(t *testing.T) {
	_, err := GetRequestTokenHeader(map[string][]string{}, "Bearer")
	assert.Error(t, err)
}

func TestGetRequestTokenHeaderEmptyHeader(t *testing.T) {
	_, err := GetRequestTokenHeader(map[string][]string{
		"Authorization": {},
	}, "Bearer")
	assert.Error(t, err)
}

func TestGetRequestTokenHeaderWrongType(t *testing.T) {
	_, err := GetRequestTokenHeader(map[string][]string{
		"Authorization": {
			"Basic asdf",
		},
	}, "Bearer")
	assert.Error(t, err)
}

func TestGetRequestTokenHeaderNoType(t *testing.T) {
	_, err := GetRequestTokenHeader(map[string][]string{
		"Authorization": {
			"asdf",
		},
	}, "Bearer")
	assert.Error(t, err)
}

// Testing utility to make a new request with the given cookie
func newRequestWithCookie(name string, value string) (*http.Request, error) {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		return req, err
	}
	req.AddCookie(&http.Cookie{Name: name, Value: value})
	return req, nil
}

func newRequestWithHeader(name string, value string) (*http.Request, error) {
	req, err := http.NewRequest("GET", "", nil)
	if err != nil {
		return req, err
	}
	req.Header.Set(name, value)
	return req, nil
}

func TestGetRequestTokenCookie(t *testing.T) {
	req, err := newRequestWithCookie("pure1-unplugged-token", "asdf")
	assert.NoError(t, err)
	token, err := GetRequestTokenCookie(req)
	assert.NoError(t, err)
	assert.Equal(t, "asdf", token)
}

func TestGetRequestTokenCookieWrongCookie(t *testing.T) {
	req, err := newRequestWithCookie("notatoken", "asdf")
	assert.NoError(t, err)
	_, err = GetRequestTokenCookie(req)
	assert.Error(t, err)
}

func TestGetRequestAuthorizationTokenHeader(t *testing.T) {
	req, err := newRequestWithHeader("Authorization", "Bearer asdf")
	assert.NoError(t, err)
	token, err := GetRequestAuthorizationToken(req)
	assert.NoError(t, err)
	assert.Equal(t, "asdf", token)
}

func TestGetRequestAuthorizationTokenCookie(t *testing.T) {
	req, err := newRequestWithCookie("pure1-unplugged-token", "asdf")
	assert.NoError(t, err)
	token, err := GetRequestAuthorizationToken(req)
	assert.NoError(t, err)
	assert.Equal(t, "asdf", token)
}

func TestGetRequestAuthorizationTokenBoth(t *testing.T) {
	req, err := newRequestWithCookie("pure1-unplugged-token", "asdf")
	req.Header.Set("Authorization", "Bearer fdsa") // should match this, not the cookie
	assert.NoError(t, err)
	token, err := GetRequestAuthorizationToken(req)
	assert.NoError(t, err)
	assert.Equal(t, "fdsa", token)
}

func TestGetRequestAuthorizationTokenMissing(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	_, err = GetRequestAuthorizationToken(req)
	assert.Error(t, err)
}
