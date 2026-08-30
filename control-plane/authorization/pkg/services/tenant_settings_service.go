package services

import (
	"context"

	"github.com/run-ai/runai/control-plane/authorization/pkg/settings"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	setting_pkg "github.com/run-ai/runai/runai-common-packages/setting"
)

const (
	CSPSettingKey                = "tenant.is_csp_managed"
	EnableScopedAuthorizationKey = "tenant.enable_scoped_authorization"
)

// TenantSettingsService defines the interface for tenant settings operations
type TenantSettingsService interface {
	IsDGXCManaged(ctx context.Context) bool
	IsScopedAuthorizationEnabled(ctx context.Context) bool
}

// TenantSettings reads tenant-scoped authorization settings from the local settings package
// (backed by the migrated authorization-service settings table).
type TenantSettings struct{}

// NewTenantSettings creates a new instance of TenantSettings
func NewTenantSettings() TenantSettingsService {
	return &TenantSettings{}
}

// IsDGXCManaged checks if the tenant is DGXC (CSP) managed.
func (ts *TenantSettings) IsDGXCManaged(ctx context.Context) bool {
	tenantId := tenantIdFromContext(ctx)
	if tenantId == "" {
		return false
	}
	s, err := setting_pkg.Get[*settings.TenantIsCspManaged](tenantId)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Warnf("Failed to get TenantIsCspManaged setting, assuming the FF is Off. err: %v", err)
		return false
	}
	return s.Enabled
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
