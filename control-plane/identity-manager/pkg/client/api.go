package identity_manager_client

import (
	"context"
	"net/http"

	openapi "github.com/run-ai/runai/control-plane/identity-manager/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/identity-manager/pkg/client/internal"
	client_utils "github.com/run-ai/runai/runai-application-packages/client-utils"
)

//go:generate mockery --name IIdm
type IIdm interface {
	GetInstallerServiceAccount(ctx context.Context) (*openapi.ServiceAccountPostResponse, error)
	GetInstallerServiceAccountForTenantId(ctx context.Context, tenantId string) (*openapi.ServiceAccountPostResponse, error)
	MintInstallerBootstrapToken(ctx context.Context, tenantId string) (string, error)
	GetInternalClientByFilter(ctx context.Context, filter string) ([]openapi.Client, error)
	AppendRedirectUriToInternalClient(ctx context.Context, client openapi.Client, uri string) error
	DeleteRealm(ctx context.Context, realmName string, isHardDelete bool) (*http.Response, error)
	CreateRealm(ctx context.Context, realmName string, tenantId string) (*openapi.RealmCreationResponse, *http.Response, error)
	GetOrCreateRealm(ctx context.Context, realmName string, tenantId string) (*openapi.RealmCreationResponse, *http.Response, error)
	CreateInternalUser(ctx context.Context, req openapi.InternalUserCreationRequest) (*openapi.UserPostResponse, *http.Response, error)
	DeleteInternalUserByEmail(ctx context.Context, realm string, email string) (*http.Response, error)
	CreateInternalClient(ctx context.Context, req openapi.ClientCreationRequest) (*openapi.CreateInternalClient201Response, *http.Response, error)
	DeleteClient(ctx context.Context, clientId string) (*http.Response, error)
	CreateClient(ctx context.Context, name string, clientType string, tenantId *string) (*string, error)
	CreateServiceAccountInternally(ctx context.Context, name string, realmName string) (*openapi.ServiceAccountPostResponse, error)
	IIdmExternalClient
}

//go:generate mockery --name IIdmExternalClient
type IIdmExternalClient interface {
	ClientCredentialsLogin(ctx context.Context, clientId string, clientSecret string) (*openapi.TokenResponse, error)
	ClientCredentialsLoginV2(ctx context.Context, clientId string, clientSecret string) (*openapi.OAuth2TokenResponse, error)
	TokenExchange(ctx context.Context, exchangeCode string, clientId string, redirectUri string, codeVerifier string) (*openapi.TokenResponse, error)
	RefreshToken(ctx context.Context, refreshToken string, clientId string) (*openapi.TokenResponse, error)
	UserLogin(ctx context.Context, username string, password string, clientId string) (*openapi.TokenResponse, error)
}

// NewIdmClient creates a new API client.  Requires a base URL.
func NewIdmClient(baseURL string) IIdm {
	baseURL = client_utils.GetInternalUrl(baseURL)
	config := openapi.NewConfiguration()
	config.Servers = openapi.ServerConfigurations{{URL: baseURL}}
	return &internal.IdmClient{ApiClient: openapi.NewAPIClient(config)}
}

func NewIdmExternalClient(baseURL string) IIdmExternalClient {
	baseURL = client_utils.GetExternalUrl(baseURL)
	return &internal.IdmClient{ApiClient: openapi.NewAPIClient(&openapi.Configuration{
		Servers: openapi.ServerConfigurations{{URL: baseURL}},
	})}
}
