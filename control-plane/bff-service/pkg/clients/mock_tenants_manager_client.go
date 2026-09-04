package clients

import (
	"context"
	"net/http"

	tmOpenapi "github.com/dsx-ai-factory/run-ai/control-plane/tenants-manager/pkg/client/generated"
)

type MockTenantsManagerClient struct{}

var _ TenantsManagerClient = &MockTenantsManagerClient{}

func (m *MockTenantsManagerClient) GetAllLogos(_ context.Context, _ int32, _ int32) (*tmOpenapi.TenantLogosResponse, *http.Response, error) {
	return &tmOpenapi.TenantLogosResponse{}, nil, nil
}
