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
	"bytes"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/oauth2"

	oidcmock "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/mock"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/tokenstore"
	tokenstoremock "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/tokenstore/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGetScopes(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	a := dexApp{}

	scopes := getScopes(&a, req)
	assert.ElementsMatch(t, scopes, []string{"openid", "profile", "email"})
}

func TestGetScopesExtras(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("extra_scopes", "ascope anotherscope yetanotherscope")

	a := dexApp{}

	scopes := getScopes(&a, req)
	assert.ElementsMatch(t, scopes, []string{"openid", "profile", "email", "ascope", "anotherscope", "yetanotherscope"})
}

func TestGetScopesClients(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("cross_client", "client client2 client3")

	a := dexApp{}

	scopes := getScopes(&a, req)
	assert.ElementsMatch(t, scopes, []string{"openid", "profile", "email", "audience:server:client_id:client", "audience:server:client_id:client2", "audience:server:client_id:client3"})
}

func TestGetScopesOffline(t *testing.T) {
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	a := dexApp{
		offlineAsScope: true,
	}

	scopes := getScopes(&a, req)
	assert.ElementsMatch(t, scopes, []string{"openid", "profile", "email", "offline_access"})
}

func TestAuthorizedNoToken(t *testing.T) {
	a := dexApp{}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)

	authorized, err := Authorized(&a, req)
	assert.Error(t, err)
	assert.False(t, authorized)
}

func TestAuthorizedNonexistentToken(t *testing.T) {
	tokenStore := &tokenstoremock.TokenStore{}

	tokenStore.On("GetUserForToken", "token").Return("", fmt.Errorf("Some error"))

	a := dexApp{
		apiTokenStore: tokenStore,
	}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer token")

	authorized, err := Authorized(&a, req)
	assert.Error(t, err)
	assert.False(t, authorized)
}

func TestAuthorizedUnauthorizedUser(t *testing.T) {
	tokenStore := &tokenstoremock.TokenStore{}

	tokenStore.On("GetUserForToken", "token").Return("userid", nil)
	tokenStore.On("HasUserCredentials", "userid").Return(false)

	a := dexApp{
		apiTokenStore: tokenStore,
	}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer token")

	authorized, err := Authorized(&a, req)
	assert.Error(t, err)
	assert.False(t, authorized)
}

func TestAuthorizedTokenFetchError(t *testing.T) {
	tokenStore := &tokenstoremock.TokenStore{}

	tokenStore.On("GetUserForToken", "token").Return("userid", nil)
	tokenStore.On("HasUserCredentials", "userid").Return(true)
	tokenStore.On("GetTokenForUser", "userid").Return(nil, fmt.Errorf("Some error"))

	a := dexApp{
		apiTokenStore: tokenStore,
	}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer token")

	authorized, err := Authorized(&a, req)
	assert.Error(t, err)
	assert.False(t, authorized)
}

func TestAuthorizedTokenFetchNil(t *testing.T) {
	tokenStore := &tokenstoremock.TokenStore{}

	tokenStore.On("GetUserForToken", "token").Return("userid", nil)
	tokenStore.On("HasUserCredentials", "userid").Return(true)
	tokenStore.On("GetTokenForUser", "userid").Return(nil, nil)

	a := dexApp{
		apiTokenStore: tokenStore,
	}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer token")

	authorized, err := Authorized(&a, req)
	assert.Error(t, err)
	assert.False(t, authorized)
}

func TestAuthorizedUnexpired(t *testing.T) {
	tokenstore.HmacSecret = "This is totally secret"

	tokenStore := &tokenstoremock.TokenStore{}

	authToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	tokenStore.On("GetUserForToken", authToken).Return("some_user_id", nil)
	tokenStore.On("HasUserCredentials", "some_user_id").Return(true)
	tokenStore.On("GetTokenForUser", "some_user_id").Return(&oauth2.Token{Expiry: time.Now().Add(time.Hour)}, nil)

	a := dexApp{
		apiTokenStore: tokenStore,
	}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	authorized, err := Authorized(&a, req)
	assert.NoError(t, err)
	assert.True(t, authorized)
}

func TestAuthorizedExpiredUserToken(t *testing.T) {
	tokenstore.HmacSecret = "This is totally secret"

	authToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(-time.Hour).Unix(),
	}).SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	a := dexApp{}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	authorized, err := Authorized(&a, req)
	assert.Error(t, err)
	assert.False(t, authorized)
}

func TestAuthorizedNoTokenUser(t *testing.T) {
	tokenstore.HmacSecret = "This is totally secret"

	tokenStore := &tokenstoremock.TokenStore{}

	authToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	tokenStore.On("GetUserForToken", authToken).Return("", fmt.Errorf("Some error"))

	a := dexApp{
		apiTokenStore: tokenStore,
	}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	authorized, err := Authorized(&a, req)
	assert.Error(t, err)
	assert.False(t, authorized)
}

func TestAuthorizedNoBackingCredentials(t *testing.T) {
	tokenstore.HmacSecret = "This is totally secret"

	tokenStore := &tokenstoremock.TokenStore{}

	authToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	tokenStore.On("GetUserForToken", authToken).Return("some_user_id", nil)
	tokenStore.On("HasUserCredentials", "some_user_id").Return(false)

	a := dexApp{
		apiTokenStore: tokenStore,
	}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	authorized, err := Authorized(&a, req)
	assert.Error(t, err)
	assert.False(t, authorized)
}

func TestAuthorizedErrorGettingBackingCredentials(t *testing.T) {
	tokenstore.HmacSecret = "This is totally secret"

	tokenStore := &tokenstoremock.TokenStore{}

	authToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	tokenStore.On("GetUserForToken", authToken).Return("some_user_id", nil)
	tokenStore.On("HasUserCredentials", "some_user_id").Return(true)
	tokenStore.On("GetTokenForUser", "some_user_id").Return(nil, fmt.Errorf("Some error"))

	a := dexApp{
		apiTokenStore: tokenStore,
	}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	authorized, err := Authorized(&a, req)
	assert.Error(t, err)
	assert.False(t, authorized)
}

func TestAuthorizedNilBackingCredentials(t *testing.T) {
	tokenstore.HmacSecret = "This is totally secret"

	tokenStore := &tokenstoremock.TokenStore{}

	authToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	tokenStore.On("GetUserForToken", authToken).Return("some_user_id", nil)
	tokenStore.On("HasUserCredentials", "some_user_id").Return(true)
	tokenStore.On("GetTokenForUser", "some_user_id").Return(nil, nil)

	a := dexApp{
		apiTokenStore: tokenStore,
	}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	authorized, err := Authorized(&a, req)
	assert.Error(t, err)
	assert.False(t, authorized)
}

func TestAuthorizedExpiredRefreshFailure(t *testing.T) {
	tokenStore := &tokenstoremock.TokenStore{}

	tokenStore.On("GetUserForToken", "token").Return("some_user_id", nil)
	tokenStore.On("HasUserCredentials", "some_user_id").Return(true)

	token := &oauth2.Token{Expiry: time.Now().Add(-time.Hour)}

	tokenStore.On("GetTokenForUser", "some_user_id").Return(token, nil)

	oauthConf := &oidcmock.OAuth2Config{}

	oauthConf.On("ExchangeRefreshForToken", mock.Anything, token).Return(nil, fmt.Errorf("Some error"))

	a := dexApp{
		apiTokenStore: tokenStore,
		oauth2config:  oauthConf,
	}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)
	req.Header.Set("Authorization", "Bearer token")

	authorized, err := Authorized(&a, req)
	assert.Error(t, err)
	assert.False(t, authorized)
}

func TestAuthorizedExpiredStoreFailure(t *testing.T) {
	tokenstore.HmacSecret = "This is totally secret"

	tokenStore := &tokenstoremock.TokenStore{}

	authToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	tokenStore.On("GetUserForToken", authToken).Return("some_user_id", nil)
	tokenStore.On("HasUserCredentials", "some_user_id").Return(true)

	token := &oauth2.Token{Expiry: time.Now().Add(-time.Hour)}

	tokenStore.On("GetTokenForUser", "some_user_id").Return(token, nil)
	tokenStore.On("StoreUser", "some_user_id", token).Return(fmt.Errorf("Some error"))

	oauthConf := &oidcmock.OAuth2Config{}

	oauthConf.On("ExchangeRefreshForToken", mock.Anything, token).Return(token, nil)

	a := dexApp{
		apiTokenStore: tokenStore,
		oauth2config:  oauthConf,
	}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	authorized, err := Authorized(&a, req)
	assert.Error(t, err)
	assert.False(t, authorized)
}

func TestAuthorizedExpiredSuccessfulRefresh(t *testing.T) {
	tokenstore.HmacSecret = "This is totally secret"

	tokenStore := &tokenstoremock.TokenStore{}

	authToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	tokenStore.On("GetUserForToken", authToken).Return("userid", nil)
	tokenStore.On("HasUserCredentials", "userid").Return(true)

	token := &oauth2.Token{Expiry: time.Now().Add(-time.Hour)}

	tokenStore.On("GetTokenForUser", "userid").Return(token, nil)
	tokenStore.On("StoreUser", "userid", token).Return(nil)

	oauthConf := &oidcmock.OAuth2Config{}

	oauthConf.On("ExchangeRefreshForToken", mock.Anything, token).Return(token, nil)

	a := dexApp{
		apiTokenStore: tokenStore,
		oauth2config:  oauthConf,
	}
	req, err := http.NewRequest("GET", "/auth", bytes.NewReader([]byte("")))
	assert.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	authorized, err := Authorized(&a, req)
	assert.NoError(t, err)
	assert.True(t, authorized)
}

func TestNewAuthCodeURL(t *testing.T) {
	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("GenerateAuthCodeURL", mock.AnythingOfType("string"), []string{"openid", "profile", "email"}, []oauth2.AuthCodeOption{}).Return("http://someurl")
	a := dexApp{
		oauth2config: &mockConfig,
	}

	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("rd", "asdf")

	url, err := NewAuthCodeURL(&a, req)
	assert.NoError(t, err)
	assert.Equal(t, "http://someurl", url)
}

func TestNewAuthCodeURLOfflineAsScope(t *testing.T) {
	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("GenerateAuthCodeURL", mock.AnythingOfType("string"), []string{"openid", "profile", "email", "offline_access"}, []oauth2.AuthCodeOption{}).Return("http://someurl")
	a := dexApp{
		oauth2config:   &mockConfig,
		offlineAsScope: true,
	}

	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("rd", "asdf")
	req.Form.Set("offline_access", "yes")

	url, err := NewAuthCodeURL(&a, req)
	assert.NoError(t, err)
	assert.Equal(t, "http://someurl", url)
}

func TestNewAuthCodeURLOfflineNotAsScope(t *testing.T) {
	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("GenerateAuthCodeURL", mock.AnythingOfType("string"), []string{"openid", "profile", "email"}, []oauth2.AuthCodeOption{oauth2.AccessTypeOffline}).Return("http://someurl")
	a := dexApp{
		oauth2config:   &mockConfig,
		offlineAsScope: false,
	}

	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("rd", "asdf")
	req.Form.Set("offline_access", "yes")

	url, err := NewAuthCodeURL(&a, req)
	assert.NoError(t, err)
	assert.Equal(t, "http://someurl", url)
}

func TestNewAuthCodeURLEmptyRedirect(t *testing.T) {
	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("GenerateAuthCodeURL", mock.AnythingOfType("string"), []string{"openid", "profile", "email"}, []oauth2.AuthCodeOption{}).Return("http://someurl")
	a := dexApp{
		oauth2config: &mockConfig,
	}

	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("rd", "")

	url, err := NewAuthCodeURL(&a, req)
	assert.NoError(t, err)
	assert.Equal(t, "http://someurl", url)
}

func TestNewAuthCodeURLNoRedirect(t *testing.T) {
	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("GenerateAuthCodeURL", mock.AnythingOfType("string"), []string{"openid", "profile", "email"}, []oauth2.AuthCodeOption{}).Return("http://someurl")
	a := dexApp{
		oauth2config: &mockConfig,
	}

	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)

	url, err := NewAuthCodeURL(&a, req)
	assert.NoError(t, err)
	assert.Equal(t, "http://someurl", url)
}

func TestGetAuthToken(t *testing.T) {
	tokenstore.HmacSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	stateToken, err := generateStateToken("http://192.168.99.100", tokenstore.HmacSecret, time.Now().Add(time.Minute*5).Unix())
	assert.NoError(t, err)
	assert.NotNil(t, stateToken)

	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("ExchangeCodeForToken", mock.Anything, "somecode").Return(new(oauth2.Token), nil)
	a := dexApp{
		oauth2config: &mockConfig,
	}

	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("state", stateToken)
	req.Form.Set("code", "somecode")

	token, jwt, err := GetOauth2Token(&a, req)
	assert.NoError(t, err)
	assert.NotNil(t, token)
	assert.NotNil(t, jwt)
}

func TestGetAuthTokenOauthError(t *testing.T) {
	a := dexApp{}

	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("error", "Some error!")

	_, _, err = GetOauth2Token(&a, req)
	assert.Error(t, err)
}

func TestGetAuthTokenMissingCode(t *testing.T) {
	a := dexApp{}

	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)

	_, _, err = GetOauth2Token(&a, req)
	assert.Error(t, err)
}

func TestGetAuthTokenMissingState(t *testing.T) {
	a := dexApp{}

	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("code", "somecode")

	_, _, err = GetOauth2Token(&a, req)
	assert.Error(t, err)
}

func TestGetAuthTokenInvalidState(t *testing.T) {
	tokenstore.HmacSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	stateToken, err := generateStateToken("http://192.168.99.100", tokenstore.HmacSecret, time.Now().Add(time.Minute*-5).Unix()) // Expired token
	assert.NoError(t, err)
	assert.NotNil(t, stateToken)

	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("ExchangeCodeForToken", mock.Anything, "somecode").Return(new(oauth2.Token), nil)
	a := dexApp{
		oauth2config: &mockConfig,
	}

	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("state", stateToken)
	req.Form.Set("code", "somecode")

	_, _, err = GetOauth2Token(&a, req)
	assert.Error(t, err)
}
