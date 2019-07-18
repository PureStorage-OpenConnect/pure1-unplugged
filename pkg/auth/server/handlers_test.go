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
	"net/http/httptest"
	"testing"
	"time"

	oidcmock "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/mock"
	"github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/tokenstore"
	tokenstoremock "github.com/PureStorage-OpenConnect/pure1-unplugged/pkg/auth/server/tokenstore/mock"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/oauth2"
)

func TestHandleVerifyFail(t *testing.T) {
	a := dexApp{}
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)

	a.handleVerify(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHandleVerifySuccess(t *testing.T) {
	tokenstore.HmacSecret = "This is totally secret"

	tokenStore := &tokenstoremock.TokenStore{}

	authToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.StandardClaims{
		ExpiresAt: time.Now().Add(time.Hour).Unix(),
	}).SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	tokenStore.On("GetUserForToken", authToken).Return("userid", nil)
	tokenStore.On("HasUserCredentials", "userid").Return(true)
	tokenStore.On("GetTokenForUser", "userid").Return(&oauth2.Token{Expiry: time.Now().Add(time.Hour)}, nil)

	a := dexApp{
		apiTokenStore: tokenStore,
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authToken))

	a.handleVerify(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandleLogin(t *testing.T) {
	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("GenerateAuthCodeURL", mock.AnythingOfType("string"), []string{"openid", "profile", "email"}, []oauth2.AuthCodeOption{}).Return("http://someurl")
	a := dexApp{
		oauth2config: &mockConfig,
	}
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("rd", "http://somewhere")

	a.handleLogin(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
}

func TestHandleLoginNoRedirectURL(t *testing.T) {
	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("GenerateAuthCodeURL", mock.AnythingOfType("string"), []string{"openid", "profile", "email"}, []oauth2.AuthCodeOption{}).Return("http://someurl")
	a := dexApp{
		oauth2config: &mockConfig,
	}
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)

	a.handleLogin(w, req)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
}

func TestValidateCallbackRequestTokenFailure(t *testing.T) {
	tokenstore.HmacSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	stateToken, err := generateStateToken("http://192.168.99.100", tokenstore.HmacSecret, time.Now().Add(time.Minute*5).Unix())
	assert.NoError(t, err)
	assert.NotNil(t, stateToken)

	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("ExchangeCodeForToken", mock.Anything, "somecode").Return(nil, fmt.Errorf("Some error"))
	a := dexApp{
		oauth2config: &mockConfig,
	}

	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("state", stateToken)
	req.Form.Set("code", "somecode")

	_, _, _, err = a.validateCallbackRequest(req)
	assert.NotNil(t, err)
}

func TestValidateCallbackRequestMissingIDToken(t *testing.T) {
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

	_, _, _, err = a.validateCallbackRequest(req)
	assert.NotNil(t, err)
}

func TestValidateCallbackRequestMissingRedirect(t *testing.T) {
	tokenstore.HmacSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	claims := tokenClaims{
		// No RedirectURL claim
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Minute * 5).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	stateToken, err := token.SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)
	assert.NotNil(t, stateToken)

	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("ExchangeCodeForToken", mock.Anything, "somecode").Return(new(oauth2.Token).WithExtra(map[string]interface{}{
		"id_token": "idtoken",
	}), nil)
	a := dexApp{
		oauth2config: &mockConfig,
	}

	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("state", stateToken)
	req.Form.Set("code", "somecode")

	_, _, _, err = a.validateCallbackRequest(req)
	assert.NotNil(t, err)
}

func TestValidateCallbackRequest(t *testing.T) {
	tokenstore.HmacSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	stateToken, err := generateStateToken("http://192.168.99.100", tokenstore.HmacSecret, time.Now().Add(time.Minute*5).Unix())
	assert.NoError(t, err)
	assert.NotNil(t, stateToken)

	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("ExchangeCodeForToken", mock.Anything, "somecode").Return(new(oauth2.Token).WithExtra(map[string]interface{}{
		"id_token": "idtoken",
	}), nil)
	a := dexApp{
		oauth2config: &mockConfig,
	}

	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("state", stateToken)
	req.Form.Set("code", "somecode")

	token, rawIDToken, rd, err := a.validateCallbackRequest(req)
	assert.Nil(t, err)
	assert.NotNil(t, token)
	assert.Equal(t, "idtoken", rawIDToken)
	assert.Equal(t, "http://192.168.99.100", rd)
}

func TestHandleCallbackInvalidMethod(t *testing.T) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("POST", "", nil)
	assert.NoError(t, err)

	a := dexApp{}

	a.handleCallback(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandleCallbackValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)

	a := dexApp{}

	a.handleCallback(w, req)

	assert.True(t, w.Code >= 400 && w.Code < 500)
}

func TestHandleCallbackErrorGettingUserID(t *testing.T) {
	tokenstore.HmacSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	stateToken, err := generateStateToken("http://192.168.99.100", tokenstore.HmacSecret, time.Now().Add(time.Minute*5).Unix())
	assert.NoError(t, err)
	assert.NotNil(t, stateToken)

	claims := jwt.MapClaims{
		"email": "pureuser@purestorage.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	idToken, err := token.SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("ExchangeCodeForToken", mock.Anything, "somecode").Return(new(oauth2.Token).WithExtra(map[string]interface{}{
		"id_token": idToken,
	}), nil)

	a := dexApp{
		oauth2config: &mockConfig,
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("state", stateToken)
	req.Form.Set("code", "somecode")

	a.handleCallback(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleCallbackErrorGettingSessionToken(t *testing.T) {
	tokenstore.HmacSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	stateToken, err := generateStateToken("http://192.168.99.100", tokenstore.HmacSecret, time.Now().Add(time.Minute*5).Unix())
	assert.NoError(t, err)
	assert.NotNil(t, stateToken)

	claims := jwt.MapClaims{
		"sub":   "some-user-id",
		"email": "pureuser@purestorage.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	idToken, err := token.SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("ExchangeCodeForToken", mock.Anything, "somecode").Return(new(oauth2.Token).WithExtra(map[string]interface{}{
		"id_token": idToken,
	}), nil)

	mockStore := tokenstoremock.TokenStore{}
	mockStore.On("GenerateSessionToken", "some-user-id", "pureuser@purestorage.com").Return("", fmt.Errorf("Some error"))

	a := dexApp{
		oauth2config:  &mockConfig,
		apiTokenStore: &mockStore,
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("state", stateToken)
	req.Form.Set("code", "somecode")

	a.handleCallback(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleCallbackErrorStoringToken(t *testing.T) {
	tokenstore.HmacSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	stateToken, err := generateStateToken("http://192.168.99.100", tokenstore.HmacSecret, time.Now().Add(time.Minute*5).Unix())
	assert.NoError(t, err)
	assert.NotNil(t, stateToken)

	claims := jwt.MapClaims{
		"sub":   "some-user-id",
		"email": "pureuser@purestorage.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	idToken, err := token.SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("ExchangeCodeForToken", mock.Anything, "somecode").Return(new(oauth2.Token).WithExtra(map[string]interface{}{
		"id_token": idToken,
	}), nil)

	mockStore := tokenstoremock.TokenStore{}
	mockStore.On("GenerateSessionToken", "some-user-id", "pureuser@purestorage.com").Return("session-token", nil)
	mockStore.On("StoreAPIToken", "_session_some-user-id", "session-token", "some-user-id").Return(fmt.Errorf("Some error"))
	mockStore.On("DeleteAPIToken", "_session_some-user-id").Return(nil)

	a := dexApp{
		oauth2config:  &mockConfig,
		apiTokenStore: &mockStore,
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("state", stateToken)
	req.Form.Set("code", "somecode")

	a.handleCallback(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleCallbackErrorStoringTokenAndErrorDeleting(t *testing.T) {
	tokenstore.HmacSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	stateToken, err := generateStateToken("http://192.168.99.100", tokenstore.HmacSecret, time.Now().Add(time.Minute*5).Unix())
	assert.NoError(t, err)
	assert.NotNil(t, stateToken)

	claims := jwt.MapClaims{
		"sub":   "some-user-id",
		"email": "pureuser@purestorage.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	idToken, err := token.SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("ExchangeCodeForToken", mock.Anything, "somecode").Return(new(oauth2.Token).WithExtra(map[string]interface{}{
		"id_token": idToken,
	}), nil)

	mockStore := tokenstoremock.TokenStore{}
	mockStore.On("GenerateSessionToken", "some-user-id", "pureuser@purestorage.com").Return("session-token", nil)
	mockStore.On("StoreAPIToken", "_session_some-user-id", "session-token", "some-user-id").Return(fmt.Errorf("Some error"))
	mockStore.On("DeleteAPIToken", "_session_some-user-id").Return(fmt.Errorf("Some error"))

	a := dexApp{
		oauth2config:  &mockConfig,
		apiTokenStore: &mockStore,
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("state", stateToken)
	req.Form.Set("code", "somecode")

	a.handleCallback(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleCallbackErrorStoringUser(t *testing.T) {
	tokenstore.HmacSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	stateToken, err := generateStateToken("http://192.168.99.100", tokenstore.HmacSecret, time.Now().Add(time.Minute*5).Unix())
	assert.NoError(t, err)
	assert.NotNil(t, stateToken)

	claims := jwt.MapClaims{
		"sub":   "some-user-id",
		"email": "pureuser@purestorage.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	idToken, err := token.SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("ExchangeCodeForToken", mock.Anything, "somecode").Return(new(oauth2.Token).WithExtra(map[string]interface{}{
		"id_token": idToken,
	}), nil)

	mockStore := tokenstoremock.TokenStore{}
	mockStore.On("GenerateSessionToken", "some-user-id", "pureuser@purestorage.com").Return("session-token", nil)
	mockStore.On("StoreAPIToken", "_session_some-user-id", "session-token", "some-user-id").Return(nil)
	mockStore.On("StoreUser", "some-user-id", mock.Anything).Return(fmt.Errorf("Some error"))

	a := dexApp{
		oauth2config:  &mockConfig,
		apiTokenStore: &mockStore,
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("state", stateToken)
	req.Form.Set("code", "somecode")

	a.handleCallback(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestHandleCallback(t *testing.T) {
	tokenstore.HmacSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	stateToken, err := generateStateToken("http://192.168.99.100", tokenstore.HmacSecret, time.Now().Add(time.Minute*5).Unix())
	assert.NoError(t, err)
	assert.NotNil(t, stateToken)

	claims := jwt.MapClaims{
		"sub":   "some-user-id",
		"email": "pureuser@purestorage.com",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	idToken, err := token.SignedString([]byte(tokenstore.HmacSecret))
	assert.NoError(t, err)

	mockConfig := oidcmock.OAuth2Config{}
	mockConfig.On("ExchangeCodeForToken", mock.Anything, "somecode").Return(new(oauth2.Token).WithExtra(map[string]interface{}{
		"id_token": idToken,
	}), nil)

	mockStore := tokenstoremock.TokenStore{}
	mockStore.On("GenerateSessionToken", "some-user-id", "pureuser@purestorage.com").Return("session-token", nil)
	mockStore.On("StoreAPIToken", "_session_some-user-id", "session-token", "some-user-id").Return(nil)
	mockStore.On("StoreUser", "some-user-id", mock.Anything).Return(nil)

	a := dexApp{
		oauth2config:  &mockConfig,
		apiTokenStore: &mockStore,
	}

	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", "", nil)
	assert.NoError(t, err)
	req.ParseForm()
	req.Form.Set("state", stateToken)
	req.Form.Set("code", "somecode")

	a.handleCallback(w, req)

	assert.Equal(t, http.StatusSeeOther, w.Code)

	assert.Contains(t, w.HeaderMap["Set-Cookie"], "pure1-unplugged-token=session-token; Path=/")
	assert.Contains(t, w.HeaderMap["Set-Cookie"], "pure1-unplugged-session=_session_some-user-id; Path=/")
	assert.Equal(t, "http://192.168.99.100", w.Header().Get("Location"))
}
