package auth

import (
	"context"

	openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/config"
	"golang.org/x/oauth2"
)

type Authenticator interface {
	Authenticate(ctx context.Context) (*oauth2.Token, error)
}

type ApplicationAuthenticator struct {
	clientLogin  *openapi.APIClient
	clientName   string
	clientSecret string
}

func NewApplicationAuthenticator(authUrl string, clientName string, clientSecret string) (*ApplicationAuthenticator, error) {
	clientLogin, err := NewLoginClient(authUrl)
	if err != nil {
		return nil, err
	}
	return &ApplicationAuthenticator{
		clientLogin:  clientLogin,
		clientName:   clientName,
		clientSecret: clientSecret,
	}, nil
}

func (aa *ApplicationAuthenticator) Authenticate(ctx context.Context) (*oauth2.Token, error) {
	logger := config.GetInstance().GetLogger()
	logger.Debug(LoginWithApplicationCreds)
	applicationId := NewApplicationID(aa.clientName, aa.clientSecret, aa.clientLogin.TokensAPI)
	return applicationId.Authenticate(ctx)
}
