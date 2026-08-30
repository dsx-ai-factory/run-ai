package authorization_test

import (
	"context"
	"testing"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"github.com/run-ai/runai/control-plane/authorization/pkg/client/structs"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/authorization"
	"github.com/run-ai/runai/control-plane/notifications-proxy/internal/config"
	runai_config "github.com/run-ai/runai/runai-common-packages/config"
	"github.com/run-ai/runai/runai-common-packages/runaictx"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
)

const tenantId = "1"

var (
	ctxNoPermissions         = context.WithValue(context.Background(), runaictx.CfTenantId, tenantId)
	ctxWithSystemPermissions = context.WithValue(context.WithValue(ctxNoPermissions, runaictx.CfIsManagementTenant, "true"), structs.PermittedScopesKey, &authorization_api.PermittedScopes{System: ptr.To(true)})
	ctxWithTenantPermissions = context.WithValue(ctxNoPermissions, structs.PermittedScopesKey, &authorization_api.PermittedScopes{Tenants: []string{tenantId}})
)

func TestIsSystem_NoMultiTenancy_Denied(t *testing.T) {
	runai_config.Set(&config.ServiceConfig{
		MultiTenancyEnabled: nil,
	})
	require.False(t, authorization.NewAuthorizer().IsSystem(ctxWithSystemPermissions))
}

func TestIsTenant_NoMultiTenancy_Denied(t *testing.T) {
	runai_config.Set(&config.ServiceConfig{
		MultiTenancyEnabled: nil,
	})
	require.True(t, authorization.NewAuthorizer().IsTenant(ctxWithTenantPermissions))
}

func enableMultiTenancy() {
	runai_config.Set(&config.ServiceConfig{
		MultiTenancyEnabled: ptr.To(true),
	})
}

func disableMultiTenancy() {
	runai_config.Set(&config.ServiceConfig{
		MultiTenancyEnabled: ptr.To(false),
	})
}
