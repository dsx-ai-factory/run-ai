package services

import (
	"context"
	"fmt"
	"net/http"

	authorization_utils "github.com/run-ai/runai/control-plane/authorization/pkg/client/helpers"
	"github.com/run-ai/runai/control-plane/cluster-service/internal/generated/pkg/api"
	"github.com/run-ai/runai/runai-common-packages/runaictx"

	authorization_api "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"
	"golang.org/x/exp/slices"
)

var insufficientPermissionsError = api.NewError(http.StatusForbidden, "Insufficient permissions", nil)

func ValidatePermittedScope(ctx context.Context, tenantId int32, clusterUuid string) (*authorization_api.PermittedScopes, *api.Error) {
	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok ||
		authorization_utils.IsPermittedScopesEmpty(permittedScopes) ||
		(!runaictx.IsManagementTenant(ctx) &&
			!slices.Contains(permittedScopes.Tenants, fmt.Sprintf("%d", tenantId)) &&
			!slices.Contains(permittedScopes.Clusters, clusterUuid)) {
		return nil, insufficientPermissionsError
	}

	return permittedScopes, nil
}

func ValidatePermittedTenantScope(ctx context.Context, tenantId int32) (*authorization_api.PermittedScopes, *api.Error) {
	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok ||
		((permittedScopes.System == nil ||
			!*permittedScopes.System) &&
			!slices.Contains(permittedScopes.Tenants, fmt.Sprintf("%d", tenantId))) {
		return nil, insufficientPermissionsError
	}

	return permittedScopes, nil
}

func ValidatePermittedClusterScope(ctx context.Context, clusterUuid string) (*authorization_api.PermittedScopes, *api.Error) {
	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok || authorization_utils.IsPermittedScopesEmpty(permittedScopes) {
		return nil, insufficientPermissionsError
	}

	tenantId, _ := ParseTenantId(ctx)
	noSystem := permittedScopes.System == nil || !*permittedScopes.System
	noTenant := !slices.Contains(permittedScopes.Tenants, fmt.Sprintf("%d", tenantId))
	noCluster := !slices.Contains(permittedScopes.Clusters, clusterUuid)
	if noSystem && noTenant && noCluster {
		return nil, insufficientPermissionsError
	}

	return permittedScopes, nil
}

func GetPermittedClusterScopes(ctx context.Context) ([]string, *api.Error) {
	permittedScopes, ok := authorization_utils.GetPermittedScopesFromCtx(ctx)
	if !ok || authorization_utils.IsPermittedScopesEmpty(permittedScopes) {
		return nil, insufficientPermissionsError
	}

	tenantId, _ := ParseTenantId(ctx)

	if (permittedScopes.System != nil && *permittedScopes.System) ||
		slices.Contains(permittedScopes.Tenants, fmt.Sprintf("%d", tenantId)) {
		return []string{"all"}, nil
	}

	return permittedScopes.Clusters, nil
}
