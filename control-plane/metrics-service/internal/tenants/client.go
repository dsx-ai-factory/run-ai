package tenants

import (
	"context"
	"fmt"
	"net/http"

	api "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/config"
)

type TenantsClient struct {
	url    string
	client *api.APIClient
}

func NewTenantsClient() *TenantsClient {
	cfg := config.Get[*TenantsConfiguration]()
	return &TenantsClient{
		url: cfg.ServiceUrl,
		client: api.NewAPIClient(&api.Configuration{
			Servers: api.ServerConfigurations{{URL: cfg.ServiceUrl}},
		}),
	}
}

func (tc *TenantsClient) GetExtendedMetricStore(ctx context.Context, tenantId int32) (*api.ExtendedMetricStore, error) {
	metricStore, res, err := tc.client.InternalAPI.GetTenantExtendedMetricStore(ctx, tenantId).Execute()
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tenant [%d] extended metric store, status: %v", tenantId, res.StatusCode)
	}
	return metricStore, nil
}

func (tc *TenantsClient) TenantExist(ctx context.Context, tenantId int32) (bool, error) {
	tenant, resp, err := tc.client.TenantsAPI.GetTenantByIdInternal(ctx, tenantId).Execute()
	if tenant != nil {
		return true, nil
	}

	if resp == nil || resp.StatusCode != http.StatusNotFound {
		return true, err
	}

	return false, nil
}
