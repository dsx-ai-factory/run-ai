/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package tenant_migration

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/clients"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/internal/notification_state"
	proxyapi "github.com/dsx-ai-factory/run-ai/control-plane/notifications-proxy/pkg/client/generated"
	"github.com/dsx-ai-factory/run-ai/control-plane/notifications-service/pkg/client/types"
)

const (
	EmailEnabled = "email_enabled"
)

var defaultMetadata = map[string]string{
	EmailEnabled: "true",
}

type TenantRoutesClienter interface {
	GetRoute(ctx context.Context, tenantId string) (*types.BasicRouteConfig, error)
	UpdateRoute(ctx context.Context, route *types.BasicRouteConfig) error
}

type TenantMigratorClient struct {
	tenantRoutesClient TenantRoutesClienter
}

func NewClient() *TenantMigratorClient {
	return &TenantMigratorClient{tenantRoutesClient: clients.NewTenantRoutes()}
}

func (tmc *TenantMigratorClient) Migrate(ctx context.Context, tenantId string) error {
	route, err := tmc.tenantRoutesClient.GetRoute(ctx, tenantId)
	if err != nil {
		if errNotFound(err) {
			return nil
		}
		return errors.Join(err, notification_state.ErrClientFailure)
	}
	if shouldNotMigrate(route) {
		return nil
	}

	return tmc.migrate(ctx, tenantId, route)
}

func (tmc *TenantMigratorClient) migrate(ctx context.Context, tenantId string, route *types.BasicRouteConfig) error {
	route.Metadata = defaultMetadata
	if err := tmc.tenantRoutesClient.UpdateRoute(ctx, route); err != nil {
		return errors.Join(fmt.Errorf("failed to migrate notification state for tenant %s: %v", tenantId, err), notification_state.ErrClientFailure)
	}

	return nil
}

func shouldNotMigrate(route *types.BasicRouteConfig) bool {
	return route == nil || route.Metadata != nil
}

func errNotFound(err error) bool {
	apiErr, ok := err.(*proxyapi.Error)
	return ok && apiErr.Code == http.StatusNotFound
}
