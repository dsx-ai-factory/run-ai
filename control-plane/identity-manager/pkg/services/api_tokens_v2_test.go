/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/generated/pkg/api"
	iamModels "github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"
)

func TestGrantTokenV2_ClientCredentials_HappyFlow(t *testing.T) {
	iamMock, _, testSubject, ctx := initTokenServiceTests(t)
	tokens := getTokens()
	iamMock.MockAuthService.On("AuthenticateClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tokens, nil)

	result, err := testSubject.GrantTokenV2(
		ctx,
		"client_credentials",
		"test-client",
		"test-secret",
		"", "", "", "", "", "", "",
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.Code)

	oauth2Response, ok := result.Body.(api.OAuth2TokenResponse)
	assert.True(t, ok)
	assert.Equal(t, tokens.AccessToken, oauth2Response.AccessToken)
	assert.Equal(t, "Bearer", oauth2Response.TokenType)
	assert.NotZero(t, oauth2Response.ExpiresIn)
}

func TestGrantTokenV2_Password_HappyFlow(t *testing.T) {
	iamMock, clientCacheMock, testSubject, ctx := initTokenServiceTests(t)
	client := getBasicClient()
	tokens := getTokens()

	clientCacheMock.On("GetClientForRealm", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(client, nil)
	iamMock.MockAuthService.On("ExchangeUsernameAndPasswordForToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tokens, nil)
	iamMock.MockAuthService.On("GetClaimsFromToken", mock.Anything, mock.Anything).Return(map[string]interface{}{"sub": "test@run.ai"}, nil)

	result, err := testSubject.GrantTokenV2(
		ctx,
		"password",
		"",
		"",
		"test@run.ai",
		"password123",
		"", "", "", "", "",
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.Code)

	oauth2Response, ok := result.Body.(api.OAuth2TokenResponse)
	assert.True(t, ok)
	assert.Equal(t, tokens.AccessToken, oauth2Response.AccessToken)
	assert.Equal(t, "Bearer", oauth2Response.TokenType)
}

func TestGrantTokenV2_RefreshToken_HappyFlow(t *testing.T) {
	iamMock, clientCacheMock, testSubject, ctx := initTokenServiceTests(t)
	client := getBasicClient()
	tokens := getTokens()
	claimsResponse := map[string]interface{}{
		"sub": "test@run.ai",
	}

	clientCacheMock.On("GetClientForRealm", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(client, nil)
	iamMock.MockAuthService.On("RefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tokens, nil)
	iamMock.MockAuthService.On("GetClaimsFromToken", mock.Anything, mock.Anything).Return(claimsResponse, nil)

	result, err := testSubject.GrantTokenV2(
		ctx,
		"refresh_token",
		utils.CLI_CLIENT_ID,
		"",
		"", "", "",
		"valid-refresh-token",
		"", "", "",
	)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.Code)

	oauth2Response, ok := result.Body.(api.OAuth2TokenResponse)
	assert.True(t, ok)
	assert.Equal(t, "Bearer", oauth2Response.TokenType)
	assert.Equal(t, tokens.RefreshToken, oauth2Response.RefreshToken)
}

func TestGrantTokenV2_InvalidGrant_ReturnsOAuth2Error(t *testing.T) {
	iamMock, _, testSubject, ctx := initTokenServiceTests(t)
	iamMock.MockAuthService.On("AuthenticateClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("keycloak error: invalid_grant: Invalid credentials"))

	result, err := testSubject.GrantTokenV2(
		ctx,
		"client_credentials",
		"invalid-client",
		"wrong-secret",
		"", "", "", "", "", "", "",
	)

	assert.NotNil(t, err)

	oauth2Error, ok := result.Body.(api.OAuth2Error)
	assert.True(t, ok, "Response should be OAuth2Error")
	assert.NotEmpty(t, oauth2Error.Error)
	assert.NotEmpty(t, oauth2Error.ErrorDescription)
}

func TestGrantTokenV2_UnauthorizedClient_ReturnsOAuth2Error(t *testing.T) {
	_, clientCacheMock, testSubject, ctx := initTokenServiceTests(t)

	clientCacheMock.On("GetClientForRealm", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("client error"))

	result, err := testSubject.GrantTokenV2(
		ctx,
		"password",
		"",
		"",
		"test@run.ai",
		"password",
		"", "", "", "", "",
	)

	assert.NotNil(t, err)

	oauth2Error, ok := result.Body.(api.OAuth2Error)
	assert.True(t, ok, "Response should be OAuth2Error")
	assert.NotEmpty(t, oauth2Error.Error)
	assert.NotEmpty(t, oauth2Error.ErrorDescription)
}

func TestConvertToOAuth2Response_Success(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)

	internalResponse := api.TokenResponse{
		AccessToken:  "access_token_value",
		RefreshToken: "refresh_token_value",
		IdToken:      "id_token_value",
		ExpiresIn:    3600,
	}
	result := api.Response(http.StatusOK, internalResponse)

	oauth2Result := testSubject.convertToOAuth2Response(result)

	assert.Equal(t, http.StatusOK, oauth2Result.Code)
	oauth2Response, ok := oauth2Result.Body.(api.OAuth2TokenResponse)
	assert.True(t, ok)
	assert.Equal(t, "access_token_value", oauth2Response.AccessToken)
	assert.Equal(t, "Bearer", oauth2Response.TokenType)
	assert.Equal(t, "refresh_token_value", oauth2Response.RefreshToken)
	assert.Equal(t, "id_token_value", oauth2Response.IdToken)
	assert.Equal(t, int32(3600), oauth2Response.ExpiresIn)
}

func TestConvertToOAuth2Error_InvalidGrant(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)

	result := api.Response(http.StatusUnauthorized, nil)
	err := fmt.Errorf("invalid_grant: credentials are wrong")

	oauth2Result, resultErr := testSubject.convertToOAuth2Error(result, err)

	assert.NotNil(t, resultErr)
	assert.Equal(t, http.StatusUnauthorized, oauth2Result.Code)
	oauth2Error, ok := oauth2Result.Body.(api.OAuth2Error)
	assert.True(t, ok)
	assert.Equal(t, "invalid_grant", oauth2Error.Error)
	assert.Contains(t, oauth2Error.ErrorDescription, "invalid_grant")
}

func TestConvertToOAuth2Error_InvalidClient(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)

	result := api.Response(http.StatusBadRequest, nil)
	err := fmt.Errorf("invalid_client: client not found")

	oauth2Result, resultErr := testSubject.convertToOAuth2Error(result, err)

	assert.NotNil(t, resultErr)
	oauth2Error, ok := oauth2Result.Body.(api.OAuth2Error)
	assert.True(t, ok)
	assert.Equal(t, "invalid_client", oauth2Error.Error)
}

func TestConvertToOAuth2Error_UnsupportedGrantType(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)

	result := api.Response(http.StatusBadRequest, nil)
	err := fmt.Errorf("unsupported_grant_type")

	oauth2Result, resultErr := testSubject.convertToOAuth2Error(result, err)

	assert.NotNil(t, resultErr)
	oauth2Error, ok := oauth2Result.Body.(api.OAuth2Error)
	assert.True(t, ok)
	assert.Equal(t, "unsupported_grant_type", oauth2Error.Error)
}

func getBasicClient() *iamModels.Client {
	clientID := "test-client"
	secret := "test-secret"
	return &iamModels.Client{
		ClientID: &clientID,
		Secret:   &secret,
	}
}
