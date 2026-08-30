package tests

import (
	"context"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/run-ai/runai/control-plane/identity-manager/internal/iam/models"
	realm_client "github.com/run-ai/runai/control-plane/identity-manager/pkg/realm/client"

	openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	serviceUtils "github.com/run-ai/runai/control-plane/identity-manager/pkg/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

const keycloakTokenEndpoint = "http://" + KeycloakDomain + ":8090/auth/realms/%s/protocol/openid-connect/token"

func TestGrantServiceAccountToken_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	serviceAccountId, err := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, nil)
	require.NoError(t, err)
	require.NotNil(t, serviceAccountId)
	defer clientCleanup(t, ctx, serviceAccountId)

	serviceAccount, err := IamServices.ClientService.GetClient(ctx, BaseRealmName, serviceAccountId)
	require.NoError(t, err)
	require.NotNil(t, serviceAccount)
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)
	clientCredentialsTokenRequest := generateClientCredentialsTokenRequest(*serviceAccount.ClientID, *serviceAccount.Secret)

	token, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(clientCredentialsTokenRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, token)

	parser := jwt.NewParser()
	parsedToken, _, err := parser.ParseUnverified(*token.AccessToken, &jwt.MapClaims{})
	require.NoError(t, err)
	require.NotNil(t, parsedToken)

	if claims, ok := parsedToken.Claims.(*jwt.MapClaims); ok {
		require.Contains(t, (*claims)["azp"], *serviceAccount.ClientID)
		require.Contains(t, (*claims)["client_type"], realm_client.ClientTypeTenantApp)
	}
}

func TestGrantAccessKeyToken_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	accessKeyId, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, accessKeyId)
	defer clientCleanup(t, ctx, accessKeyId)

	accessKey, err := IamServices.ClientService.GetClient(ctx, BaseRealmName, accessKeyId)
	require.NoError(t, err)
	require.NotNil(t, accessKey)

	clientCredentialsTokenRequest := generateClientCredentialsTokenRequest(*accessKey.ClientID, *accessKey.Secret)

	token, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(clientCredentialsTokenRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, token)

	parser := jwt.NewParser()
	parsedToken, _, err := parser.ParseUnverified(*token.AccessToken, &jwt.MapClaims{})
	require.NoError(t, err)
	require.NotNil(t, parsedToken)

	if claims, ok := parsedToken.Claims.(*jwt.MapClaims); ok {
		require.Contains(t, (*claims)["sub"], *accessKey.Attributes.CreatedBy)
		require.Contains(t, (*claims)["client_type"], realm_client.ClientTypeUserApp)
	}
}

func TestGrantAccessKeyToken_EmailClaim(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	knownCreatedBy := GenerateUsername()
	accessKeyId, err := TestStepsClient.CreateAccessKey(ctx, BaseRealmName, BaseTenantId, nil, &knownCreatedBy)
	require.NoError(t, err)
	require.NotNil(t, accessKeyId)
	defer clientCleanup(t, ctx, accessKeyId)

	accessKey, err := IamServices.ClientService.GetClient(ctx, BaseRealmName, accessKeyId)
	require.NoError(t, err)
	require.NotNil(t, accessKey)

	token, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(generateClientCredentialsTokenRequest(*accessKey.ClientID, *accessKey.Secret)).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	parser := jwt.NewParser()
	parsedToken, _, err := parser.ParseUnverified(*token.AccessToken, &jwt.MapClaims{})
	require.NoError(t, err)

	claims, ok := parsedToken.Claims.(*jwt.MapClaims)
	require.True(t, ok)
	require.Equal(t, knownCreatedBy, (*claims)["email"], "access key token email claim must equal the creator's ID")
}

func TestGrantServiceAccountToken_NoEmailVerifiedClaim(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	serviceAccountId, err := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, nil)
	require.NoError(t, err)
	require.NotNil(t, serviceAccountId)
	defer clientCleanup(t, ctx, serviceAccountId)

	serviceAccount, err := IamServices.ClientService.GetClient(ctx, BaseRealmName, serviceAccountId)
	require.NoError(t, err)
	require.NotNil(t, serviceAccount)

	token, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(generateClientCredentialsTokenRequest(*serviceAccount.ClientID, *serviceAccount.Secret)).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	parser := jwt.NewParser()
	parsedToken, _, err := parser.ParseUnverified(*token.AccessToken, &jwt.MapClaims{})
	require.NoError(t, err)

	claims, ok := parsedToken.Claims.(*jwt.MapClaims)
	require.True(t, ok)
	_, hasEmail := (*claims)["email_verified"]
	require.False(t, hasEmail, "service account token must not contain email_verified claim")
}

func TestGrantServiceAccountToken_InvalidInput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	serviceAccountId, err := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, nil)
	require.NoError(t, err)
	require.NotNil(t, serviceAccountId)
	defer clientCleanup(t, ctx, serviceAccountId)

	serviceAccount, err := IamServices.ClientService.GetClient(ctx, BaseRealmName, serviceAccountId)
	require.NoError(t, err)
	require.NotNil(t, serviceAccount)

	clientCredentialsTokenRequest1 := generateClientCredentialsTokenRequest("invalid-input", *serviceAccount.Secret)

	token1, resp1, err1 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(clientCredentialsTokenRequest1).Execute()
	require.NotNil(t, err1)
	require.Equal(t, http.StatusUnauthorized, resp1.StatusCode)
	require.Nil(t, token1)

	clientCredentialsTokenRequest2 := generateClientCredentialsTokenRequest(*serviceAccount.ClientID, "invalid-input")

	token2, resp2, err2 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(clientCredentialsTokenRequest2).Execute()
	require.NotNil(t, err2)
	require.Equal(t, http.StatusUnauthorized, resp2.StatusCode)
	require.Nil(t, token2)
}

func TestGrantServiceAccountToken_InvalidGrantType(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	serviceAccountId, err := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, nil)
	require.NoError(t, err)
	require.NotNil(t, serviceAccountId)
	defer clientCleanup(t, ctx, serviceAccountId)

	serviceAccount, err := IamServices.ClientService.GetClient(ctx, BaseRealmName, serviceAccountId)
	require.NoError(t, err)
	require.NotNil(t, serviceAccount)

	clientCredentialsTokenRequest := openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_PASSWORD),
		ClientID:     serviceAccount.ClientID,
		ClientSecret: serviceAccount.Secret,
	}

	token, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(clientCredentialsTokenRequest).Execute()
	require.NotNil(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	require.Nil(t, token)
}

func TestGrantUserToken_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	userID, username, password, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	require.NotNil(t, userID)
	defer userCleanup(t, newTestContext(), userID)

	userTokenRequest := getUserTokenRequest(username, password)

	token, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(userTokenRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, token)

	parser := jwt.NewParser()
	parsedToken, _, err := parser.ParseUnverified(*token.AccessToken, &jwt.MapClaims{})
	require.NoError(t, err)
	require.NotNil(t, parsedToken)

	claims, ok := parsedToken.Claims.(*jwt.MapClaims)
	require.True(t, ok)
	require.Contains(t, (*claims)["azp"], serviceUtils.CLI_CLIENT_ID)
	require.Equal(t, true, (*claims)["email_verified"], "user token must contain email_verified: true")
}

func TestGrantUserToken_InvalidInput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	userID, usename, password, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	require.NotNil(t, userID)
	defer userCleanup(t, newTestContext(), userID)

	userTokenRequest1 := getUserTokenRequest("invalid-input", password)
	token1, resp1, err1 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(userTokenRequest1).Execute()
	require.NotNil(t, err1)
	require.Equal(t, http.StatusUnauthorized, resp1.StatusCode)
	require.Nil(t, token1)

	userTokenRequest2 := getUserTokenRequest(usename, "invalid-input")
	token2, resp2, err2 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(userTokenRequest2).Execute()
	require.NotNil(t, err2)
	require.Equal(t, http.StatusUnauthorized, resp2.StatusCode)
	require.Nil(t, token2)
}

func TestGrantRefreshToken_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	userID, username, password, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	require.NotNil(t, userID)
	defer userCleanup(t, newTestContext(), userID)

	userTokenRequest := getUserTokenRequest(username, password)
	token, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(userTokenRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, token)

	refreshTokenRequest := openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_REFRESH_TOKEN),
		RefreshToken: token.RefreshToken,
		ClientID:     ptr.To(serviceUtils.CLI_CLIENT_ID),
	}
	token1, resp1, err1 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(refreshTokenRequest).Execute()
	require.NoError(t, err1)
	require.Equal(t, http.StatusOK, resp1.StatusCode)
	require.NotNil(t, token1)
}

func TestGrantRefreshToken_ValidRefreshTokenFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	userID, username, password, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	require.NotNil(t, userID)
	defer userCleanup(t, newTestContext(), userID)

	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	userTokenRequest := getUserTokenRequest(username, password)
	initialToken, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(userTokenRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, initialToken)
	require.NotNil(t, initialToken.AccessToken)
	require.NotNil(t, initialToken.RefreshToken)
	require.NotNil(t, initialToken.IdToken)

	refreshTokenRequest := openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_REFRESH_TOKEN),
		RefreshToken: initialToken.RefreshToken,
		ClientID:     ptr.To(serviceUtils.CLI_CLIENT_ID),
	}
	newToken, resp1, err1 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(refreshTokenRequest).Execute()

	require.NoError(t, err1)
	require.Equal(t, http.StatusOK, resp1.StatusCode)
	require.NotNil(t, newToken)

	require.NotNil(t, newToken.AccessToken)
	require.NotNil(t, newToken.RefreshToken)
	require.NotNil(t, newToken.IdToken)
	require.NotEmpty(t, *newToken.AccessToken)
	require.NotEmpty(t, *newToken.RefreshToken)
	require.NotEmpty(t, *newToken.IdToken)

	require.NotEqual(t, *initialToken.AccessToken, *newToken.AccessToken)
	require.NotEqual(t, *initialToken.RefreshToken, *newToken.RefreshToken)
	require.NotEqual(t, *initialToken.IdToken, *newToken.IdToken)

	authenticatedCtx := TestStepsClient.SetContextValuesWithUser(context.Background(), BaseTenantId, username)
	authenticatedCtx = context.WithValue(authenticatedCtx, ctxKey("Authorization"), "Bearer "+*newToken.AccessToken)

	users, usersResp, usersErr := OpenapiClient.UsersAPI.GetUsers(authenticatedCtx).Execute()
	require.NoError(t, usersErr)
	require.Equal(t, http.StatusOK, usersResp.StatusCode)
	require.NotNil(t, users)
}

func TestGrantRefreshToken_InvalidInput(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	userID, username, password, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	require.NotNil(t, userID)
	defer userCleanup(t, newTestContext(), userID)

	userTokenRequest := getUserTokenRequest(username, password)
	token, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(userTokenRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, token)

	refreshTokenRequest := openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_REFRESH_TOKEN),
		RefreshToken: ptr.To("invalid-input"),
		ClientID:     ptr.To(serviceUtils.CLI_CLIENT_ID),
	}
	token1, resp1, err1 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(refreshTokenRequest).Execute()
	require.NotNil(t, err1)
	require.Equal(t, http.StatusUnauthorized, resp1.StatusCode)
	require.Nil(t, token1)
}

func TestGrantRefreshToken_RandomString(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	randomString := GenerateRandomString("random-refresh-token-")

	refreshTokenRequest := openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_REFRESH_TOKEN),
		RefreshToken: ptr.To(randomString),
		ClientID:     ptr.To(serviceUtils.CLI_CLIENT_ID),
	}

	token, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(refreshTokenRequest).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	require.Nil(t, token)
}

func TestGrantServiceAccountToken_BadRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	invalidClientCredentialsTokenRequest1 := openapi.TokenRequest{
		GrantType: ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_CLIENT_CREDENTIALS),
		Password:  ptr.To("password"),
	}
	token1, resp1, err1 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(invalidClientCredentialsTokenRequest1).Execute()
	require.NotNil(t, err1)
	require.Equal(t, http.StatusBadRequest, resp1.StatusCode)
	require.Nil(t, token1)

	invalidclientCredentialsTokenRequest2 := openapi.TokenRequest{
		GrantType: ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_CLIENT_CREDENTIALS),
		ClientID:  ptr.To("clientID"),
	}
	token2, resp2, err2 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(invalidclientCredentialsTokenRequest2).Execute()
	require.NotNil(t, err2)
	require.Equal(t, http.StatusBadRequest, resp2.StatusCode)
	require.Nil(t, token2)

	invalidclientCredentialsTokenRequest3 := openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_CLIENT_CREDENTIALS),
		ClientSecret: ptr.To("clientSecret"),
	}
	token3, resp3, err3 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(invalidclientCredentialsTokenRequest3).Execute()
	require.NotNil(t, err3)
	require.Equal(t, http.StatusBadRequest, resp3.StatusCode)
	require.Nil(t, token3)
}

func TestGrantToken_InvalidCredentials(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	serviceAccountId, err := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, nil)
	require.NoError(t, err)
	require.NotNil(t, serviceAccountId)
	defer clientCleanup(t, ctx, serviceAccountId)

	serviceAccount, err := IamServices.ClientService.GetClient(ctx, BaseRealmName, serviceAccountId)
	require.NoError(t, err)
	require.NotNil(t, serviceAccount)

	tokenRequest := openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_CLIENT_CREDENTIALS),
		ClientID:     serviceAccount.ClientID,
		ClientSecret: ptr.To("wrong-secret"),
	}

	_, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(tokenRequest).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGrantUserToken_BadRequest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	invalidUserTokenRequest1 := openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_PASSWORD),
		ClientSecret: ptr.To("clientSecret"),
	}
	token1, resp1, err1 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(invalidUserTokenRequest1).Execute()
	require.NotNil(t, err1)
	require.Equal(t, http.StatusBadRequest, resp1.StatusCode)
	require.Nil(t, token1)

	invalidUserTokenRequest2 := openapi.TokenRequest{
		GrantType: ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_PASSWORD),
		Username:  ptr.To("Username"),
	}
	token2, resp2, err2 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(invalidUserTokenRequest2).Execute()
	require.NotNil(t, err2)
	require.Equal(t, http.StatusBadRequest, resp2.StatusCode)
	require.Nil(t, token2)

	invalidUserTokenRequest3 := openapi.TokenRequest{
		GrantType: ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_PASSWORD),
		Password:  ptr.To("Password"),
	}
	token3, resp3, err3 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(invalidUserTokenRequest3).Execute()
	require.NotNil(t, err3)
	require.Equal(t, http.StatusBadRequest, resp3.StatusCode)
	require.Nil(t, token3)
}

func TestGrantRefreshToken_BadReqeust(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	refreshTokenRequest := openapi.TokenRequest{
		GrantType: ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_REFRESH_TOKEN),
		Username:  ptr.To("Username"),
	}
	token1, resp1, err1 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(refreshTokenRequest).Execute()
	require.NotNil(t, err1)
	require.Equal(t, http.StatusBadRequest, resp1.StatusCode)
	require.Nil(t, token1)
}

func TestGrantExchangeToken_BadReqeust(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	invalidExchangeTokenRequest1 := openapi.TokenRequest{
		GrantType: ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_EXCHANGE_TOKEN),
		Password:  ptr.To("password"),
	}
	token1, resp1, err1 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(invalidExchangeTokenRequest1).Execute()
	require.NotNil(t, err1)
	require.Equal(t, http.StatusBadRequest, resp1.StatusCode)
	require.Nil(t, token1)

	invalidExchangeTokenRequest2 := openapi.TokenRequest{
		GrantType:   ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_EXCHANGE_TOKEN),
		RedirectUri: ptr.To("RedirectUri"),
	}
	token2, resp2, err2 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(invalidExchangeTokenRequest2).Execute()
	require.NotNil(t, err2)
	require.Equal(t, http.StatusBadRequest, resp2.StatusCode)
	require.Nil(t, token2)

	invalidExchangeTokenRequest3 := openapi.TokenRequest{
		GrantType: ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_EXCHANGE_TOKEN),
		Code:      ptr.To("Code"),
	}
	token3, resp3, err3 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(invalidExchangeTokenRequest3).Execute()
	require.NotNil(t, err3)
	require.Equal(t, http.StatusBadRequest, resp3.StatusCode)
	require.Nil(t, token3)
}

func TestGrantExchangeToken_InvalidRedirectUri(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	fakeHost := fmt.Sprintf("%s.fake.run.ai", BaseRealmName)

	cases := []struct {
		name        string
		redirectUri string
	}{
		{
			name:        "open redirect to different host",
			redirectUri: "https://evil.com/login/callback",
		},
		{
			name:        "http instead of https",
			redirectUri: fmt.Sprintf("http://%s/login/callback", fakeHost),
		},
		{
			name:        "disallowed path",
			redirectUri: fmt.Sprintf("https://%s/admin", fakeHost),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := openapi.TokenRequest{
				GrantType:   ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_EXCHANGE_TOKEN),
				Code:        ptr.To("dummy-code"),
				RedirectUri: ptr.To(tc.redirectUri),
			}
			token, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(req).Execute()
			require.Error(t, err)
			require.Equal(t, http.StatusBadRequest, resp.StatusCode)
			require.Nil(t, token)
		})
	}
}

func TestGrantExternalTokenExchange_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	externalToken := "valid-external-token"
	tokenRequest := openapi.TokenRequest{
		GrantType:     ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_EXTERNAL_TOKEN_EXCHANGE),
		ExternalToken: ptr.To(externalToken),
	}
	token, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(tokenRequest).Execute()

	_ = token
	_ = resp
	_ = err
}

func TestGrantExternalTokenExchange_FailureCases(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	type testCase struct {
		name       string
		request    openapi.TokenRequest
		expectCode int
	}

	tests := []testCase{
		{
			name:       "Missing externalToken",
			request:    openapi.TokenRequest{GrantType: ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_EXTERNAL_TOKEN_EXCHANGE)},
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "No IdP configured for exchange",
			request:    openapi.TokenRequest{GrantType: ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_EXTERNAL_TOKEN_EXCHANGE), ExternalToken: ptr.To("valid-but-no-idp")},
			expectCode: http.StatusMethodNotAllowed,
		},
		{
			name:       "Unsupported grant type",
			request:    openapi.TokenRequest{GrantType: nil, ExternalToken: ptr.To("token")},
			expectCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(tc.request).Execute()
			require.Error(t, err)
			if resp != nil {
				require.Equal(t, tc.expectCode, resp.StatusCode)
			}
		})
	}
}

func TestGrantTokenV2_ClientCredentials_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	serviceAccountId, err := TestStepsClient.CreateServiceAccount(ctx, BaseRealmName, BaseTenantId, nil)
	require.NoError(t, err)
	require.NotNil(t, serviceAccountId)
	defer clientCleanup(t, ctx, serviceAccountId)

	serviceAccount, err := IamServices.ClientService.GetClient(ctx, BaseRealmName, serviceAccountId)
	require.NoError(t, err)
	require.NotNil(t, serviceAccount)

	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	oauth2Token, resp, err := OpenapiClient.TokensAPI.GrantTokenV2(ctx).
		GrantType("client_credentials").
		ClientId(*serviceAccount.ClientID).
		ClientSecret(*serviceAccount.Secret).
		Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, oauth2Token)

	require.Equal(t, "Bearer", oauth2Token.TokenType)
	require.NotEmpty(t, oauth2Token.AccessToken)
	require.NotZero(t, oauth2Token.ExpiresIn)

	parser := jwt.NewParser()
	parsedToken, _, err := parser.ParseUnverified(oauth2Token.AccessToken, &jwt.MapClaims{})
	require.NoError(t, err)
	require.NotNil(t, parsedToken)
}

func TestGrantTokenV2_Password_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	userID, username, password, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	require.NotNil(t, userID)
	defer userCleanup(t, newTestContext(), userID)

	oauth2Token, resp, err := OpenapiClient.TokensAPI.GrantTokenV2(ctx).
		GrantType("password").
		Username(username).
		Password(password).
		ClientId(serviceUtils.CLI_CLIENT_ID).
		Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, oauth2Token)

	require.Equal(t, "Bearer", oauth2Token.TokenType)
	require.NotEmpty(t, oauth2Token.AccessToken)
	require.NotEmpty(t, oauth2Token.RefreshToken)
	require.NotEmpty(t, oauth2Token.IdToken)
	require.NotZero(t, oauth2Token.ExpiresIn)
}

func TestGrantTokenV2_RefreshToken_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	userID, username, password, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	require.NotNil(t, userID)
	defer userCleanup(t, newTestContext(), userID)

	initialToken, resp, err := OpenapiClient.TokensAPI.GrantTokenV2(ctx).
		GrantType("password").
		Username(username).
		Password(password).
		ClientId(serviceUtils.CLI_CLIENT_ID).
		Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)

	refreshedToken, resp2, err2 := OpenapiClient.TokensAPI.GrantTokenV2(ctx).
		GrantType("refresh_token").
		RefreshToken(*initialToken.RefreshToken).
		ClientId(serviceUtils.CLI_CLIENT_ID).
		Execute()

	require.NoError(t, err2)
	require.Equal(t, http.StatusOK, resp2.StatusCode)
	require.NotNil(t, refreshedToken)
	require.Equal(t, "Bearer", refreshedToken.TokenType)
	require.NotEmpty(t, refreshedToken.AccessToken)
}

func TestGrantTokenV2_InvalidCredentials_ReturnsOAuth2Error(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	oauth2Error, resp, err := OpenapiClient.TokensAPI.GrantTokenV2(ctx).
		GrantType("client_credentials").
		ClientId("invalid-client").
		ClientSecret("wrong-secret").
		Execute()

	require.Error(t, err)
	require.Nil(t, oauth2Error)
	require.NotNil(t, resp)
}

func TestGetOpenIDConfiguration_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	config, resp, err := OpenapiClient.OpenIDConnectAPI.GetOpenidConfiguration(ctx).Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, config)

	require.Contains(t, config.TokenEndpoint, "/api/v2/token")
	require.NotEmpty(t, config.Issuer)
	require.NotEmpty(t, config.JwksUri)
	require.NotEmpty(t, config.AuthorizationEndpoint)

	require.Contains(t, config.GrantTypesSupported, "client_credentials")
	require.Contains(t, config.GrantTypesSupported, "password")
	require.Contains(t, config.GrantTypesSupported, "refresh_token")
	require.Contains(t, config.GrantTypesSupported, "authorization_code")

	require.Contains(t, config.ResponseTypesSupported, "code")
	require.Contains(t, config.SubjectTypesSupported, "public")
	require.Contains(t, config.IdTokenSigningAlgValuesSupported, "RS256")
}

func TestExternalTokenExchange_UserFlow(t *testing.T) {
	ctx := TestStepsClient.SetContextValues(context.Background(), BaseTenantId)
	client, clientId, _, _, cleanup := setupExternalTokenExchangeTest(t, ctx, "external-token-user")
	defer cleanup()

	userID, username, password, err := createTestUser(ctx, MasterRealmName)
	require.NoError(t, err)
	defer userCleanup(t, newTestContext(), userID)

	userAccessToken, err := exchangeUserToken(clientId, *client.Secret, username, password)
	require.NoError(t, err)

	exchangedUserToken, resp, err := exchangeExternalToken(ctx, userAccessToken)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, exchangedUserToken)
	exchangedUser, err := TestStepsClient.iamClient.UserService.GetUserByEmail(ctx, BaseRealmName, username)
	require.NoError(t, err)
	require.Equal(t, username, *exchangedUser.Email)
}

func TestExternalTokenExchange_ApplicationFlow(t *testing.T) {
	ctx := TestStepsClient.SetContextValues(context.Background(), BaseTenantId)
	client, clientId, _, _, cleanup := setupExternalTokenExchangeTest(t, ctx, "external-token-app")
	defer cleanup()

	tokenURL := fmt.Sprintf(keycloakTokenEndpoint, MasterRealmName)
	clientToken, err := GetKeycloakToken(tokenURL, clientId, *client.Secret, "", "", "client_credentials")
	require.NoError(t, err)
	require.NotEmpty(t, clientToken)

	exchangedAppToken, resp, err := exchangeExternalToken(ctx, clientToken)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, exchangedAppToken)
}

func TestGrantTokenV2_ExternalTokenExchange_UserFlow(t *testing.T) {
	ctx := TestStepsClient.SetContextValues(context.Background(), BaseTenantId)
	client, clientId, _, _, cleanup := setupExternalTokenExchangeTest(t, ctx, "external-token-v2-user")
	defer cleanup()

	userID, username, password, err := createTestUser(ctx, MasterRealmName)
	require.NoError(t, err)
	defer userCleanup(t, newTestContext(), userID)

	userAccessToken, err := exchangeUserToken(clientId, *client.Secret, username, password)
	require.NoError(t, err)

	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)
	oauth2Token, resp, err := OpenapiClient.TokensAPI.GrantTokenV2(ctx).
		GrantType("external_token_exchange").
		ExternalToken(userAccessToken).
		Execute()

	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, oauth2Token)
	require.Equal(t, "Bearer", oauth2Token.TokenType)
	require.NotEmpty(t, oauth2Token.AccessToken)
}

func TestGrantTokenV2_ExternalTokenExchange_FailureCases(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	type testCase struct {
		name       string
		grantType  string
		extToken   string
		expectCode int
	}

	tests := []testCase{
		{
			name:       "Missing external_token",
			grantType:  "external_token_exchange",
			extToken:   "",
			expectCode: http.StatusBadRequest,
		},
		{
			name:       "No IdP configured for exchange",
			grantType:  "external_token_exchange",
			extToken:   "valid-but-no-idp",
			expectCode: http.StatusMethodNotAllowed,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)
			req := OpenapiClient.TokensAPI.GrantTokenV2(ctx).GrantType(tc.grantType)
			if tc.extToken != "" {
				req = req.ExternalToken(tc.extToken)
			}
			_, resp, err := req.Execute()
			require.Error(t, err)
			require.Equal(t, tc.expectCode, resp.StatusCode)
		})
	}
}

func TestGrantToken_ClientNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	tokenRequest := openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_REFRESH_TOKEN),
		ClientID:     ptr.To("non-existent-client"),
		ClientSecret: ptr.To("secret"),
	}

	_, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(tokenRequest).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGrantToken_ExpiredRefreshToken(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)

	userID, username, password, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	require.NotNil(t, userID)
	defer userCleanup(t, newTestContext(), userID)

	userTokenRequest := getUserTokenRequest(username, password)
	token, resp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(userTokenRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.NotNil(t, token)

	refreshTokenRequest := openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_REFRESH_TOKEN),
		ClientID:     ptr.To("cli"),
		RefreshToken: token.RefreshToken,
	}

	token1, resp1, err1 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(refreshTokenRequest).Execute()
	require.NoError(t, err1)
	require.Equal(t, http.StatusOK, resp1.StatusCode)
	require.NotNil(t, token1)
	require.NotEqual(t, token.RefreshToken, token1.RefreshToken)

	token2, resp2, err2 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(refreshTokenRequest).Execute()
	require.NoError(t, err2)
	require.Equal(t, http.StatusOK, resp2.StatusCode)
	require.NotNil(t, token2)
	require.NotEqual(t, token1.RefreshToken, token2.RefreshToken)

	_, resp3, err3 := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(refreshTokenRequest).Execute()
	require.Error(t, err3)
	require.Equal(t, http.StatusUnauthorized, resp3.StatusCode)
	if openapiErr, ok := err.(*openapi.GenericOpenAPIError); ok {
		errMsg := string(openapiErr.Body())
		require.Contains(t, errMsg, "Invalid refresh token")
	}
}

func exchangeExternalToken(ctx context.Context, externalToken string) (interface{}, *http.Response, error) {
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)
	return OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(openapi.TokenRequest{
		GrantType:     ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_EXTERNAL_TOKEN_EXCHANGE),
		ExternalToken: &externalToken,
	}).Execute()
}

func setupExternalTokenExchangeTest(t *testing.T, ctx context.Context, testName string) (*models.Client, string, string, *openapi.Mappers, func()) {
	_, clientId, _, idpAlias, mappers := newRandomIdpTestData(ctx, testName)
	client, err := createAndConfigureClient(ctx, clientId)
	require.NoError(t, err)
	createOidcIdp(t, ctx, idpAlias, clientId, *client.Secret, mappers)
	cleanup := func() {
		deleteIDP(t, ctx, idpAlias)
	}
	return client, clientId, idpAlias, mappers, cleanup
}

func createAndConfigureClient(ctx context.Context, clientId string) (*models.Client, error) {
	if _, err := TestStepsClient.CreateServiceAccount(ctx, MasterRealmName, BaseTenantId, &clientId); err != nil {
		return nil, err
	}
	client, err := TestStepsClient.iamClient.ClientService.GetClientByClientId(ctx, "master", clientId)
	if err != nil {
		return nil, err
	}
	client.TenantId = ptr.To(BaseTenantId)
	realm_client.ApplyClientProfile(client, realm_client.ClientTypeTokenExchange)
	client.ClientConfig.DirectAccessGrantsEnabled = ptr.To(true)
	client.ClientConfig.ServiceAccountsEnabled = ptr.To(true)
	if err = IamServices.ClientService.UpdateClient(ctx, MasterRealmName, client); err != nil {
		return nil, err
	}
	secret, err := IamServices.ClientService.GetClientSecret(ctx, MasterRealmName, *client.ID)
	if err != nil {
		return nil, err
	}
	client.Secret = &secret
	return client, nil
}

func createTestUser(ctx context.Context, realm string) (string, string, string, error) {
	userID, username, password, err := TestStepsClient.CreateUser(ctx, realm, nil)
	return userID, username, password, err
}

func exchangeUserToken(clientId, clientSecret, username, password string) (string, error) {
	tokenURL := fmt.Sprintf(keycloakTokenEndpoint, MasterRealmName)
	return GetKeycloakToken(tokenURL, clientId, clientSecret, username, password, "password")
}

func generateClientCredentialsTokenRequest(clientID, clientSecret string) openapi.TokenRequest {
	return openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_CLIENT_CREDENTIALS),
		ClientID:     &clientID,
		ClientSecret: &clientSecret,
	}
}

func getUserTokenRequest(username, password string) openapi.TokenRequest {
	return openapi.TokenRequest{
		GrantType: ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_PASSWORD),
		Username:  ptr.To(username),
		Password:  ptr.To(password),
		ClientID:  ptr.To(serviceUtils.CLI_CLIENT_ID),
	}
}

func TestPKCE_LoginAndCodeExchange_HappyFlow(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	userID, username, password, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	require.NotNil(t, userID)
	defer userCleanup(t, newTestContext(), userID)

	codeVerifier := generatePKCECodeVerifier()
	codeChallenge := generatePKCECodeChallengeS256(codeVerifier)

	// Call identity-manager's login API with PKCE params and verify the redirect URL
	fakeHost := fmt.Sprintf("%s.fake.run.ai", BaseRealmName)
	loginURL := fmt.Sprintf("%s/api/v1/login?code_challenge=%s&code_challenge_method=S256",
		hostUrl, url.QueryEscape(codeChallenge))
	loginReq, err := http.NewRequest("GET", loginURL, nil)
	require.NoError(t, err)
	loginReq.Header.Set("User-Agent", "runai-cli/1.0")
	loginReq.Header.Set(FakeHostHeader, fakeHost)

	noRedirectClient := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	loginResp, err := noRedirectClient.Do(loginReq)
	require.NoError(t, err)
	defer loginResp.Body.Close()
	require.Equal(t, http.StatusTemporaryRedirect, loginResp.StatusCode)

	locationURL := loginResp.Header.Get("Location")
	require.NotEmpty(t, locationURL, "Location header should contain the redirect URL")
	parsedLocation, err := url.Parse(locationURL)
	require.NoError(t, err)
	require.Equal(t, codeChallenge, parsedLocation.Query().Get("code_challenge"))
	require.Equal(t, "S256", parsedLocation.Query().Get("code_challenge_method"))
	require.Equal(t, serviceUtils.CLI_CLIENT_ID, parsedLocation.Query().Get("client_id"))

	// Simulate Keycloak browser login to obtain an authorization code
	keycloakAuthorizeURL := buildKeycloakAuthorizeURL(BaseRealmName, serviceUtils.CLI_CLIENT_ID, "http://localhost:8080", codeChallenge)
	authCode := simulateKeycloakBrowserLogin(t, keycloakAuthorizeURL, username, password)
	require.NotEmpty(t, authCode)

	// Exchange auth code + code_verifier via identity-manager's token endpoint
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)
	tokenRequest := openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_EXCHANGE_TOKEN),
		Code:         ptr.To(authCode),
		RedirectUri:  ptr.To("http://localhost:8080"),
		CodeVerifier: ptr.To(codeVerifier),
	}
	token, tokenResp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(tokenRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, tokenResp.StatusCode)
	require.NotNil(t, token)
	require.NotNil(t, token.AccessToken)
	require.NotEmpty(t, *token.AccessToken)
}

func TestPKCE_CodeExchange_UIClient_OmitsRefreshTokenInBody(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	userID, username, password, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	require.NotNil(t, userID)
	defer userCleanup(t, newTestContext(), userID)

	codeVerifier := generatePKCECodeVerifier()
	codeChallenge := generatePKCECodeChallengeS256(codeVerifier)

	keycloakAuthorizeURL := buildKeycloakAuthorizeURL(BaseRealmName, serviceUtils.RUNAI_ADMIN_UI_CLIENT_ID, "http://localhost:8080", codeChallenge)
	authCode := simulateKeycloakBrowserLogin(t, keycloakAuthorizeURL, username, password)
	require.NotEmpty(t, authCode)

	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)
	tokenRequest := openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_EXCHANGE_TOKEN),
		ClientID:     ptr.To(serviceUtils.RUNAI_ADMIN_UI_CLIENT_ID),
		Code:         ptr.To(authCode),
		RedirectUri:  ptr.To("http://localhost:8080"),
		CodeVerifier: ptr.To(codeVerifier),
	}
	token, tokenResp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(tokenRequest).Execute()
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, tokenResp.StatusCode)
	require.NotNil(t, token)
	require.NotNil(t, token.AccessToken)
	require.NotEmpty(t, *token.AccessToken)

	require.Nil(t, token.RefreshToken, "UI client should not receive the refresh token in the response body")

	setCookies := tokenResp.Header.Values("Set-Cookie")
	require.NotEmpty(t, setCookies, "refresh token cookie should be set for UI client")
	var hasRefreshCookie bool
	for _, c := range setCookies {
		if strings.Contains(c, "refresh-token=") {
			hasRefreshCookie = true
			break
		}
	}
	require.True(t, hasRefreshCookie, "Set-Cookie header should contain refresh-token cookie")
}

func TestPKCE_CodeExchange_WrongVerifier(t *testing.T) {
	t.Parallel()
	ctx := newTestContext()

	userID, username, password, err := TestStepsClient.CreateUser(ctx, BaseRealmName, nil)
	require.NoError(t, err)
	require.NotNil(t, userID)
	defer userCleanup(t, newTestContext(), userID)

	codeVerifier := generatePKCECodeVerifier()
	codeChallenge := generatePKCECodeChallengeS256(codeVerifier)

	keycloakAuthorizeURL := buildKeycloakAuthorizeURL(BaseRealmName, serviceUtils.CLI_CLIENT_ID, "http://localhost:8080", codeChallenge)
	authCode := simulateKeycloakBrowserLogin(t, keycloakAuthorizeURL, username, password)
	require.NotEmpty(t, authCode)

	// Exchange with a WRONG code_verifier — Keycloak should reject this
	wrongVerifier := generatePKCECodeVerifier()
	ctx = TestStepsClient.SetFakeHostHeader(ctx, OpenapiClient, BaseRealmName)
	tokenRequest := openapi.TokenRequest{
		GrantType:    ptr.To(openapi.TOKENREQUESTGRANTTYPEENUM_EXCHANGE_TOKEN),
		Code:         ptr.To(authCode),
		RedirectUri:  ptr.To("http://localhost:8080"),
		CodeVerifier: ptr.To(wrongVerifier),
	}
	token, tokenResp, err := OpenapiClient.TokensAPI.GrantToken(ctx).TokenRequest(tokenRequest).Execute()
	require.Error(t, err)
	require.Equal(t, http.StatusBadRequest, tokenResp.StatusCode)
	require.Nil(t, token)
}

// generatePKCECodeVerifier creates a cryptographically random PKCE code_verifier (43 chars, base64url).
func generatePKCECodeVerifier() string {
	b := make([]byte, 32)
	if _, err := crand.Read(b); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// generatePKCECodeChallengeS256 computes the S256 code_challenge from a code_verifier.
func generatePKCECodeChallengeS256(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

func buildKeycloakAuthorizeURL(realm, clientID, redirectURI, codeChallenge string) string {
	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", clientID)
	params.Set("scope", "openid email")
	params.Set("redirect_uri", redirectURI)
	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")
	return fmt.Sprintf("http://%s:8090/auth/realms/%s/protocol/openid-connect/auth?%s",
		KeycloakDomain, realm, params.Encode())
}

// simulateKeycloakBrowserLogin performs the Keycloak browser-based login flow
// and returns the authorization code from the callback redirect.
func simulateKeycloakBrowserLogin(t *testing.T, authorizeURL, username, password string) string {
	t.Helper()

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)

	dialer := &net.Dialer{
		Timeout:   defaultTimeout,
		KeepAlive: defaultKeepAlive,
	}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, _ := net.SplitHostPort(addr)
			if host == KeycloakDomain {
				if isDocker() {
					return dialer.DialContext(ctx, network, net.JoinHostPort(dockerKeycloakHost, dockerKeycloakPort))
				}
				return dialer.DialContext(ctx, network, net.JoinHostPort(localhostIP, port))
			}
			return dialer.DialContext(ctx, network, addr)
		},
	}
	client := &http.Client{
		Transport: transport,
		Jar:       jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(authorizeURL)
	require.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	require.NoError(t, err)

	for resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusTemporaryRedirect {
		location := resp.Header.Get("Location")
		require.NotEmpty(t, location)
		if !strings.HasPrefix(location, "http") {
			parsed, _ := url.Parse(authorizeURL)
			location = fmt.Sprintf("%s://%s%s", parsed.Scheme, parsed.Host, location)
		}
		resp, err = client.Get(location)
		require.NoError(t, err)
		body, err = io.ReadAll(resp.Body)
		resp.Body.Close()
		require.NoError(t, err)
	}

	require.Equal(t, http.StatusOK, resp.StatusCode,
		"expected Keycloak login page, got %d: %s", resp.StatusCode, string(body))

	formActionRegex := regexp.MustCompile(`action="([^"]+)"`)
	matches := formActionRegex.FindSubmatch(body)
	require.NotEmpty(t, matches, "could not find login form action in Keycloak HTML response")
	formAction := html.UnescapeString(string(matches[1]))

	resp, err = client.PostForm(formAction, url.Values{
		"username": {username},
		"password": {password},
	})
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusFound, resp.StatusCode, "expected redirect after Keycloak login")
	callbackURL := resp.Header.Get("Location")
	require.NotEmpty(t, callbackURL)

	parsedCallback, err := url.Parse(callbackURL)
	require.NoError(t, err)
	code := parsedCallback.Query().Get("code")
	require.NotEmpty(t, code, "authorization code not found in callback URL: %s", callbackURL)

	return code
}
