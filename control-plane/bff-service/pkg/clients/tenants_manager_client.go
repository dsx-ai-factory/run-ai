package clients

import (
	"context"
	"net/http"

	tmclient "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client"
	tmOpenapi "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
)

type TenantsManagerClient interface {
	GetAllLogos(ctx context.Context, offset int32, pageSize int32) (*tmOpenapi.TenantLogosResponse, *http.Response, error)
}

type DefaultTenantsManagerClient struct {
	client tmclient.ITenantsManager
}

func (c *DefaultTenantsManagerClient) GetAllLogos(ctx context.Context, offset int32, pageSize int32) (*tmOpenapi.TenantLogosResponse, *http.Response, error) {
	return c.client.GetAllLogos(ctx, offset, pageSize)
}

var mockTMClient TenantsManagerClient

func SetMockTenantsManagerClient(mock TenantsManagerClient) {
	mockTMClient = mock
}

func NewTenantsManagerClient(baseURL string) TenantsManagerClient {
	if mockTMClient != nil {
		return mockTMClient
	}
	return &DefaultTenantsManagerClient{client: tmclient.NewTenantsManagerClient(baseURL)}
}
