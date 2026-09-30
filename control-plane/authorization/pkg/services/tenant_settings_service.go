/*
 * SPDX-FileCopyrightText: Copyright (c) 2026 NVIDIA CORPORATION & AFFILIATES. All rights reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package services

import (
	"context"

	"github.com/dsx-ai-factory/run-ai/control-plane/authorization/pkg/settings"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/logging"
	"github.com/dsx-ai-factory/run-ai/runai-common-packages/runaictx"
	setting_pkg "github.com/dsx-ai-factory/run-ai/runai-common-packages/setting"
)

const (
	EnableScopedAuthorizationKey = "tenant.enable_scoped_authorization"
)

// TenantSettingsService defines the interface for tenant settings operations
type TenantSettingsService interface {
	IsScopedAuthorizationEnabled(ctx context.Context) bool
}

// TenantSettings reads tenant-scoped authorization settings from the local settings package
// (backed by the migrated authorization-service settings table).
type TenantSettings struct{}

// NewTenantSettings creates a new instance of TenantSettings
func NewTenantSettings() TenantSettingsService {
	return &TenantSettings{}
}

// IsScopedAuthorizationEnabled checks if scoped authorization is enabled for the tenant.
func (ts *TenantSettings) IsScopedAuthorizationEnabled(ctx context.Context) bool {
	tenantId := tenantIdFromContext(ctx)
	if tenantId == "" {
		return false
	}
	s, err := setting_pkg.Get[*settings.EnableScopedAuthorization](tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get EnableScopedAuthorization setting, assuming the FF is Off. err: %v", err)
		return false
	}
	return s.Enabled
}

func tenantIdFromContext(ctx context.Context) string {
	t := runaictx.GetString(ctx, runaictx.CfTenantId)
	if t == nil {
		return ""
	}
	return *t
}
