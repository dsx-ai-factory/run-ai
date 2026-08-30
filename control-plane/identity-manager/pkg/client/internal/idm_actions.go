package internal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
	openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
)

func clientCredentialsLogin(ctx context.Context, client *openapi.APIClient, clientId string, clientSecret string) (*openapi.TokenResponse, error) {
	if clientId == "" || clientSecret == "" {
		return nil, errors.New("missing required field")
	}
	requestBody := openapi.TokenRequest{
		GrantType:    openapi.TOKENREQUESTGRANTTYPEENUM_CLIENT_CREDENTIALS.Ptr(),
		ClientID:     openapi.PtrString(clientId),
		ClientSecret: openapi.PtrString(clientSecret),
	}
	return tokenRequest(ctx, fmt.Sprintf("client login. client id: %s", clientId), client, requestBody)
}

func clientCredentialsLoginV2(ctx context.Context, client *openapi.APIClient, clientId string, clientSecret string) (*openapi.OAuth2TokenResponse, error) {
	if clientId == "" || clientSecret == "" {
		return nil, errors.New("missing required field clientId or clientSecret")
	}
	req := client.TokensAPI.GrantTokenV2(ctx).
		GrantType(string(openapi.TOKENREQUESTGRANTTYPEENUM_CLIENT_CREDENTIALS)).
		ClientId(clientId).
		ClientSecret(clientSecret)
	res, _, err := client.TokensAPI.GrantTokenV2Execute(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get token. request: client login. client id: %s. err: %v", clientId, err)
	}
	return res, nil
}

func (client *IdmClient) exchangeCodeForToken(ctx context.Context, exchangeCode string, clientId string, redirectUri string, codeVerifier string) (*openapi.TokenResponse, error) {
	requestBody := openapi.TokenRequest{
		GrantType:   openapi.TOKENREQUESTGRANTTYPEENUM_EXCHANGE_TOKEN.Ptr(),
		Code:        openapi.PtrString(exchangeCode),
		ClientID:    openapi.PtrString(clientId),
		RedirectUri: openapi.PtrString(redirectUri),
	}
	if codeVerifier != "" {
		requestBody.CodeVerifier = openapi.PtrString(codeVerifier)
	}
	return tokenRequest(ctx, "exchange code", client.ApiClient, requestBody)
}

func (client *IdmClient) userLogin(ctx context.Context, username string, password string, clientId string) (*openapi.TokenResponse, error) {
	requestBody := openapi.TokenRequest{
		GrantType: openapi.TOKENREQUESTGRANTTYPEENUM_PASSWORD.Ptr(),
		Username:  openapi.PtrString(username),
		Password:  openapi.PtrString(password),
		ClientID:  openapi.PtrString(clientId),
	}
	return tokenRequest(ctx, fmt.Sprintf("username and password. username: %s", username), client.ApiClient, requestBody)
}

func (client *IdmClient) refreshToken(ctx context.Context, token string, clientId string) (*openapi.TokenResponse, error) {
	requestBody := openapi.TokenRequest{
		GrantType:    openapi.TOKENREQUESTGRANTTYPEENUM_REFRESH_TOKEN.Ptr(),
		RefreshToken: openapi.PtrString(token),
		ClientID:     openapi.PtrString(clientId),
	}
	return tokenRequest(ctx, "refresh", client.ApiClient, requestBody)
}

func tokenRequest(ctx context.Context, requester string, client *openapi.APIClient, requestBody openapi.TokenRequest) (*openapi.TokenResponse, error) {
	req := client.TokensAPI.GrantToken(ctx)
	req = req.TokenRequest(requestBody)
	res, httpResponse, err := client.TokensAPI.GrantTokenExecute(req)
	if err != nil || httpResponse == nil || httpResponse.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get token. request: %s. StatusCode: %v, err: %v", requester, httpResponse, err)
	}
	return res, nil
}
