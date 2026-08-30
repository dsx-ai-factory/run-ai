package common

import (
	"context"
	"fmt"

	authz "github.com/run-ai/runai/control-plane/authorization/pkg/client/generated"

	"github.com/run-ai/runai/runai-mcp-server/internal/runaiclient"
	"github.com/run-ai/runai/runai-mcp-server/internal/runaimcp"
)

func PermittedScopeForAction(ctx context.Context, client *runaiclient.Client, resourceType authz.ResourceType, action authz.Action) (authz.PermittedScopes, error) {
	req := authz.GetPermittedScopesRequest{ResourceType: resourceType}
	resp, httpResp, err := client.Authorization.PermissionsAPI.GetPermittedScopes(ctx).GetPermittedScopesRequest(req).Execute()
	if err != nil {
		return authz.PermittedScopes{}, fmt.Errorf("check %s permissions: %w", resourceType, runaimcp.MapHTTPError(httpResp, err))
	}
	if resp == nil {
		return authz.PermittedScopes{}, fmt.Errorf("check %s permissions: nil response: %w", resourceType, runaimcp.ErrUpstream)
	}
	switch action {
	case authz.ACTION_CREATE:
		return resp.Create, nil
	case authz.ACTION_READ:
		return resp.Read, nil
	case authz.ACTION_UPDATE:
		return resp.Update, nil
	case authz.ACTION_DELETE:
		return resp.Delete, nil
	default:
		return authz.PermittedScopes{}, fmt.Errorf("unsupported permission action %q: %w", action, runaimcp.ErrValidation)
	}
}

func ScopeAllowsTenant(scope authz.PermittedScopes) bool {
	if scope.System != nil && *scope.System {
		return true
	}
	return len(scope.Tenants) > 0
}

func ScopeAllowsCluster(scope authz.PermittedScopes, clusterID string) bool {
	if scope.System != nil && *scope.System {
		return true
	}
	if len(scope.Tenants) > 0 {
		return true
	}
	for _, c := range scope.Clusters {
		if c == clusterID {
			return true
		}
	}
	return false
}

func ScopeAllowsOrgUnit(scope authz.PermittedScopes, orgType, clusterID, orgUnitID, parentDeptID string) bool {
	if ScopeAllowsCluster(scope, clusterID) {
		return true
	}
	if orgType == OrgTypeProject {
		if parentDeptID != "" && contains(scope.Departments, parentDeptID) {
			return true
		}
		return contains(scope.Projects, orgUnitID)
	}
	return contains(scope.Departments, orgUnitID)
}

func contains(ids []string, want string) bool {
	if want == "" {
		return false
	}
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}
