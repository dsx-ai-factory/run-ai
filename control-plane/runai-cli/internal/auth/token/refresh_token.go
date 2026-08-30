package token

import (
	"context"
	"fmt"
	"io"

	api "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/runai-cli/internal/auth"
	"k8s.io/utils/ptr"

	"golang.org/x/oauth2"
)

type RefresherClient struct {
	apiClient auth.TokenApi
}

func NewTokenRefresher(apiClient auth.TokenApi) *RefresherClient {
	return &RefresherClient{apiClient: apiClient}
}

func (r *RefresherClient) RefreshToken(ctx context.Context, refreshToken string) (*oauth2.Token, error) {
	var err error
	token, err := r.getTokenFromAPI(ctx, refreshToken)
	if err != nil {
		return nil, err
	}
	claims, err := auth.GetClaims(token.GetAccessToken())
	if err != nil {
		return nil, fmt.Errorf("failed to parse token claims: %w", err)
	}

	return &oauth2.Token{
		AccessToken:  token.GetAccessToken(),
		RefreshToken: token.GetRefreshToken(),
		Expiry:       claims.ExpiresAt.Time,
	}, nil
}

func (r *RefresherClient) getTokenFromAPI(ctx context.Context, refreshToken string) (*api.TokenResponse, error) {
	requestBody := r.buildTokenRequest(refreshToken)
	request := r.apiClient.GrantToken(ctx).TokenRequest(requestBody)
	token, response, err := request.Execute()
	if err != nil {
		if response != nil {
			defer response.Body.Close()
			data, _ := io.ReadAll(response.Body)
			return nil, fmt.Errorf("failed to refresh token: %s. %w", string(data), err)
		}
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}
	return token, nil
}

func (r *RefresherClient) buildTokenRequest(refreshToken string) api.TokenRequest {
	requestBody := api.TokenRequest{
		GrantType:    api.TOKENREQUESTGRANTTYPEENUM_REFRESH_TOKEN.Ptr(),
		RefreshToken: ptr.To(refreshToken),
		ClientID:     ptr.To(auth.CLIClientID),
	}
	return requestBody
}
