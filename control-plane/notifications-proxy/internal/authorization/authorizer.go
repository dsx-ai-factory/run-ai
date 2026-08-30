package authorization

import (
	"context"
	"slices"

	authorization_client "github.com/run-ai/runai/control-plane/authorization/pkg/client"
	openapi "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	auth_helpers "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/config"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/logging"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
)

type authorizer struct{}

func NewAuthorizer() *authorizer {
	return &authorizer{}
}

func (a *authorizer) IsSystem(ctx context.Context) bool {
	multiTenancyEnabled := runai_config.Get[*config.ServiceConfig]().MultiTenancyEnabled
	if multiTenancyEnabled == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("multi tenancy enabled is not set")
		return false
	}
	if !*multiTenancyEnabled {
		return a.IsTenant(ctx)
	}
	return auth_helpers.HasSystemPermissions(ctx)
}

func (a *authorizer) IsTenant(ctx context.Context) bool {
	tenantId := getTenantId(ctx)
	if tenantId == "" {
		return false
	}

	permittedScopes := getPermittedScopes(ctx)
	if permittedScopes == nil {
		return false
	}
	logging.WithCtx(runaictx.LoggerHook(ctx)).Debugf("checking tenant scope permissions for tenant %s with tenants permission %v", tenantId, permittedScopes.Tenants)
	return slices.Contains(permittedScopes.GetTenants(), tenantId)
}

func getPermittedScopes(ctx context.Context) *openapi.PermittedScopes {
	permittedScopes, _ := authorization_client.ParsePermittedScopes(ctx)
	if permittedScopes == nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("permitted scopes are not set")
		return nil
	}
	return permittedScopes
}

func getTenantId(ctx context.Context) string {
	tenantId, err := runaictx.ParseTenantId(ctx)
	if err != nil {
		logging.WithCtx(runaictx.LoggerHook(ctx)).Errorf("failed to parse tenant id: %v", err)
		return ""
	}
	return tenantId
}
