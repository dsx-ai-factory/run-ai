package tenant_migration

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/clients/tenant_migration/mocks"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/notification_state"
	proxyapi "github.com/run-ai/runai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/notifications-service/pkg/client/types"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const tenantId = "1"

func TestMigrate_Success(t *testing.T) {
	ctx := context.Background()
	trc := mocks.NewTenantRoutesClienter(t)
	client := &TenantMigratorClient{tenantRoutesClient: trc}

	route := &types.BasicRouteConfig{RouteId: "route-1", Metadata: nil}
	trc.On("GetRoute", ctx, tenantId).Return(route, nil)
	trc.On("UpdateRoute", ctx, mock.MatchedBy(func(r *types.BasicRouteConfig) bool {
		return r != nil && r.Metadata != nil && r.Metadata[EmailEnabled] == "true"
	})).Return(nil)

	err := client.Migrate(ctx, tenantId)
	require.NoError(t, err)
}

func TestMigrate_GetRouteNotFound(t *testing.T) {
	ctx := context.Background()
	trc := mocks.NewTenantRoutesClienter(t)
	client := &TenantMigratorClient{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(nil, proxyapi.NewError(http.StatusNotFound, "not found"))

	err := client.Migrate(ctx, tenantId)
	require.NoError(t, err)
}

func TestMigrate_GetRouteError(t *testing.T) {
	ctx := context.Background()
	trc := mocks.NewTenantRoutesClienter(t)
	client := &TenantMigratorClient{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(nil, fmt.Errorf("get route failed"))

	err := client.Migrate(ctx, tenantId)
	require.EqualError(t, err, errors.Join(fmt.Errorf("get route failed"), notification_state.ErrClientFailure).Error())
}

func TestMigrate_AlreadyMigrated_SkipsUpdate(t *testing.T) {
	ctx := context.Background()
	trc := mocks.NewTenantRoutesClienter(t)
	client := &TenantMigratorClient{tenantRoutesClient: trc}

	route := &types.BasicRouteConfig{RouteId: "route-id", Metadata: map[string]string{"existing": "value"}}
	trc.On("GetRoute", ctx, tenantId).Return(route, nil)

	err := client.Migrate(ctx, tenantId)
	require.NoError(t, err)
}

func TestMigrate_NoRoute_SkipsUpdate(t *testing.T) {
	ctx := context.Background()
	trc := mocks.NewTenantRoutesClienter(t)
	client := &TenantMigratorClient{tenantRoutesClient: trc}

	trc.On("GetRoute", ctx, tenantId).Return(nil, nil)

	err := client.Migrate(ctx, tenantId)
	require.NoError(t, err)
}

func TestMigrate_UpdateRouteFailed(t *testing.T) {
	ctx := context.Background()
	trc := mocks.NewTenantRoutesClienter(t)
	client := &TenantMigratorClient{tenantRoutesClient: trc}

	route := &types.BasicRouteConfig{RouteId: "route-1", Metadata: nil}
	trc.On("GetRoute", ctx, tenantId).Return(route, nil)
	trc.On("UpdateRoute", ctx, mock.Anything).Return(fmt.Errorf("update failed"))

	err := client.Migrate(ctx, tenantId)
	require.EqualError(t, err, errors.Join(fmt.Errorf("failed to migrate notification state for tenant 1: update failed"), notification_state.ErrClientFailure).Error())
}
