package clients

import (
	"context"
	"net/http"

	proxyapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/types"
)

type TenantRoutes struct {
	tenantRoutesClient *config_client.TenantRouteAPIService
}

func NewTenantRoutes() *TenantRoutes {
	return &TenantRoutes{tenantRoutesClient: GetConfigClient().TenantRouteAPI}
}

func (tr *TenantRoutes) GetRoute(ctx context.Context, tenantId string) (*types.BasicRouteConfig, error) {
	tenantRoutes, httpResponse, err := tr.tenantRoutesClient.GetRoutes(ctx, tenantId).Execute()
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
			return nil, nil
		}

		return nil, proxyapi.NewErrorFromResponse(httpResponse, err)
	}

	if len(tenantRoutes) == 0 {
		return nil, nil
	}

	route, err := parseRoute(tenantRoutes[0], tenantId)
	if err != nil {
		return nil, err
	}
	return route, nil
}

func (tr *TenantRoutes) DeleteRoute(ctx context.Context, tenantId string, routeId string) error {
	_, httpResponse, err := tr.tenantRoutesClient.DeleteRoute(ctx, tenantId, routeId).Execute()
	if err == nil {
		return nil
	}

	return proxyapi.NewErrorFromResponse(httpResponse, err)
}

func (tr *TenantRoutes) CreateOrUpdateRoute(ctx context.Context, route *types.BasicRouteConfig) error {
	cfg, err := DecodeToMap(route)
	if err != nil {
		return err
	}

	_, httpResponse, err := tr.tenantRoutesClient.UpdateRoute(ctx, route.TenantId, route.RouteId).RouteDataForUpdate(config_client.RouteDataForUpdate{
		TenantId:    route.TenantId,
		RouteConfig: cfg,
	}).Execute()
	if err == nil {
		return nil
	}

	if httpResponse == nil || httpResponse.StatusCode != http.StatusNotFound {
		return proxyapi.NewErrorFromResponse(httpResponse, err)
	}

	_, _, err = tr.tenantRoutesClient.CreateRoute(ctx, route.TenantId).RouteData(config_client.RouteData{
		RouteId:     route.RouteId,
		TenantId:    route.TenantId,
		RouteConfig: cfg,
	}).Execute()
	if err == nil {
		return nil
	}

	return proxyapi.NewErrorFromResponse(httpResponse, err)
}

func (tr *TenantRoutes) UpdateRoute(ctx context.Context, route *types.BasicRouteConfig) error {
	cfg, err := DecodeToMap(route)
	if err != nil {
		return err
	}

	_, httpResponse, err := tr.tenantRoutesClient.UpdateRoute(ctx, route.TenantId, route.RouteId).RouteDataForUpdate(config_client.RouteDataForUpdate{
		TenantId:    route.TenantId,
		RouteConfig: cfg,
	}).Execute()
	if err == nil {
		return nil
	}

	return proxyapi.NewErrorFromResponse(httpResponse, err)
}
