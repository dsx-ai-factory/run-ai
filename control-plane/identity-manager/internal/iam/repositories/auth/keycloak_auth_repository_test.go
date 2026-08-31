/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Nerzal/gocloak/v13"
	"github.com/golang-jwt/jwt/v5"
	authMocks "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/auth/mocks"
	kcMocks "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/gocloak/mocks"
	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	runaicloakMocks "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/runaicloak/mocks"
	"github.com/stretchr/testify/assert"
)

const (
	realm           = "realm_name"
	clientId        = "client"
	clientSecret    = "secret"
	// Fake JWTs for testing: synthetic header/payload, no real signature, fake issuer/subject.
	TokenWithOcpIdp = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk5OTk5OTk5OTksImlhdCI6MTYwMDAwMDAwMCwiaXNzIjoiaHR0cHM6Ly9mYWtlLWlzc3Vlci5leGFtcGxlLmNvbS9hdXRoL3JlYWxtcy9ydW5haSIsImF1ZCI6WyJydW5haS1hZG1pbi11aSIsInJ1bmFpIl0sInR5cCI6IklEIiwiYXpwIjoicnVuYWktYWRtaW4tdWkiLCJ0ZW5hbnRfaWQiOjEwMDEsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJ0ZXN0QHJ1bi5haSIsImVtYWlsIjoidGVzdEBydW4uYWkiLCJzdWIiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDEiLCJpZGVudGl0eV9wcm92aWRlciI6Im9wZW5zaGlmdC12NCIsInN1YmplY3RfdXVpZCI6IjAwMDAwMDAwLTAwMDAtMDAwMC0wMDAwLTAwMDAwMDAwMDAwMSJ9.mock"
	TokenWithoutIdp = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjk5OTk5OTk5OTksImlhdCI6MTYwMDAwMDAwMCwiaXNzIjoiaHR0cHM6Ly9mYWtlLWlzc3Vlci5leGFtcGxlLmNvbS9hdXRoL3JlYWxtcy9ydW5haSIsImF1ZCI6WyJydW5haS1hZG1pbi11aSIsInJ1bmFpIl0sInR5cCI6IklEIiwiYXpwIjoicnVuYWktYWRtaW4tdWkiLCJ0ZW5hbnRfaWQiOjEwMDEsImVtYWlsX3ZlcmlmaWVkIjp0cnVlLCJwcmVmZXJyZWRfdXNlcm5hbWUiOiJ0ZXN0QHJ1bi5haSIsImVtYWlsIjoidGVzdEBydW4uYWkiLCJzdWIiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDIifQ.mock"
)

func initTest(t *testing.T) (authMock *authMocks.Auth, goCloakMock *kcMocks.GoCloak, testSubject *KeycloakAuthRepository) {
	goCloak := kcMocks.NewGoCloak(t)
	auth := authMocks.NewAuth(t)
	runaicloakClient := runaicloakMocks.NewRunAICloakClient(t)
	testSubject = &KeycloakAuthRepository{
		auth:             auth,
		kcClient:         goCloak,
		runaicloakClient: runaicloakClient,
	}
	return auth, goCloak, testSubject
}

func TestRefreshToken_HappyFlow(t *testing.T) {
	_, goCloakMock, testSubject := initTest(t)
	tokenOptions := gocloak.TokenOptions{
		RefreshToken: gocloak.StringP("refreshToken"),
		GrantType:    gocloak.StringP("refresh_token"),
		ClientID:     gocloak.StringP("clientId"),
		ClientSecret: gocloak.StringP("clientSecret"),
	}
	goCloakMock.On("GetToken", context.Background(), realm, tokenOptions).Return(getJwt(), nil)

	tokens, err := testSubject.RefreshToken(context.Background(), realm, "refreshToken", "clientId", "clientSecret")
	assert.Nil(t, err)
	assert.Equal(t, tokens.IdToken, "IDToken")
	assert.Equal(t, tokens.RefreshToken, "refreshToken")
	assert.Equal(t, tokens.AccessToken, "AccessToken")
}

func TestRefreshToken_KcError(t *testing.T) {
	_, goCloakMock, testSubject := initTest(t)
	tokenOptions := gocloak.TokenOptions{
		RefreshToken: gocloak.StringP("refreshToken"),
		GrantType:    gocloak.StringP("refresh_token"),
		ClientID:     gocloak.StringP("clientId"),
		ClientSecret: gocloak.StringP("clientSecret"),
	}
	goCloakMock.On("GetToken", context.Background(), realm, tokenOptions).Return(nil, assert.AnError)

	tokens, err := testSubject.RefreshToken(context.Background(), realm, "refreshToken", "clientId", "clientSecret")
	assert.Nil(t, tokens)
	assert.NotNil(t, err)
}

func TestAuthenticateApp_HappyFlow(t *testing.T) {
	_, goCloakMock, testSubject := initTest(t)
	appId := "appId"
	appSecret := "appSecret"
	tokenOptions := gocloak.TokenOptions{
		ClientID:     &appId,
		ClientSecret: &appSecret,
		GrantType:    gocloak.StringP("client_credentials"),
		Scope:        gocloak.StringP("openid"),
		ResponseType: gocloak.StringP("id_token"),
	}
	goCloakMock.On("GetToken", context.Background(), realm, tokenOptions).Return(getJwt(), nil)

	tokens, err := testSubject.AuthenticateClient(context.Background(), realm, appId, appSecret)
	assert.Nil(t, err)
	assert.Equal(t, "IDToken", tokens.IdToken)
	assert.Equal(t, "AccessToken", tokens.AccessToken)
}

func TestAuthenticateApp_KeycloakError(t *testing.T) {
	_, goCloakMock, testSubject := initTest(t)
	appId := "appId"
	appSecret := "appSecret"
	tokenOptions := gocloak.TokenOptions{
		ClientID:     &appId,
		ClientSecret: &appSecret,
		GrantType:    gocloak.StringP("client_credentials"),
		Scope:        gocloak.StringP("openid"),
		ResponseType: gocloak.StringP("id_token"),
	}
	goCloakMock.On("GetToken", context.Background(), realm, tokenOptions).Return(nil, assert.AnError)

	tokens, err := testSubject.AuthenticateClient(context.Background(), realm, appId, appSecret)
	assert.Nil(t, tokens)
	assert.NotNil(t, err)
}

func TestExchangeCodeForToken_HappyFlow(t *testing.T) {
	_, goCloakMock, testSubject := initTest(t)
	tokenOptions := gocloak.TokenOptions{
		Code:         gocloak.StringP("code"),
		GrantType:    gocloak.StringP("authorization_code"),
		RedirectURI:  gocloak.StringP("redirectUri"),
		ClientID:     gocloak.StringP("clientId"),
		ClientSecret: gocloak.StringP("clientSecret"),
	}
	jwt := &gocloak.JWT{
		AccessToken:  TokenWithoutIdp,
		RefreshToken: TokenWithoutIdp,
		IDToken:      TokenWithoutIdp,
	}
	goCloakMock.On("GetToken", context.Background(), realm, tokenOptions).Return(jwt, nil)
	tokens, err := testSubject.ExchangeCodeForToken(context.Background(), realm, "redirectUri", "code", "clientId", "clientSecret", "")
	assert.Nil(t, err)
	assert.Equal(t, TokenWithoutIdp, tokens.IdToken)
	assert.Equal(t, TokenWithoutIdp, tokens.AccessToken)
	assert.Equal(t, TokenWithoutIdp, tokens.RefreshToken)
}

func TestExchangeCodeForToken_KeycloakError(t *testing.T) {
	_, goCloakMock, testSubject := initTest(t)
	tokenOptions := gocloak.TokenOptions{
		Code:         gocloak.StringP("code"),
		GrantType:    gocloak.StringP("authorization_code"),
		RedirectURI:  gocloak.StringP("redirectUri"),
		ClientID:     gocloak.StringP("clientId"),
		ClientSecret: gocloak.StringP("clientSecret"),
	}
	goCloakMock.On("GetToken", context.Background(), realm, tokenOptions).Return(nil, assert.AnError)
	tokens, err := testSubject.ExchangeCodeForToken(context.Background(), realm, "redirectUri", "code", "clientId", "clientSecret", "")
	assert.NotNil(t, err)
	assert.Nil(t, tokens)
}

func TestExchangeCodeForToken_WithPKCE_HappyFlow(t *testing.T) {
	authMock, _, testSubject := initTest(t)

	keycloakResponse := `{"access_token":"accessToken","refresh_token":"refreshToken","id_token":"idToken","expires_in":300}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		assert.NoError(t, r.ParseForm())
		assert.Equal(t, "authorization_code", r.FormValue("grant_type"))
		assert.Equal(t, "myVerifier", r.FormValue("code_verifier"))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(keycloakResponse))
	}))
	defer server.Close()

	authMock.On("GetHost").Return(server.URL)

	tokens, err := testSubject.ExchangeCodeForToken(context.Background(), realm, "redirectUri", "code", "clientId", "clientSecret", "myVerifier")
	assert.Nil(t, err)
	assert.Equal(t, "accessToken", tokens.AccessToken)
	assert.Equal(t, "refreshToken", tokens.RefreshToken)
	assert.Equal(t, "idToken", tokens.IdToken)
}

func TestExchangeCodeForToken_WithPKCE_KeycloakError(t *testing.T) {
	authMock, _, testSubject := initTest(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid_grant"}`))
	}))
	defer server.Close()

	authMock.On("GetHost").Return(server.URL)

	tokens, err := testSubject.ExchangeCodeForToken(context.Background(), realm, "redirectUri", "code", "clientId", "clientSecret", "myVerifier")
	assert.NotNil(t, err)
	assert.Nil(t, tokens)
}

func TestExchangeUsernameAndPasswordForToken_HappyFlow(t *testing.T) {
	_, goCloakMock, testSubject := initTest(t)
	tokenOptions := gocloak.TokenOptions{
		GrantType:    gocloak.StringP("password"),
		Username:     gocloak.StringP("username"),
		Password:     gocloak.StringP("password"),
		ClientID:     gocloak.StringP("clientId"),
		ClientSecret: gocloak.StringP("clientSecret"),
		Scope:        gocloak.StringP("openid email"),
	}
	jwt := &gocloak.JWT{
		AccessToken:  TokenWithoutIdp,
		RefreshToken: TokenWithoutIdp,
		IDToken:      TokenWithoutIdp,
	}
	goCloakMock.On("GetToken", context.Background(), realm, tokenOptions).Return(jwt, nil)
	tokens, err := testSubject.ExchangeUsernameAndPasswordForToken(context.Background(), realm, "username", "password", "clientId", "clientSecret")
	assert.Nil(t, err)
	assert.Equal(t, TokenWithoutIdp, tokens.IdToken)
	assert.Equal(t, TokenWithoutIdp, tokens.AccessToken)
	assert.Equal(t, TokenWithoutIdp, tokens.RefreshToken)
}

func TestExchangeUsernameAndPasswordForToken_KeycloakError(t *testing.T) {
	_, goCloakMock, testSubject := initTest(t)
	tokenOptions := gocloak.TokenOptions{
		GrantType:    gocloak.StringP("password"),
		Username:     gocloak.StringP("username"),
		Password:     gocloak.StringP("password"),
		ClientID:     gocloak.StringP("clientId"),
		ClientSecret: gocloak.StringP("clientSecret"),
		Scope:        gocloak.StringP("openid email"),
	}
	goCloakMock.On("GetToken", context.Background(), realm, tokenOptions).Return(nil, assert.AnError)
	tokens, err := testSubject.ExchangeUsernameAndPasswordForToken(context.Background(), realm, "username", "password", "clientId", "clientSecret")
	assert.NotNil(t, err)
	assert.Nil(t, tokens)
}

func TestGetUserInfo_HappyFlow(t *testing.T) {
	_, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	userInfo := make(map[string]interface{})

	accessToken := getMockTokens().AccessToken
	email := "test@run.ai"
	uid := "6babbb3c-cd7c-49d9-898a-a46d04aa8fcc"
	groups := []string{"group1", "group2", "group3"}
	userInfo["sub"] = email
	userInfo["subject_uuid"] = uid
	userInfo["groups"] = groups
	goCloakMock.On("GetRawUserInfo", ctx, accessToken, realm).Return(userInfo, nil)

	info, err := testSubject.GetUserInfo(ctx, realm, accessToken)
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, email, *info.Email)
	assert.Equal(t, uid, *info.Uuid)
	assert.Equal(t, groups, info.Groups)
}

func TestGetUserInfoMissMatchOnData_HappyFlow(t *testing.T) {
	_, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()
	userInfo := make(map[string]interface{})

	accessToken := getMockTokens().AccessToken
	email := "test@run.ai"
	uid := "6babbb3c-cd7c-49d9-898a-a46d04aa8fcc"
	groups := []string{"group1", "group2", "group3"}
	userInfo["extra"] = email
	userInfo["subject_uuid"] = uid
	userInfo["groups"] = groups
	goCloakMock.On("GetRawUserInfo", ctx, accessToken, realm).Return(userInfo, nil)

	info, err := testSubject.GetUserInfo(ctx, realm, accessToken)
	assert.Nil(t, err)
	assert.NotNil(t, info)
	assert.NotNil(t, email)
	assert.Equal(t, uid, *info.Uuid)
	assert.Equal(t, groups, info.Groups)
}

func TestGetUserInfo_errorFlow(t *testing.T) {
	_, goCloakMock, testSubject := initTest(t)
	ctx := context.Background()

	accessToken := getMockTokens().AccessToken
	goCloakMock.On("GetRawUserInfo", ctx, accessToken, realm).Return(nil, errors.New("any"))

	info, err := testSubject.GetUserInfo(ctx, realm, accessToken)
	assert.Nil(t, info)
	assert.NotNil(t, err)
}

func TestGetClaimsFromToken_HappyFlow(t *testing.T) {
	_, _, testSubject := initTest(t)
	tokenString, err := generateFakeToken()
	assert.NoError(t, err)
	claims, err := testSubject.GetClaimsFromToken(context.Background(), tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, claims["email"], "test@run.ai")
}

func TestGetClaimsFromToken_Error(t *testing.T) {
	_, _, testSubject := initTest(t)
	tokenString := "invalid"
	claims, err := testSubject.GetClaimsFromToken(context.Background(), tokenString)
	assert.Error(t, err)
	assert.Nil(t, claims)
}

func TestValidateServiceAccess_HappyFlow(t *testing.T) {
	authMock, _, testSubject := initTest(t)
	authMock.On("GetAdminToken", context.Background()).Return(nil, nil)
	err := testSubject.ValidateServiceAccess(context.Background())
	assert.Nil(t, err)
}

func TestValidateServiceAccess_FailKeycloakError(t *testing.T) {
	authMock, _, testSubject := initTest(t)
	authMock.On("GetAdminToken", context.Background()).Return(nil, assert.AnError)
	err := testSubject.ValidateServiceAccess(context.Background())
	assert.NotNil(t, err)
}

func getMockTokens() *models.Tokens {
	jwt := kcMocks.GetMockJWT()
	return &models.Tokens{
		AccessToken:  jwt.AccessToken,
		RefreshToken: jwt.RefreshToken,
		IdToken:      jwt.IDToken,
	}
}

func getJwt() *gocloak.JWT {
	return &gocloak.JWT{
		RefreshToken: "refreshToken",
		IDToken:      "IDToken",
		AccessToken:  "AccessToken",
	}
}

func generateFakeToken() (string, error) {
	secretKey := []byte("secret")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": "test@run.ai",
		"uid":   "6babbb3c-cd7c-49d9-898a-a46d04aa8fcc",
	})
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}
