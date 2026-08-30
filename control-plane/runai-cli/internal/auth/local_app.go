package auth

import (
	"context"
	"fmt"

	"k8s.io/utils/ptr"

	api "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	clierrors "github.com/run-ai/runai/control-plane/runai-cli/internal/errors"
	"golang.org/x/oauth2"
)

const (
	clientCredentialsGrantType = api.TOKENREQUESTGRANTTYPEENUM_CLIENT_CREDENTIALS
)

type LocalApplication struct {
	name   string
	secret string
	client api.TokensAPI
}

func NewApplicationID(name, secret string, client api.TokensAPI) *LocalApplication {
	return &LocalApplication{
		name:   name,
		secret: secret,
		client: client,
	}
}

func (la *LocalApplication) Authenticate(ctx context.Context) (*oauth2.Token, error) {
	request := api.TokenRequest{
		ClientID:     ptr.To(la.name),
		ClientSecret: ptr.To(la.secret),
		GrantType:    clientCredentialsGrantType.Ptr(),
	}
	token, res, err := la.client.GrantToken(ctx).TokenRequest(request).Execute()

	if res != nil {
		defer res.Body.Close()
	}

	if err != nil {
		return nil, clierrors.HandleResponseError(res, err)
	}
	claims, err := GetClaims(token.GetAccessToken())
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return &oauth2.Token{
		AccessToken:  token.GetAccessToken(),
		RefreshToken: token.GetRefreshToken(),
		Expiry:       claims.ExpiresAt.Time,
	}, nil
}
