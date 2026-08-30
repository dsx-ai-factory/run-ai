package internal

import (
	"context"
	"fmt"
	"net/http"

	api "github.com/run-ai/runai/control-plane/tenants-manager/pkg/client/generated"
	"github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type TenantsClienter interface {
	GetTenants(ctx context.Context) ([]api.Tenant, error)
	GetMetricStore(ctx context.Context, tenantId int32) (*api.ExtendedMetricStore, error)
}

type TenantsConfiguration struct {
	ServiceUrl string `mapstructure:"tenant_manager_url" validate:"required,url"`
}

func (tc *TenantsConfiguration) SetDefaults() {
	tc.ServiceUrl = "http://runai-backend-tenants-manager:8080"
}

type TenantsClient struct {
	client *api.APIClient
}

func NewTenantsClient() *TenantsClient {
	cfg := config.Get[*TenantsConfiguration]()
	return &TenantsClient{
		client: api.NewAPIClient(&api.Configuration{
			Servers: api.ServerConfigurations{{URL: cfg.ServiceUrl}},
		}),
	}
}
func (tc *TenantsClient) GetTenants(ctx context.Context) ([]api.Tenant, error) {
	tenants, res, err := tc.client.InternalAPI.ManagementGetTenantsInternal(ctx).Execute()
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get tenants from tenant service, res: %v, err: %v", res, err)
		return nil, err
	}

	return tenants, err
}

func (tc *TenantsClient) GetMetricStore(ctx context.Context, tenantId int32) (*api.ExtendedMetricStore, error) {
	metricStore, err := tc.getMetricStoreFromTenantService(ctx, tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("failed to get metric store for tenant %v from tenant service", tenantId)
		return nil, err
	}

	return metricStore, nil
}

func (tc *TenantsClient) getMetricStoreFromTenantService(ctx context.Context, tenantId int32) (*api.ExtendedMetricStore, error) {
	metricStore, res, err := tc.client.InternalAPI.GetTenantExtendedMetricStore(ctx, tenantId).Execute()
	if res == nil {
		return nil, fmt.Errorf("failed to get tenant [%d] connection data, err: %v", tenantId, err)
	}

	if res.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get tenant [%d] connection data, status: %v", tenantId, res.StatusCode)
	}

	return metricStore, nil
}
