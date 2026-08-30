package clients

import (
	"context"

	proxyapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/generated/config_client"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/types"
)

type UserRoutes struct {
	userRoutesClient *config_client.UserRouteAPIService
}

func NewUserRoutes() *UserRoutes {
	return &UserRoutes{userRoutesClient: GetConfigClient().UserRouteAPI}
}

func (tr *UserRoutes) CreateRoute(ctx context.Context, tenantId, userId string, route *types.UserRouteConfig) error {
	cfg, err := DecodeToMap(route)
	if err != nil {
		return err
	}

	_, httpResponse, err := tr.userRoutesClient.CreateUserRoute(ctx, tenantId, userId).UserRouteData(config_client.UserRouteData{
		TenantId:    tenantId,
		UserId:      userId,
		RouteId:     route.RouteId,
		RouteConfig: cfg,
	}).Execute()
	if err == nil {
		return nil
	}

	return proxyapi.NewErrorFromResponse(httpResponse, err)
}

func (tr *UserRoutes) GetRoutes(ctx context.Context, tenantId, userId string) ([]*types.UserRouteConfig, error) {
	routes, httpResponse, err := tr.userRoutesClient.GetUserRoutes(ctx, tenantId, userId).Execute()
	if err != nil {
		return nil, proxyapi.NewErrorFromResponse(httpResponse, err)
	}

	routesRes := []*types.UserRouteConfig{}
	for _, r := range routes {
		cfg, err := parseRoute(r, tenantId)
		if err != nil {
			return nil, err
		}
		routesRes = append(routesRes, &types.UserRouteConfig{BasicRouteConfig: *cfg, UserId: userId})
	}

	return routesRes, nil
}

func (tr *UserRoutes) DeleteRoute(ctx context.Context, tenantId, userId, routeId string) error {
	_, httpResponse, err := tr.userRoutesClient.DeleteUserRoute(ctx, tenantId, userId, routeId).Execute()
	if err == nil {
		return nil
	}

	return proxyapi.NewErrorFromResponse(httpResponse, err)
}

func (tr *UserRoutes) DeleteAllUserRoutes(ctx context.Context, tenantId string) error {
	_, httpResponse, err := tr.userRoutesClient.DeleteAllUserRoutes(ctx, tenantId).Execute()
	if err == nil {
		return nil
	}

	return proxyapi.NewErrorFromResponse(httpResponse, err)
}

func (tr *UserRoutes) UpdateRoute(ctx context.Context, tenantId, userId string, route *types.UserRouteConfig) error {
	cfg, err := DecodeToMap(route)
	if err != nil {
		return err
	}

	_, httpResponse, err := tr.userRoutesClient.UpdateUserRoute(ctx, tenantId, userId, route.RouteId).UserRouteDataForUpdate(config_client.UserRouteDataForUpdate{
		TenantId:    route.TenantId,
		UserId:      userId,
		RouteConfig: cfg,
	}).Execute()
	if err == nil {
		return nil
	}

	return proxyapi.NewErrorFromResponse(httpResponse, err)
}

func (tr *UserRoutes) DeleteNotifyIdFromAllUserRoutes(ctx context.Context, tenantId string, notifyId string) error {
	_, _, err := tr.userRoutesClient.DeleteNotifyIdFromAllUserRoutes(ctx, tenantId, notifyId).Execute()
	return err
}

func parseRoute(r config_client.RouteData, tenantId string) (*types.BasicRouteConfig, error) {
	cfg := types.BasicRouteConfig{}
	err := DecodeFromMap(r.RouteConfig, &cfg)
	if err != nil {
		return nil, err
	}

	cfg.RouteId = r.RouteId
	cfg.TenantId = tenantId
	return &cfg, nil
}
