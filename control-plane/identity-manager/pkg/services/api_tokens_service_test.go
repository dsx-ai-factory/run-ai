/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/utils/ptr"

	cacheMocks "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/cache/mocks"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/generated/pkg/api"
	iamError "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/errors"
	iamMock "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/mocks"
	iamModels "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/internal/iam/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/config"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/models"
	"github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils"
	mockUtils "github.com/dsx-ai-factory/run-ai/control-plane/identity-manager/pkg/utils/mocks"
)

func TestMain(m *testing.M) {
	config.InitConfig()
	m.Run()
}

func initTokenServiceTests(t *testing.T) (iamServicesContainerMock *iamMock.MockServicesContainer, clientCacheMock *cacheMocks.ClientCache, testSubject *TokensApiService, ctx context.Context) {
	mockIamServicesContainer := iamMock.NewMockServicesContainer()
	clientCacheMock = cacheMocks.NewClientCache(t)
	config := config.TenantRealm{
		DefaultClientId: "default",
	}
	testSubject = &TokensApiService{
		config:      &config,
		iamServices: *mockIamServicesContainer.ToServicesContainer(),
		clientCache: clientCacheMock,
	}
	ctx = context.WithValue(context.Background(), utils.AuthRealmKey, "realm")
	return mockIamServicesContainer, clientCacheMock, testSubject, ctx
}

func TestAppToken_HappyFlow(t *testing.T) {
	iamMock, _, testSubject, ctx := initTokenServiceTests(t)
	client := mockUtils.GetBasicClientArray()[0]
	tokens := getTokens()
	iamMock.MockAuthService.On("AuthenticateClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tokens, nil)
	req := api.AppTokenRequest{
		Name:   *client.ClientID,
		Secret: *client.Secret,
		Id:     "id",
	}
	resp, err := testSubject.AppToken(ctx, req)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, tokens.AccessToken, resp.Body.(api.AppTokenResponse).AccessToken)
	assert.Nil(t, err)
}

func TestAppToken_AuthenticateAppError(t *testing.T) {
	iamMock, _, testSubject, ctx := initTokenServiceTests(t)
	client := mockUtils.GetBasicClientArray()[0]
	iamMock.MockAuthService.On("AuthenticateClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, assert.AnError)
	req := api.AppTokenRequest{
		Name:   *client.ClientID,
		Secret: *client.Secret,
		Id:     "id",
	}
	resp, err := testSubject.AppToken(ctx, req)
	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.NotNil(t, err)
}

func TestAppToken_400InvalidClient(t *testing.T) {
	iamMock, _, testSubject, ctx := initTokenServiceTests(t)
	client := mockUtils.GetBasicClientArray()[0]
	iamMock.MockAuthService.On("AuthenticateClient", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, iamError.ErrInvalidInput)
	req := api.AppTokenRequest{
		Name:   *client.ClientID,
		Secret: *client.Secret,
		Id:     "id",
	}
	resp, err := testSubject.AppToken(ctx, req)
	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.NotNil(t, err)
}

func TestGetClientIDForRequest_NonEmptyClientId(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)
	client := mockUtils.GetBasicClientArray()[0]
	tokenRequest := api.TokenRequest{
		ClientID: *client.ClientID,
	}
	clientID := testSubject.getClientIDForRequest(tokenRequest)
	assert.Equal(t, mockUtils.APP_CLIENT_ID, clientID)
}

func TestGetClientIDForRequest_CliClient(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)
	tokenRequest := api.TokenRequest{
		ClientID:  "",
		GrantType: models.GrantTypeUserPassword,
	}
	clientID := testSubject.getClientIDForRequest(tokenRequest)
	assert.Equal(t, utils.CLI_CLIENT_ID, clientID)
}

func TestGetClientIDForRequest_CliClient2(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)
	tokenRequest := api.TokenRequest{
		ClientID: "",
	}
	clientID := testSubject.getClientIDForRequest(tokenRequest)
	assert.Equal(t, "default", clientID)
}

func TestGetTokens_ExchangeToken(t *testing.T) {
	iamMock, _, testSubject, _ := initTokenServiceTests(t)
	tokenRequest := api.TokenRequest{
		ClientID: "",
	}
	client := mockUtils.GetBasicClientArray()[0]
	iamMock.MockAuthService.On("ExchangeCodeForToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	testSubject.getTokens(context.Background(), models.GrantTypeExchangeToken, "realm", client, *client.Secret, tokenRequest)
	iamMock.MockAuthService.AssertNumberOfCalls(t, "ExchangeCodeForToken", 1)
}

func TestGetTokens_RefreshToken(t *testing.T) {
	iamMock, _, testSubject, _ := initTokenServiceTests(t)
	tokenRequest := api.TokenRequest{
		ClientID: "",
	}
	client := mockUtils.GetBasicClientArray()[0]
	iamMock.MockAuthService.On("RefreshToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	testSubject.getTokens(context.Background(), models.GrantTypeRefreshToken, "realm", client, *client.Secret, tokenRequest)
	iamMock.MockAuthService.AssertNumberOfCalls(t, "RefreshToken", 1)
}

func TestGetTokens_ExchangeUsernameAndPasswordForToken(t *testing.T) {
	iamMock, _, testSubject, _ := initTokenServiceTests(t)
	tokenRequest := api.TokenRequest{
		ClientID: "",
	}
	client := mockUtils.GetBasicClientArray()[0]
	iamMock.MockAuthService.On("ExchangeUsernameAndPasswordForToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
	testSubject.getTokens(context.Background(), models.GrantTypeUserPassword, "realm", client, *client.Secret, tokenRequest)
	iamMock.MockAuthService.AssertNumberOfCalls(t, "ExchangeUsernameAndPasswordForToken", 1)
}

func TestGetTokens_InvalidGrantType(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)
	tokenRequest := api.TokenRequest{
		ClientID: "",
	}
	client := mockUtils.GetBasicClientArray()[0]
	tokens, err := testSubject.getTokens(context.Background(), "INVALID", "realm", client, *client.Secret, tokenRequest)
	assert.Nil(t, tokens)
	assert.NotNil(t, err)
}

func TestGrantUserToken_HappyFlow(t *testing.T) {
	iamMock, clientCacheMock, testSubject, _ := initTokenServiceTests(t)
	claimsResponse := map[string]interface{}{
		"sub": "test@example.com",
	}
	iamMock.MockAuthService.On("GetClaimsFromToken", mock.Anything, mock.Anything).Return(claimsResponse, nil)
	client := mockUtils.GetBasicClientArray()[0]
	tokenRequest := api.TokenRequest{
		ClientID:     *client.ClientID,
		ClientSecret: *client.Secret,
		GrantType:    models.GrantTypeExchangeToken,
	}
	tokensResponse := getTokens()

	clientCacheMock.On("GetClientForRealm", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(client, nil)
	iamMock.MockAuthService.On("ExchangeCodeForToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(tokensResponse, nil)
	response, err := testSubject.grantUserToken(context.Background(), "realm", tokenRequest)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestGrantUserToken_NonAllowlistedClient_WrongSecret(t *testing.T) {
	iamMock, clientCacheMock, testSubject, _ := initTokenServiceTests(t)
	client := mockUtils.GetBasicClientArray()[0]
	tokenRequest := api.TokenRequest{
		ClientID:     *client.ClientID,
		ClientSecret: "wrong-secret",
		GrantType:    models.GrantTypeExchangeToken,
	}

	clientCacheMock.On("GetClientForRealm", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(client, nil)
	clientCacheMock.On("DeleteClientFromRealm", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	// The wrong secret is forwarded to Keycloak - identity-manager never fetches
	// or compares against the real secret for a non-allowlisted client.
	iamMock.MockAuthService.On("ExchangeCodeForToken", mock.Anything, mock.Anything, mock.Anything, mock.Anything, *client.ClientID, "wrong-secret", mock.Anything).Return(nil, iamError.ErrUnauthorized)
	response, err := testSubject.grantUserToken(context.Background(), "realm", tokenRequest)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.NotNil(t, err)
}

func TestGrantUserToken_NonAllowlistedClient_MissingSecret(t *testing.T) {
	_, clientCacheMock, testSubject, _ := initTokenServiceTests(t)
	client := mockUtils.GetBasicClientArray()[0]
	tokenRequest := api.TokenRequest{
		ClientID:  *client.ClientID,
		GrantType: models.GrantTypeExchangeToken,
	}

	clientCacheMock.On("GetClientForRealm", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(client, nil)
	clientCacheMock.On("DeleteClientFromRealm", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	response, err := testSubject.grantUserToken(context.Background(), "realm", tokenRequest)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.NotNil(t, err)
}

func TestGrantUserToken_EmptyClient(t *testing.T) {
	_, clientCacheMock, testSubject, _ := initTokenServiceTests(t)
	tokenRequest := api.TokenRequest{
		ClientID:  "",
		GrantType: models.GrantTypeExchangeToken,
	}
	clientCacheMock.On("GetClientForRealm", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("error"))
	response, err := testSubject.grantUserToken(context.Background(), "realm", tokenRequest)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.NotNil(t, err)
}

func TestGrantUserToken_ClientDoesNotExist(t *testing.T) {
	_, clientCacheMock, testSubject, _ := initTokenServiceTests(t)
	tokenRequest := api.TokenRequest{
		ClientID:  "",
		GrantType: models.GrantTypeExchangeToken,
	}
	clientCacheMock.On("GetClientForRealm", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("%s", ClientDoesntExistError))
	response, err := testSubject.grantUserToken(context.Background(), "realm", tokenRequest)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), ClientDoesntExistError)
}

func TestGrantUserToken_EmptyClientSecret(t *testing.T) {
	_, clientCacheMock, testSubject, _ := initTokenServiceTests(t)
	client := mockUtils.GetMultiClientsArray()[0]
	tokenRequest := api.TokenRequest{
		ClientID:  "",
		GrantType: models.GrantTypeExchangeToken,
	}
	clientCacheMock.On("GetClientForRealm", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(client, nil)
	clientCacheMock.On("DeleteClientFromRealm", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	response, err := testSubject.grantUserToken(context.Background(), "realm", tokenRequest)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.NotNil(t, err)
}

func TestGrantUserToken_InvalidGrantType(t *testing.T) {
	_, clientCacheMock, testSubject, _ := initTokenServiceTests(t)
	client := mockUtils.GetBasicClientArray()[0]
	tokenRequest := api.TokenRequest{
		ClientID:     *client.ClientID,
		ClientSecret: *client.Secret,
		GrantType:    "invalid",
	}
	clientCacheMock.On("GetClientForRealm", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(client, nil)
	clientCacheMock.On("DeleteClientFromRealm", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	response, err := testSubject.grantUserToken(context.Background(), "realm", tokenRequest)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotNil(t, err)
}

func TestHandleTokenResponse_HappyFlow(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)
	response, err := testSubject.handleTokenResponse(context.Background(), getTokens(), "", nil)
	assert.Equal(t, http.StatusOK, response.Code)
	assert.Nil(t, err)
}

func TestHandleTokenResponse_400Error(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)
	response, err := testSubject.handleTokenResponse(context.Background(), getTokens(), "", iamError.ErrInvalidInput)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.NotNil(t, err)
}

func TestHandleTokenResponse_401Error(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)
	response, err := testSubject.handleTokenResponse(context.Background(), getTokens(), "", iamError.ErrUnauthorized)
	assert.Equal(t, http.StatusUnauthorized, response.Code)
	assert.NotNil(t, err)
}

func TestHandleTokenResponse_GeneralError(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)
	response, err := testSubject.handleTokenResponse(context.Background(), getTokens(), "", iamError.ErrInternal)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
	assert.NotNil(t, err)
}

func TestGetCachedClientForRealm_BackendError(t *testing.T) {
	_, cacheMock, testSubject, ctx := initTokenServiceTests(t)
	clientID := "client-id-123"
	realm := "test-realm"

	expectedError := fmt.Errorf("backend service error")
	cacheMock.On("GetClientForRealm", mock.Anything, realm, clientID, mock.Anything).Return(nil, expectedError)

	client, err := testSubject.getCachedClientForRealm(ctx, realm, clientID, true)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Contains(t, err.Error(), expectedError.Error())
}

func getTokens() *iamModels.Tokens {
	return &iamModels.Tokens{
		AccessToken:  "accessToken",
		IdToken:      "idToken",
		RefreshToken: "refreshToken",
		ExpiresIn:    3600,
	}
}

func TestValidateTokenRequest(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)

	t.Run("client_credentials missing fields", func(t *testing.T) {
		err := testSubject.validateTokenRequest(api.TokenRequest{
			GrantType: models.GrantTypeAppToken,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), models.TokenRequestFieldClientID)

		err = testSubject.validateTokenRequest(api.TokenRequest{
			GrantType: models.GrantTypeAppToken,
			ClientID:  "client-id",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), models.TokenRequestFieldClientSecret)
	})
	t.Run("password missing fields", func(t *testing.T) {
		err := testSubject.validateTokenRequest(api.TokenRequest{
			GrantType: models.GrantTypeUserPassword,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), models.TokenRequestFieldUsername)

		err = testSubject.validateTokenRequest(api.TokenRequest{
			GrantType: models.GrantTypeUserPassword,
			Username:  "user",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), models.TokenRequestFieldPassword)
	})
	t.Run("refresh_token missing field", func(t *testing.T) {
		err := testSubject.validateTokenRequest(api.TokenRequest{
			GrantType: models.GrantTypeRefreshToken,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), models.TokenRequestFieldRefreshToken)
	})
	t.Run("exchange_token missing fields", func(t *testing.T) {
		err := testSubject.validateTokenRequest(api.TokenRequest{
			GrantType: models.GrantTypeExchangeToken,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), models.TokenRequestFieldCode)

		err = testSubject.validateTokenRequest(api.TokenRequest{
			GrantType: models.GrantTypeExchangeToken,
			Code:      "code",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), models.TokenRequestFieldRedirectUri)
	})
	t.Run("app_token (deprecated) missing fields", func(t *testing.T) {
		err := testSubject.validateTokenRequest(api.TokenRequest{
			GrantType: models.GrantTypeAppTokenDeprecated,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), models.TokenRequestFieldAppID)

		err = testSubject.validateTokenRequest(api.TokenRequest{
			GrantType: models.GrantTypeAppTokenDeprecated,
			AppID:     "appid",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), models.TokenRequestFieldAppSecret)
	})
	t.Run("all required fields present", func(t *testing.T) {
		cases := []api.TokenRequest{
			{GrantType: models.GrantTypeAppToken, ClientID: "id", ClientSecret: "secret"},
			{GrantType: models.GrantTypeUserPassword, Username: "u", Password: "p"},
			{GrantType: models.GrantTypeRefreshToken, RefreshToken: "r"},
			{GrantType: models.GrantTypeExchangeToken, Code: "c", RedirectUri: "uri"},
			{GrantType: models.GrantTypeAppTokenDeprecated, AppID: "id", AppSecret: "secret"},
		}
		for _, req := range cases {
			err := testSubject.validateTokenRequest(req)
			assert.NoError(t, err)
		}
	})
	t.Run("exchange_external_token missing external token", func(t *testing.T) {
		err := testSubject.validateTokenRequest(api.TokenRequest{
			GrantType: models.GrantTypeExternalTokenExchange,
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), models.TokenRequestFieldExternalToken)
	})

	t.Run("codeVerifier validation", func(t *testing.T) {
		validVerifier := strings.Repeat("a", 43)

		err := testSubject.validateTokenRequest(api.TokenRequest{
			GrantType:    models.GrantTypeAuthorizationCode,
			Code:         "code",
			RedirectUri:  "https://example.com",
			CodeVerifier: "",
		})
		assert.NoError(t, err, "empty codeVerifier should be accepted")

		err = testSubject.validateTokenRequest(api.TokenRequest{
			GrantType:    models.GrantTypeAuthorizationCode,
			Code:         "code",
			RedirectUri:  "https://example.com",
			CodeVerifier: validVerifier,
		})
		assert.NoError(t, err, "valid 43-char codeVerifier should be accepted")

		err = testSubject.validateTokenRequest(api.TokenRequest{
			GrantType:    models.GrantTypeAuthorizationCode,
			Code:         "code",
			RedirectUri:  "https://example.com",
			CodeVerifier: strings.Repeat("B", 128),
		})
		assert.NoError(t, err, "valid 128-char codeVerifier should be accepted")

		err = testSubject.validateTokenRequest(api.TokenRequest{
			GrantType:    models.GrantTypeAuthorizationCode,
			Code:         "code",
			RedirectUri:  "https://example.com",
			CodeVerifier: strings.Repeat("x", 42),
		})
		assert.Error(t, err, "42-char codeVerifier should be rejected (too short)")

		err = testSubject.validateTokenRequest(api.TokenRequest{
			GrantType:    models.GrantTypeAuthorizationCode,
			Code:         "code",
			RedirectUri:  "https://example.com",
			CodeVerifier: strings.Repeat("x", 129),
		})
		assert.Error(t, err, "129-char codeVerifier should be rejected (too long)")

		err = testSubject.validateTokenRequest(api.TokenRequest{
			GrantType:    models.GrantTypeAuthorizationCode,
			Code:         "code",
			RedirectUri:  "https://example.com",
			CodeVerifier: strings.Repeat("a", 42) + "!",
		})
		assert.Error(t, err, "codeVerifier with invalid character should be rejected")

		err = testSubject.validateTokenRequest(api.TokenRequest{
			GrantType:    models.GrantTypeUserPassword,
			Username:     "user",
			Password:     "pass",
			CodeVerifier: "",
		})
		assert.NoError(t, err, "password grant without codeVerifier should pass")

		err = testSubject.validateTokenRequest(api.TokenRequest{
			GrantType:    models.GrantTypeRefreshToken,
			RefreshToken: "token",
			CodeVerifier: "",
		})
		assert.NoError(t, err, "refresh_token grant without codeVerifier should pass")
	})
}

func TestValidateCodeVerifier(t *testing.T) {
	t.Run("empty is valid", func(t *testing.T) {
		assert.NoError(t, validateCodeVerifier(""))
	})
	t.Run("valid min length", func(t *testing.T) {
		assert.NoError(t, validateCodeVerifier(strings.Repeat("a", 43)))
	})
	t.Run("valid max length", func(t *testing.T) {
		assert.NoError(t, validateCodeVerifier(strings.Repeat("Z", 128)))
	})
	t.Run("valid with all allowed chars", func(t *testing.T) {
		assert.NoError(t, validateCodeVerifier("abcdefghijklmnopqrstuvwxyz0123456789._~-ABCDE"))
	})
	t.Run("too short", func(t *testing.T) {
		err := validateCodeVerifier(strings.Repeat("a", 42))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "between 43 and 128")
	})
	t.Run("too long", func(t *testing.T) {
		err := validateCodeVerifier(strings.Repeat("a", 129))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "between 43 and 128")
	})
	t.Run("invalid chars - space", func(t *testing.T) {
		err := validateCodeVerifier(strings.Repeat("a", 42) + " ")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid characters")
	})
	t.Run("invalid chars - special", func(t *testing.T) {
		err := validateCodeVerifier(strings.Repeat("a", 42) + "!")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid characters")
	})
	t.Run("invalid chars - slash", func(t *testing.T) {
		err := validateCodeVerifier(strings.Repeat("a", 42) + "/")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid characters")
	})
}

func TestGetClientForRealmNoCache_HappyFlow(t *testing.T) {
	iamMockForTestCache, _, testSubject, ctx := initTokenServiceTests(t)
	clientID := "client-id-123"
	realm := "test-realm"

	backendClient := &iamModels.Client{ClientID: &clientID, Secret: ptr.To("secret")}

	iamMockForTestCache.MockClientService.On("GetClientByClientId", mock.Anything, realm, clientID).Return(backendClient, nil)

	client, err := testSubject.getClientForRealmNoCache(ctx, realm, clientID, true)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, clientID, *client.ClientID)
	iamMockForTestCache.MockClientService.AssertExpectations(t)
	iamMockForTestCache.MockClientService.AssertCalled(t, "GetClientByClientId", mock.Anything, realm, clientID)
}

func TestGetClientForRealmNoCache_BackendError(t *testing.T) {
	iamMockForTestCache, _, testSubject, ctx := initTokenServiceTests(t)
	clientID := "client-id-123"
	realm := "test-realm"

	expectedError := fmt.Errorf("backend service error")
	iamMockForTestCache.MockClientService.On("GetClientByClientId", mock.Anything, realm, clientID).Return(nil, expectedError)

	client, err := testSubject.getClientForRealmNoCache(ctx, realm, clientID, true)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, expectedError, err)
}

func TestGetClientForRealmNoCache_ClientDoesNotExistError(t *testing.T) {
	iamMockForTestCache, _, testSubject, ctx := initTokenServiceTests(t)
	clientID := "client-id-123"
	realm := "test-realm"

	iamMockForTestCache.MockClientService.On("GetClientByClientId", mock.Anything, realm, clientID).Return(nil, nil)

	client, err := testSubject.getClientForRealmNoCache(ctx, realm, clientID, true)

	assert.Error(t, err)
	assert.Nil(t, client)
	assert.Equal(t, ClientDoesntExistError, err.Error())
}

func TestGetClientForRealmNoCache_FetchSecretFalse_SkipsSecretLookup(t *testing.T) {
	iamMockForTestCache, _, testSubject, ctx := initTokenServiceTests(t)
	clientID := "client-id-123"
	realm := "test-realm"

	backendClient := &iamModels.Client{ClientID: &clientID}
	iamMockForTestCache.MockClientService.On("GetClientByClientId", mock.Anything, realm, clientID).Return(backendClient, nil)

	client, err := testSubject.getClientForRealmNoCache(ctx, realm, clientID, false)

	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Nil(t, client.Secret)
	iamMockForTestCache.MockClientService.AssertNotCalled(t, "GetClientSecret", mock.Anything, mock.Anything, mock.Anything)
}

func TestGetClientForRealmNoCache_MissingClientID(t *testing.T) {
	iamMockForTestCache, _, testSubject, ctx := initTokenServiceTests(t)
	clientID := "client-id-123"
	realm := "test-realm"

	iamMockForTestCache.MockClientService.On("GetClientByClientId", mock.Anything, realm, clientID).Return(&iamModels.Client{ClientID: &clientID}, nil)

	client, err := testSubject.getClientForRealmNoCache(ctx, realm, clientID, true)

	assert.ErrorContains(t, err, "client client-id-123 in realm test-realm is missing required fields")
	assert.Nil(t, client)
}

func TestHandleTokenResponse_SetsCookieHeader(t *testing.T) {
	_, _, testSubject, ctx := initTokenServiceTests(t)

	tokens := &iamModels.Tokens{
		AccessToken:  "access-token",
		IdToken:      "id-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    900,
	}

	result, err := testSubject.handleTokenResponse(ctx, tokens, "", nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.Code)
	assert.NotNil(t, result.Headers)
	assert.Contains(t, result.Headers, "Set-Cookie")
	assert.Len(t, result.Headers["Set-Cookie"], 1)
	assert.Contains(t, result.Headers["Set-Cookie"][0], "refresh-token=")
	assert.Contains(t, result.Headers["Set-Cookie"][0], "HttpOnly")
	assert.Contains(t, result.Headers["Set-Cookie"][0], "Secure")
}

func TestHandleTokenResponse_OmitsRefreshTokenForUIClient(t *testing.T) {
	_, _, testSubject, ctx := initTokenServiceTests(t)

	tokens := &iamModels.Tokens{
		AccessToken:  "access-token",
		IdToken:      "id-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    900,
	}

	result, err := testSubject.handleTokenResponse(ctx, tokens, utils.RUNAI_ADMIN_UI_CLIENT_ID, nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.Code)
	resp, ok := result.Body.(api.TokenResponse)
	assert.True(t, ok)
	assert.Equal(t, "access-token", resp.AccessToken)
	assert.Empty(t, resp.RefreshToken)
}

func TestHandleTokenResponse_IncludesRefreshTokenForNonUIClient(t *testing.T) {
	_, _, testSubject, ctx := initTokenServiceTests(t)

	tokens := &iamModels.Tokens{
		AccessToken:  "access-token",
		IdToken:      "id-token",
		RefreshToken: "refresh-token",
		ExpiresIn:    900,
	}

	result, err := testSubject.handleTokenResponse(ctx, tokens, "some-other-client", nil)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.Code)
	resp, ok := result.Body.(api.TokenResponse)
	assert.True(t, ok)
	assert.Equal(t, "refresh-token", resp.RefreshToken)
}

func TestConvertToOAuth2Response_PreservesHeaders(t *testing.T) {
	_, _, testSubject, _ := initTokenServiceTests(t)

	internalResponse := api.TokenResponse{
		AccessToken:  "access",
		IdToken:      "id",
		RefreshToken: "refresh",
		ExpiresIn:    900,
	}

	headers := map[string][]string{
		"Set-Cookie": {"refresh-token=test; HttpOnly"},
	}

	result := api.ImplResponse{
		Code:    http.StatusOK,
		Headers: headers,
		Body:    internalResponse,
	}

	oauth2Result := testSubject.convertToOAuth2Response(result)

	assert.Equal(t, http.StatusOK, oauth2Result.Code)
	assert.NotNil(t, oauth2Result.Headers)
	assert.Equal(t, headers, oauth2Result.Headers)
}

func TestGrantToken_UsesRefreshTokenFromContext(t *testing.T) {
	iamMock, clientCacheMock, testSubject, ctx := initTokenServiceTests(t)

	ctx = context.WithValue(ctx, utils.RefreshTokenFromCookieKey, "cookie-refresh-token")

	client := mockUtils.GetBasicClientArray()[0]
	tokens := getTokens()

	clientCacheMock.On("GetClientForRealm", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(client, nil)
	iamMock.MockAuthService.On("RefreshToken", mock.Anything, mock.Anything, "cookie-refresh-token", mock.Anything, mock.Anything).
		Return(tokens, nil)
	iamMock.MockAuthService.On("GetClaimsFromToken", mock.Anything, mock.Anything).
		Return(map[string]interface{}{"sub": "test-subject"}, nil)

	tokenRequest := api.TokenRequest{
		GrantType:    models.GrantTypeRefreshToken,
		ClientID:     *client.ClientID,
		ClientSecret: *client.Secret,
		RefreshToken: "",
	}

	result, err := testSubject.GrantToken(ctx, tokenRequest)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, result.Code)
	iamMock.MockAuthService.AssertCalled(t, "RefreshToken", mock.Anything, mock.Anything, "cookie-refresh-token", mock.Anything, mock.Anything)
}
